package workers

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

// StartReorderWorker starts a background worker that checks for products at or below
// reorder level and pushes notifications to admin users, and optionally creates draft POs.
func StartReorderWorker(db *gorm.DB, notifSvc *services.NotificationService) {
	go func() {
		log.Println("📦 Starting Reorder Alert worker...")
		// Run immediately on startup, then every 6 hours
		processReorderAlerts(db, notifSvc)
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			processReorderAlerts(db, notifSvc)
		}
	}()
}

func processReorderAlerts(db *gorm.DB, notifSvc *services.NotificationService) {
	var alerts []TenantAlert
	if err := db.Table("products").
		Select("tenant_id, id as product_id, name, sku, current_stock as stock, reorder_level as reorder, branch_id").
		Where("track_inventory = ? AND current_stock <= reorder_level AND current_stock >= 0 AND is_active = ?", true, true).
		Where("reorder_level > 0").
		Scan(&alerts).Error; err != nil {
		log.Printf("[ReorderWorker] Error querying low-stock products: %v", err)
		return
	}

	if len(alerts) == 0 {
		return
	}

	// Group by tenant
	tenantAlerts := make(map[string][]TenantAlert)
	for _, a := range alerts {
		tenantAlerts[a.TenantID] = append(tenantAlerts[a.TenantID], a)
	}

	for tenantIDStr, products := range tenantAlerts {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			continue
		}

		// Anti-spam: Check if we already sent a reorder notification in the last 12 hours for this tenant
		var recentCount int64
		db.Model(&models.Notification{}).
			Where("tenant_id = ? AND category = ? AND created_at > ?", tenantID, "inventory", time.Now().Add(-12*time.Hour)).
			Where("title LIKE ?", "Reorder Alert%").
			Count(&recentCount)
		if recentCount > 0 {
			continue
		}

		// Build notification message
		title := fmt.Sprintf("Reorder Alert: %d product(s) need restocking", len(products))
		var msgBody string
		limit := 3
		for i, p := range products {
			if i >= limit {
				msgBody += fmt.Sprintf("...and %d more.", len(products)-limit)
				break
			}
			msgBody += fmt.Sprintf("• %s (SKU: %s) — %.0f left (reorder at %.0f)\n", p.Name, p.SKU, p.Stock, p.Reorder)
		}

		notifSvc.CreateAndPushToAdmins(
			tenantID,
			title,
			msgBody,
			"inventory",
			"/inventory?tab=alerts",
			"warning",
		)

		// Auto-create draft PurchaseOrders for products with a supplier set
		autoDraftPOs(db, tenantID, products)

		log.Printf("[ReorderWorker] Sent reorder alert for tenant %s (%d products)", tenantIDStr, len(products))
	}
}

type TenantAlert struct {
	TenantID  string
	ProductID string
	Name      string
	SKU       string
	Stock     float64
	Reorder   float64
	BranchID  *string
}

// autoDraftPOs groups low-stock products by supplier and creates draft POs.
func autoDraftPOs(db *gorm.DB, tenantID uuid.UUID, alerts []TenantAlert) {
	// Map supplierID → []products
	type ProductWithSupplier struct {
		ProductID  string
		SupplierID *string
		BranchID   *string
		Stock      float64
		Reorder    float64
		CostPrice  float64
	}

	var productsWithSuppliers []ProductWithSupplier
	productIDs := make([]string, len(alerts))
	for i, a := range alerts {
		productIDs[i] = a.ProductID
	}

	db.Table("products").
		Select("id as product_id, supplier_id, branch_id, current_stock as stock, reorder_level as reorder, cost_price").
		Where("id IN ? AND supplier_id IS NOT NULL", productIDs).
		Scan(&productsWithSuppliers)

	// Group by supplier+branch
	type POKey struct {
		SupplierID string
		BranchID   string
	}
	poGroups := make(map[POKey][]ProductWithSupplier)
	for _, p := range productsWithSuppliers {
		if p.SupplierID == nil {
			continue
		}
		branchStr := ""
		if p.BranchID != nil {
			branchStr = *p.BranchID
		}
		key := POKey{SupplierID: *p.SupplierID, BranchID: branchStr}
		poGroups[key] = append(poGroups[key], p)
	}

	for key, items := range poGroups {
		supplierID, err := uuid.Parse(key.SupplierID)
		if err != nil {
			continue
		}

		// Check if there's already a draft/open PO for this supplier in the last 7 days
		var existingCount int64
		db.Model(&models.PurchaseOrder{}).
			Where("supplier_id = ? AND status = ? AND created_at > ?", supplierID, "draft", time.Now().AddDate(0, 0, -7)).
			Count(&existingCount)
		if existingCount > 0 {
			continue
		}

		// Build the PO
		var branchID *uuid.UUID
		if key.BranchID != "" {
			bid, err := uuid.Parse(key.BranchID)
			if err == nil {
				branchID = &bid
			}
		}

		po := models.PurchaseOrder{
			BranchScoped: models.BranchScoped{
				BranchID: branchID,
			},
			PONumber:   fmt.Sprintf("AUTO-PO-%d", time.Now().UnixNano()),
			SupplierID: supplierID,
			Status:     "draft",
		}

		var poItems []models.PurchaseOrderItem
		var totalAmount float64
		for _, item := range items {
			productID, err := uuid.Parse(item.ProductID)
			if err != nil {
				continue
			}
			// Suggest: order enough to reach 2× reorder level
			suggestedQty := (item.Reorder * 2) - item.Stock
			if suggestedQty <= 0 {
				suggestedQty = item.Reorder
			}
			poItems = append(poItems, models.PurchaseOrderItem{
				ProductID:       productID,
				QuantityOrdered: suggestedQty,
				UnitCost:        item.CostPrice,
			})
			totalAmount += suggestedQty * item.CostPrice
		}

		po.TotalAmount = totalAmount
		po.Items = poItems

		if err := db.Create(&po).Error; err != nil {
			log.Printf("[ReorderWorker] Failed to create draft PO for supplier %s: %v", key.SupplierID, err)
		} else {
			log.Printf("[ReorderWorker] Created draft PO %s for supplier %s", po.PONumber, key.SupplierID)
		}
	}
}
