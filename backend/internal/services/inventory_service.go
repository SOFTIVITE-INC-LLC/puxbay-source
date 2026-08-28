package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type InventoryService struct {
	db       *gorm.DB
	tenantID uuid.UUID
}

func NewInventoryService(db *gorm.DB, tenantID uuid.UUID) *InventoryService {
	return &InventoryService{db: db, tenantID: tenantID}
}

// ---------------------------------------------------------
// Transfers
// ---------------------------------------------------------

func (s *InventoryService) ListTransfers(tenantID uuid.UUID, limit, offset int, branchID string) ([]models.StockTransfer, int64, error) {
	var transfers []models.StockTransfer
	var total int64

	query := s.db.Model(&models.StockTransfer{})
	if branchID != "" {
		query = query.Where("from_branch_id = ? OR to_branch_id = ?", branchID, branchID)
	}
	query.Count(&total)

	if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&transfers).Error; err != nil {
		return nil, 0, err
	}
	return transfers, total, nil
}

type TransferCreateInput struct {
	ReferenceNo  string
	FromBranchID uuid.UUID
	ToBranchID   uuid.UUID
	Notes        string
	CreatedByID  uuid.UUID
	Items        []TransferItemInput
}

type TransferItemInput struct {
	ProductID uuid.UUID
	Quantity  float64
}

func (s *InventoryService) CreateTransfer(input TransferCreateInput) (*models.StockTransfer, error) {
	ref := input.ReferenceNo
	if ref == "" || ref == "TRF-0001" || ref == "STK-0001" {
		ref = "TRF-" + uuid.New().String()[:8]
	} else {
		// Just to be safe from duplicates
		ref = ref + "-" + uuid.New().String()[:4]
	}

	transfer := models.StockTransfer{
		ReferenceNo:  ref,
		FromBranchID: input.FromBranchID,
		ToBranchID:   input.ToBranchID,
		Status:       "requested",
		Notes:        &input.Notes,
		CreatedByID:  &input.CreatedByID,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&transfer).Error; err != nil {
			return err
		}
		for _, item := range input.Items {
			tItem := models.StockTransferItem{
				TransferID: transfer.ID,
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
			}
			if err := tx.Create(&tItem).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

// ---------------------------------------------------------
// Purchase Orders
// ---------------------------------------------------------

func (s *InventoryService) ListPOs(tenantID uuid.UUID, limit, offset int, branchID string) ([]models.PurchaseOrder, int64, error) {
	var pos []models.PurchaseOrder
	var total int64

	query := s.db.Model(&models.PurchaseOrder{})
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	query.Count(&total)

	if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	return pos, total, nil
}

type POCreateInput struct {
	PONumber   string
	SupplierID uuid.UUID
	BranchID   uuid.UUID
	Notes      string
	Items      []POItemInput
}

type POItemInput struct {
	ProductID       uuid.UUID
	QuantityOrdered float64
	UnitCost        float64
}

func (s *InventoryService) CreatePO(input POCreateInput) (*models.PurchaseOrder, error) {
	po := models.PurchaseOrder{
		BranchScoped: models.BranchScoped{
			BranchID: &input.BranchID,
		},
		PONumber:   input.PONumber,
		SupplierID: input.SupplierID,
		Status:     "draft",
		Notes:      &input.Notes,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&po).Error; err != nil {
			return err
		}

		var totalAmount float64
		for _, item := range input.Items {
			poItem := models.PurchaseOrderItem{
				POID:            po.ID,
				ProductID:       item.ProductID,
				QuantityOrdered: item.QuantityOrdered,
				UnitCost:        item.UnitCost,
			}
			if err := tx.Create(&poItem).Error; err != nil {
				return err
			}
			totalAmount += item.QuantityOrdered * item.UnitCost
		}

		po.TotalAmount = totalAmount
		if err := tx.Save(&po).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &po, nil
}

// ---------------------------------------------------------
// Stocktakes
// ---------------------------------------------------------

func (s *InventoryService) ListStocktakes(tenantID uuid.UUID, limit, offset int, branchID string) ([]models.StocktakeSession, int64, error) {
	var stocktakes []models.StocktakeSession
	var total int64

	query := s.db.Model(&models.StocktakeSession{})
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	query.Count(&total)

	if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&stocktakes).Error; err != nil {
		return nil, 0, err
	}
	return stocktakes, total, nil
}

func (s *InventoryService) GetTransfer(id string) (*models.StockTransfer, error) {
	var transfer models.StockTransfer
	if err := s.db.Where("id = ?", id).Preload("Items").Preload("Items.Product").First(&transfer).Error; err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (s *InventoryService) UpdateTransferStatus(id string, status string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var transfer models.StockTransfer
		if err := tx.Where("id = ?", id).Preload("Items").First(&transfer).Error; err != nil {
			return err
		}

		if transfer.Status == status {
			return nil
		}

		// When shipped, deduct from source branch
		if status == "shipped" && transfer.Status != "shipped" && transfer.Status != "completed" {
			for _, item := range transfer.Items {
				var product models.Product
				if err := tx.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
					return err
				}
				newStock := product.CurrentStock - item.Quantity
				if err := tx.Model(&product).Update("current_stock", newStock).Error; err != nil {
					return err
				}
				movement := models.StockMovement{
					ProductID:     product.ID,
					BranchID:      &transfer.FromBranchID,
					Quantity:      -item.Quantity,
					PreviousStock: product.CurrentStock,
					NewStock:      newStock,
					Reason:        "transfer_out",
					ReferenceID:   &transfer.ReferenceNo,
				}
				if err := tx.Create(&movement).Error; err != nil {
					return err
				}
			}
		}

		// When received, add to destination branch
		if status == "received" && transfer.Status != "received" {
			for _, item := range transfer.Items {
				var product models.Product
				if err := tx.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
					return err
				}
				// Find if destination branch already has this product by SKU
				var destProduct models.Product
				var newStock float64
				err := tx.Where("sku = ? AND branch_id = ?", product.SKU, transfer.ToBranchID).First(&destProduct).Error
				if err != nil {
					// Duplicate product for destination branch
					newProduct := product
					newProduct.ID = uuid.New()
					newProduct.BranchID = &transfer.ToBranchID
					newProduct.CurrentStock = item.Quantity
					if err := tx.Create(&newProduct).Error; err != nil {
						return err
					}
					product = newProduct // Use this for the movement record
					newStock = item.Quantity
				} else {
					// Add to existing product in destination branch
					newStock = destProduct.CurrentStock + item.Quantity
					if err := tx.Model(&destProduct).Updates(map[string]interface{}{
						"current_stock": newStock,
					}).Error; err != nil {
						return err
					}
					product = destProduct // Use this for the movement record
				}

				movement := models.StockMovement{
					ProductID:     product.ID,
					BranchID:      &transfer.ToBranchID,
					Quantity:      item.Quantity,
					PreviousStock: product.CurrentStock,
					NewStock:      newStock,
					Reason:        "transfer_in",
					ReferenceID:   &transfer.ReferenceNo,
				}
				if err := tx.Create(&movement).Error; err != nil {
					return err
				}
			}
		}

		transfer.Status = status
		return tx.Save(&transfer).Error
	})
}

func (s *InventoryService) GetPO(id string) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := s.db.Where("id = ?", id).Preload("Items").Preload("Items.Product").First(&po).Error; err != nil {
		return nil, err
	}
	return &po, nil
}

type POReceiveInput struct {
	Items []POReceiveItemInput
}

type POReceiveItemInput struct {
	ItemID           uuid.UUID
	QuantityReceived float64
}

func (s *InventoryService) ReceivePO(id string, input POReceiveInput) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var po models.PurchaseOrder
		if err := tx.Where("id = ?", id).Preload("Items").First(&po).Error; err != nil {
			return err
		}

		allReceived := true
		receiveMap := make(map[uuid.UUID]float64)
		for _, item := range input.Items {
			receiveMap[item.ItemID] = item.QuantityReceived
		}

		for i, item := range po.Items {
			qtyToReceive := receiveMap[item.ID]
			if qtyToReceive > 0 {
				po.Items[i].QuantityReceived += qtyToReceive
				if err := tx.Save(&po.Items[i]).Error; err != nil {
					return err
				}

				var product models.Product
				if err := tx.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
					return err
				}

				var branchID uuid.UUID
				if po.BranchID != nil {
					branchID = *po.BranchID
				} else {
					branchID = s.tenantID // Fallback
				}

				// Verify product exists in the destination branch
				var destProduct models.Product
				var newStock float64
				err := tx.Where("sku = ? AND branch_id = ?", product.SKU, branchID).First(&destProduct).Error
				if err != nil {
					// Duplicate product for destination branch
					destProduct = product
					destProduct.ID = uuid.New()
					destProduct.BranchID = &branchID
					destProduct.CurrentStock = qtyToReceive
					if err := tx.Create(&destProduct).Error; err != nil {
						return err
					}
					newStock = qtyToReceive
				} else {
					newStock = destProduct.CurrentStock + qtyToReceive
					if err := tx.Model(&destProduct).Update("current_stock", newStock).Error; err != nil {
						return err
					}
				}

				movement := models.StockMovement{
					ProductID:     destProduct.ID,
					BranchID:      &branchID,
					Quantity:      qtyToReceive,
					PreviousStock: destProduct.CurrentStock,
					NewStock:      newStock,
					Reason:        "po_receipt",
					ReferenceID:   &po.PONumber,
				}
				if err := tx.Create(&movement).Error; err != nil {
					return err
				}
			}

			if po.Items[i].QuantityReceived < item.QuantityOrdered {
				allReceived = false
			}
		}

		if allReceived {
			po.Status = "received"
		} else {
			po.Status = "partially_received"
		}

		if err := tx.Save(&po).Error; err != nil {
			return err
		}

		var supplier models.Supplier
		if err := tx.Where("id = ?", po.SupplierID).First(&supplier).Error; err != nil {
			return err
		}

		if allReceived {
			newBalance := supplier.CreditBalance + po.TotalAmount
			if err := tx.Model(&supplier).Update("credit_balance", newBalance).Error; err != nil {
				return err
			}

			// Automatically create a ledger entry for the invoice
			poRef := po.PONumber
			notes := "Auto-generated invoice from PO receipt"
			ledgerEntry := models.SupplierLedgerEntry{
				SupplierID:  supplier.ID,
				EntryType:   "invoice",
				Amount:      po.TotalAmount,
				Balance:     newBalance,
				ReferenceID: &poRef,
				Notes:       &notes,
			}
			if err := tx.Create(&ledgerEntry).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

type StocktakeInput struct {
	Name     string
	BranchID uuid.UUID
}

func (s *InventoryService) CreateStocktake(input StocktakeInput) (*models.StocktakeSession, error) {
	token := uuid.New()
	st := models.StocktakeSession{
		Name:         input.Name,
		BranchScoped: models.BranchScoped{BranchID: &input.BranchID},
		AccessToken:  &token,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&st).Error; err != nil {
			return err
		}

		var products []models.Product
		if err := tx.Where("branch_id = ? AND track_inventory = ?", input.BranchID, true).Find(&products).Error; err != nil {
			return err
		}

		if len(products) > 0 {
			entries := make([]models.StocktakeEntry, 0, len(products))
			for _, p := range products {
				entries = append(entries, models.StocktakeEntry{
					SessionID:     st.ID,
					ProductID:     p.ID,
					ExpectedStock: p.CurrentStock,
					ActualStock:   0,
					Difference:    -p.CurrentStock,
				})
			}
			if err := tx.Create(&entries).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &st, nil
}

func (s *InventoryService) GetStocktake(id string) (*models.StocktakeSession, error) {
	var st models.StocktakeSession
	if err := s.db.Where("id = ?", id).Preload("Entries").Preload("Entries.Product").First(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *InventoryService) FinalizeStocktake(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var session models.StocktakeSession
		if err := tx.Where("id = ?", id).Preload("Entries").Preload("Entries.Product").First(&session).Error; err != nil {
			return err
		}

		if session.Status == "completed" {
			return nil
		}

		for _, entry := range session.Entries {
			if entry.Difference != 0 {
				newStock := entry.Product.CurrentStock + entry.Difference
				if err := tx.Model(entry.Product).Update("current_stock", newStock).Error; err != nil {
					return err
				}

				movement := models.StockMovement{
					BranchID:      session.BranchID,
					ProductID:     entry.ProductID,
					Quantity:      entry.Difference,
					PreviousStock: entry.Product.CurrentStock,
					NewStock:      newStock,
					Reason:        "adjustment",
				}
				if err := tx.Create(&movement).Error; err != nil {
					return err
				}
			}
		}

		now := time.Now()
		session.CompletedAt = &now
		session.Status = "completed"
		return tx.Save(&session).Error
	})
}

func (s *InventoryService) ScanStocktakeProduct(token string, query string) ([]map[string]interface{}, error) {
	var session models.StocktakeSession
	if err := s.db.Where("access_token = ?", token).First(&session).Error; err != nil {
		return nil, err
	}

	var products []models.Product
	q := s.db.Where("branch_id = ? OR branch_id IS NULL", session.BranchID).Where("is_active = ?", true)
	if query != "" {
		q = q.Where("barcode = ? OR sku ILIKE ? OR name ILIKE ?", query, query, "%"+query+"%")
	}
	if err := q.Limit(20).Find(&products).Error; err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for _, p := range products {
		var entry models.StocktakeEntry
		countedQuantity := 0.0
		if err := s.db.Where("session_id = ? AND product_id = ?", session.ID, p.ID).First(&entry).Error; err == nil {
			countedQuantity = entry.ActualStock
		}

		results = append(results, map[string]interface{}{
			"id":            p.ID,
			"name":          p.Name,
			"sku":           p.SKU,
			"barcode":       p.Barcode,
			"current_count": countedQuantity,
		})
	}
	return results, nil
}

func (s *InventoryService) UpdateStocktakeCount(token string, productID uuid.UUID, quantity float64, mode string) (float64, error) {
	var newCount float64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var session models.StocktakeSession
		if err := tx.Where("access_token = ?", token).First(&session).Error; err != nil {
			return err
		}
		if session.Status == "completed" {
			return fmt.Errorf("session closed")
		}

		var product models.Product
		if err := tx.Where("id = ?", productID).First(&product).Error; err != nil {
			return err
		}

		var entry models.StocktakeEntry
		err := tx.Where("session_id = ? AND product_id = ?", session.ID, productID).First(&entry).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				entry = models.StocktakeEntry{
					SessionID:     session.ID,
					ProductID:     productID,
					ExpectedStock: product.CurrentStock,
					ActualStock:   0,
					Difference:    -product.CurrentStock,
				}
				if err := tx.Create(&entry).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		if mode == "add" {
			entry.ActualStock += quantity
		} else {
			entry.ActualStock = quantity
		}
		entry.Difference = entry.ActualStock - entry.ExpectedStock
		newCount = entry.ActualStock
		return tx.Save(&entry).Error
	})
	return newCount, err
}

func (s *InventoryService) ListMovements(tenantID uuid.UUID, limit, offset int, branchID string) ([]models.StockMovement, int64, error) {
	var movements []models.StockMovement
	var total int64

	query := s.db.Model(&models.StockMovement{})
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	query.Count(&total)

	if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&movements).Error; err != nil {
		return nil, 0, err
	}
	return movements, total, nil
}

func (s *InventoryService) ReceiveStock(productID string, quantity float64, reason string, userID *uuid.UUID, tenantID uuid.UUID, reqBranchID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var product models.Product
		if err := tx.Where("id = ?", productID).First(&product).Error; err != nil {
			return err
		}

		previousStock := product.CurrentStock
		newStock := previousStock + quantity

		if err := tx.Model(&models.Product{}).Where("id = ?", productID).Update("current_stock", newStock).Error; err != nil {
			return err
		}

		var branchID uuid.UUID
		if reqBranchID != uuid.Nil {
			branchID = reqBranchID
		} else if product.BranchID != nil {
			branchID = *product.BranchID
		} else {
			// If all fails, we query a branch for this tenant
			var b models.Branch
			tx.Where("tenant_id = ?", tenantID).First(&b)
			branchID = b.ID
		}

		movement := models.StockMovement{
			BranchID:      &branchID,
			ProductID:     product.ID,
			Quantity:      quantity,
			PreviousStock: previousStock,
			NewStock:      newStock,
			Reason:        reason,
			UserID:        userID,
		}

		return tx.Create(&movement).Error
	})
}

func (s *InventoryService) GetProductHistory(productID string) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	if err := s.db.Where("product_id = ?", productID).Order("created_at desc").Find(&movements).Error; err != nil {
		return nil, err
	}
	return movements, nil
}
func (s *InventoryService) LowStockAlerts() ([]models.Product, error) {
	var alerts []models.Product
	// Using a simple query for products that track inventory and have current stock below their reorder level
	if err := s.db.Where("track_inventory = ? AND current_stock <= reorder_level", true).Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// ---------------------------------------------------------
// Batch & Expiry Tracking
// ---------------------------------------------------------

type BatchCreateInput struct {
	ProductID       uuid.UUID
	BranchID        uuid.UUID
	BatchNumber     string
	Quantity        float64
	ExpiryDate      *time.Time
	ManufactureDate *time.Time
}

func (s *InventoryService) ListBatches(productID string) ([]models.StockBatch, error) {
	var batches []models.StockBatch
	if err := s.db.Where("product_id = ?", productID).
		Preload("Product").
		Order("expiry_date asc").
		Find(&batches).Error; err != nil {
		return nil, err
	}
	return batches, nil
}

func (s *InventoryService) CreateBatch(input BatchCreateInput) (*models.StockBatch, error) {
	batch := models.StockBatch{
		ProductID:       input.ProductID,
		BranchID:        input.BranchID,
		BatchNumber:     input.BatchNumber,
		Quantity:        input.Quantity,
		ExpiryDate:      input.ExpiryDate,
		ManufactureDate: input.ManufactureDate,
	}
	if err := s.db.Create(&batch).Error; err != nil {
		return nil, err
	}
	s.db.Preload("Product").First(&batch, batch.ID)
	return &batch, nil
}

func (s *InventoryService) UpdateBatch(id string, input BatchCreateInput) (*models.StockBatch, error) {
	var batch models.StockBatch
	if err := s.db.First(&batch, "id = ?", id).Error; err != nil {
		return nil, err
	}
	batch.BatchNumber = input.BatchNumber
	batch.Quantity = input.Quantity
	batch.ExpiryDate = input.ExpiryDate
	batch.ManufactureDate = input.ManufactureDate
	if err := s.db.Save(&batch).Error; err != nil {
		return nil, err
	}
	s.db.Preload("Product").First(&batch, batch.ID)
	return &batch, nil
}

func (s *InventoryService) DeleteBatch(id string) error {
	return s.db.Delete(&models.StockBatch{}, "id = ?", id).Error
}

func (s *InventoryService) ListExpiringBatches(branchID string, days int) ([]models.StockBatch, error) {
	var batches []models.StockBatch
	cutoff := time.Now().AddDate(0, 0, days)
	query := s.db.Where("expiry_date IS NOT NULL AND expiry_date <= ?", cutoff).
		Preload("Product")
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	if err := query.Order("expiry_date asc").Find(&batches).Error; err != nil {
		return nil, err
	}
	return batches, nil
}
