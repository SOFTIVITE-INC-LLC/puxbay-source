package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/dto"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderService struct {
	db         *gorm.DB
	smsService *SMSService
	tenantID   uuid.UUID
}

func NewOrderService(db *gorm.DB, sms *SMSService, tenantID uuid.UUID) *OrderService {
	return &OrderService{db: db, smsService: sms, tenantID: tenantID}
}

type OrderListParams struct {
	BranchID   string
	Status     string
	OrderType  string
	ProductID  string
	CustomerID string
	CashierID  string
	Search     string
	Limit      int
	Offset     int
}

func (s *OrderService) ListOrders(params OrderListParams) ([]dto.OrderListResponse, int64, error) {
	query := s.db.Model(&models.Order{})

	if params.BranchID != "" {
		query = query.Where("(orders.branch_id = ? OR orders.branch_id IS NULL OR orders.order_type IN ('online', 'storefront', 'pickup', 'delivery'))", params.BranchID)
	}
	if params.Status != "" {
		query = query.Where("orders.status = ?", params.Status)
	}
	if params.OrderType != "" {
		if params.OrderType == "online" || params.OrderType == "storefront" {
			query = query.Where("orders.order_type IN ('online', 'storefront', 'pickup', 'delivery')")
		} else if params.OrderType == "pos" || params.OrderType == "in_store" {
			query = query.Where("orders.order_type IN ('pos', 'in_store')")
		} else {
			query = query.Where("orders.order_type = ?", params.OrderType)
		}
	}
	if params.ProductID != "" {
		query = query.Where("EXISTS (SELECT 1 FROM order_items WHERE order_items.order_id = orders.id AND order_items.product_id = ? AND order_items.deleted_at IS NULL)", params.ProductID)
	}
	if params.CustomerID != "" {
		query = query.Where("orders.customer_id = ?", params.CustomerID)
	}
	if params.CashierID != "" {
		query = query.Where("orders.cashier_id = ?", params.CashierID)
	}
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("orders.id::text ILIKE ? OR orders.order_number ILIKE ?", search, search)
	}

	var total int64
	query.Count(&total)

	// Fetch base orders with Customer preloaded (single query with IN clause, not N+1)
	var orders []models.Order
	if err := query.Preload("Customer").Order("orders.created_at desc").Offset(params.Offset).Limit(params.Limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	// Fetch item counts in bulk (one query)
	var orderIDs []uuid.UUID
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
	}

	itemCounts := make(map[uuid.UUID]int)
	if len(orderIDs) > 0 {
		type countResult struct {
			OrderID uuid.UUID
			Count   int
		}
		var counts []countResult
		s.db.Table("order_items").
			Select("order_id, COUNT(*) as count").
			Where("order_id IN ? AND deleted_at IS NULL", orderIDs).
			Group("order_id").
			Scan(&counts)

		for _, c := range counts {
			itemCounts[c.OrderID] = c.Count
		}
	}

	// Map to DTO
	var responses []dto.OrderListResponse
	for _, o := range orders {
		resp := dto.OrderListResponse{
			ID:            o.ID,
			CreatedAt:     o.CreatedAt,
			UpdatedAt:     o.UpdatedAt,
			OrderNumber:   o.OrderNumber,
			Total:         o.Total,
			AmountPaid:    o.AmountPaid,
			Status:        o.Status,
			PaymentStatus: o.PaymentStatus,
			PaymentMethod: o.PaymentMethod,
			OrderType:     o.OrderType,
			ReceiptToken:  o.ReceiptToken,
			CustomerID:    o.CustomerID,
			CashierID:     o.CashierID,
			ItemCount:     itemCounts[o.ID],
			Customer:      o.Customer,
		}
		if o.Customer != nil {
			resp.CustomerName = o.Customer.Name
		}
		responses = append(responses, resp)
	}

	return responses, total, nil
}

type OrderSummaryStats struct {
	TotalOrders      int64 `json:"total_orders"`
	FailedOrders     int64 `json:"failed_orders"`
	PosOrders        int64 `json:"pos_orders"`
	StorefrontOrders int64 `json:"storefront_orders"`
	AppOrders        int64 `json:"app_orders"`
}

func (s *OrderService) GetOrderSummaryStats(branchID string, cashierID string) (*OrderSummaryStats, error) {
	var stats OrderSummaryStats
	query := s.db.Model(&models.Order{})
	
	if branchID != "" {
		query = query.Where("(branch_id = ? OR branch_id IS NULL OR order_type IN ('online', 'storefront', 'pickup', 'delivery'))", branchID)
	}
	if cashierID != "" {
		query = query.Where("cashier_id = ?", cashierID)
	}

	// Calculate counts using conditional aggregation
	err := query.Select(`
		COUNT(id) as total_orders,
		COUNT(id) FILTER (WHERE status IN ('voided', 'cancelled')) as failed_orders,
		COUNT(id) FILTER (WHERE order_type IN ('in_store', 'pos')) as pos_orders,
		COUNT(id) FILTER (WHERE order_type IN ('online', 'storefront', 'pickup', 'delivery')) as storefront_orders,
		COUNT(id) FILTER (WHERE order_type = 'kiosk') as app_orders
	`).Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (s *OrderService) GetOrder(id string) (*models.Order, error) {
	var order models.Order
	if err := s.db.Where("id = ?", id).Preload("Items").Preload("Items.Product").Preload("Customer").First(&order).Error; err != nil {
		return nil, errors.New("order not found")
	}
	return &order, nil
}

type OrderItemInput struct {
	ProductID uuid.UUID
	VariantID *uuid.UUID
	Quantity  float64
	UnitPrice float64
	Discount  float64
	Total     float64
}

type OrderPaymentInput struct {
	Method string
	Amount float64
}

type OrderCreateInput struct {
	BranchID     *uuid.UUID
	CustomerID   *uuid.UUID
	CashierID    *uuid.UUID
	Subtotal     float64
	Tax          float64
	Discount     float64
	Total        float64
	AmountPaid   float64
	RedeemPoints float64
	Payments     []OrderPaymentInput
	OrderType    string
	Notes        string
	Items        []OrderItemInput
}

// CreateOrder creates an order and deducts inventory for tracked products (Gap #17).
func (s *OrderService) CreateOrder(input OrderCreateInput) (*models.Order, error) {
	status := "completed"
	if input.OrderType == "kiosk" {
		status = "pending"
	}

	order := models.Order{
		OrderNumber:   generateOrderNumber(),
		CustomerID:    input.CustomerID,
		CashierID:     input.CashierID,
		Subtotal:      input.Subtotal,
		Tax:           input.Tax,
		Discount:      input.Discount,
		Total:         input.Total,
		AmountPaid:    input.AmountPaid,
		PaymentMethod: determinePrimaryPaymentMethod(input.Payments),
		OrderType:     input.OrderType,
		Status:        status,
		PaymentStatus: "paid",
		ReceiptToken:  generateReceiptToken(),
	}

	if input.Notes != "" {
		order.Notes = &input.Notes
	}

	if input.BranchID != nil {
		order.BranchID = input.BranchID
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var movements []models.StockMovement

		for _, item := range input.Items {
			var product models.Product
			if err := tx.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
				return fmt.Errorf("invalid product: %s", item.ProductID)
			}

			orderItem := models.OrderItem{
				ProductID: item.ProductID,
				VariantID: item.VariantID,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Discount:  item.Discount,
				Total:     item.Total,
				CostPrice: product.CostPrice,
			}
			order.Items = append(order.Items, orderItem)

			// Gap #17: Deduct inventory for tracked products
			// Gap #11: Prevent negative stock with WHERE guard
			if product.TrackInventory {
				result := tx.Model(&models.Product{}).
					Where("id = ? AND current_stock >= ?", product.ID, item.Quantity).
					Update("current_stock", gorm.Expr("current_stock - ?", item.Quantity))
				if result.RowsAffected == 0 {
					return fmt.Errorf("insufficient stock for product %s (available: %.2f, requested: %.2f)",
						product.Name, product.CurrentStock, item.Quantity)
				}

				// Queue stock movement for history
				branchID := uuid.Nil
				if order.BranchID != nil {
					branchID = *order.BranchID
				}
				movements = append(movements, models.StockMovement{
					TenantID:      s.tenantID,
					BranchID:      branchID,
					ProductID:     product.ID,
					VariantID:     item.VariantID,
					Quantity:      -item.Quantity,
					PreviousStock: product.CurrentStock,
					NewStock:      product.CurrentStock - item.Quantity,
					Reason:        "sale",
					UserID:        nil, // Kiosk/Online might not have cashier
				})
			}
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// Assign ReferenceID as Order ID and create movements
		orderIDStr := order.ID.String()
		for i := range movements {
			movements[i].ReferenceID = &orderIDStr
		}
		if len(movements) > 0 {
			if err := tx.Create(&movements).Error; err != nil {
				return err
			}
		}

		// Automatically create a KDS Ticket for F&B orders
		if order.OrderType == "dine_in" || order.OrderType == "takeout" {
			kdsTicket := models.KDSTicket{
				BranchScoped: models.BranchScoped{
					TenantScoped: models.TenantScoped{},
					BranchID:     order.BranchID,
				},
				OrderID: order.ID,
				Status:  "pending",
			}
			if err := tx.Create(&kdsTicket).Error; err != nil {
				return err
			}
		}

		// Feature 5: Automated Loyalty Program with CRM Settings and Tier Upgrades
		if input.CustomerID != nil {
			var customer models.Customer
			if err := tx.Where("id = ?", input.CustomerID).First(&customer).Error; err == nil {
				// 1. Fetch CRM Settings for tenant
				var crmSettings models.CRMSettings
				if err := tx.First(&crmSettings).Error; err != nil {
					crmSettings.PointsPerCurrency = 1.0
					crmSettings.RedemptionRate = 0.01
				}

				// 2. Handle points redemption if requested
				if input.RedeemPoints > 0 && customer.LoyaltyPts >= input.RedeemPoints {
					redeemTx := models.LoyaltyTransaction{
						TenantID:        s.tenantID,
						CustomerID:      *input.CustomerID,
						OrderID:         &order.ID,
						Points:          -input.RedeemPoints,
						TransactionType: "redeemed",
					}
					desc := fmt.Sprintf("Redeemed %.0f points on Order #%s", input.RedeemPoints, order.OrderNumber)
					redeemTx.Description = &desc
					_ = tx.Create(&redeemTx)

					_ = tx.Model(&customer).Update("loyalty_pts", gorm.Expr("loyalty_pts - ?", input.RedeemPoints))
				}

				// 3. Earn points based on order total and CRMSettings
				if order.Total > 0 && status == "completed" {
					pointsRate := crmSettings.PointsPerCurrency
					if pointsRate <= 0 {
						pointsRate = 1.0
					}
					pointsEarned := order.Total * pointsRate

					earnTx := models.LoyaltyTransaction{
						TenantID:        s.tenantID,
						CustomerID:      *input.CustomerID,
						OrderID:         &order.ID,
						Points:          pointsEarned,
						TransactionType: "earned",
					}
					desc := fmt.Sprintf("Earned points on Order #%s", order.OrderNumber)
					earnTx.Description = &desc
					_ = tx.Create(&earnTx)

					newTotalSpend := customer.TotalSpend + order.Total

					// 4. Auto-upgrade customer tier if eligible
					var eligibleTier models.CustomerTier
					var newTierID *uuid.UUID
					if err := tx.Where("min_spend <= ?", newTotalSpend).Order("min_spend DESC").First(&eligibleTier).Error; err == nil {
						newTierID = &eligibleTier.ID
					}

					updates := map[string]interface{}{
						"total_spend": gorm.Expr("total_spend + ?", order.Total),
						"order_count": gorm.Expr("order_count + 1"),
						"loyalty_pts": gorm.Expr("loyalty_pts + ?", pointsEarned),
						"last_visit":  time.Now(),
					}
					if newTierID != nil {
						updates["tier_id"] = newTierID
					}

					if err := tx.Model(&customer).Updates(updates).Error; err != nil {
						return err
					}
				}

				if customer.Phone != nil && *customer.Phone != "" && s.smsService != nil {
					msg := fmt.Sprintf("Thank you for your order at Puxbay! Order #%s for %.2f has been completed.", order.OrderNumber, order.Total)
					desc := fmt.Sprintf("Order Receipt SMS: Order #%s", order.OrderNumber)
					_ = s.smsService.SendTenantSMS(tx, []string{*customer.Phone}, msg, desc)
				}
			}
		}

		// Process Store Credit / BNPL Drawdown if payment was on credit
		var creditAmount float64
		for _, p := range input.Payments {
			if strings.EqualFold(p.Method, "credit") || strings.EqualFold(p.Method, "bnpl") || strings.EqualFold(p.Method, "store_credit") {
				creditAmount += p.Amount
			}
		}
		if creditAmount == 0 && (strings.EqualFold(order.PaymentMethod, "credit") || strings.EqualFold(order.PaymentMethod, "bnpl") || strings.EqualFold(order.PaymentMethod, "store_credit")) {
			creditAmount = order.Total
		}

		if creditAmount > 0 && input.CustomerID != nil {
			var acc models.CreditAccount
			if err := tx.Where("customer_id = ?", *input.CustomerID).First(&acc).Error; err != nil {
				acc = models.CreditAccount{
					CustomerID:  *input.CustomerID,
					CreditLimit: 0,
					Balance:     0,
					Status:      "active",
					DaysToRepay: 30,
				}
				if err := tx.Create(&acc).Error; err != nil {
					return err
				}
			}

			newBalance := acc.Balance + creditAmount
			now := time.Now()
			days := acc.DaysToRepay
			if days <= 0 {
				days = 30
			}
			dueDate := now.AddDate(0, 0, days)

			ref := fmt.Sprintf("BNPL-%s", order.OrderNumber)
			creditTx := models.CreditTransaction{
				CreditAccountID: acc.ID,
				CustomerID:      *input.CustomerID,
				OrderID:         &order.ID,
				Amount:          creditAmount,
				BalanceAfter:    newBalance,
				TransactionType: "drawdown",
				Reference:       ref,
				DueDate:         &dueDate,
				Status:          "pending",
				Notes:           fmt.Sprintf("POS Credit / BNPL Sale #%s", order.OrderNumber),
				CreatedByID:     order.CashierID,
			}

			if err := tx.Create(&creditTx).Error; err != nil {
				return err
			}

			acc.Balance = newBalance
			acc.LastDrawdownAt = &now
			if err := tx.Save(&acc).Error; err != nil {
				return err
			}

			// Update Customer debt_balance
			if err := tx.Model(&models.Customer{}).Where("id = ?", *input.CustomerID).
				Update("debt_balance", newBalance).Error; err != nil {
				return err
			}

			// Create scheduled BNPL Instalment
			inst := models.BNPLInstalment{
				CreditTransactionID: creditTx.ID,
				CustomerID:          *input.CustomerID,
				OrderID:             &order.ID,
				InstalmentNumber:    1,
				TotalInstalments:    1,
				Amount:              creditAmount,
				DueDate:             dueDate,
				Status:              "pending",
			}
			if err := tx.Create(&inst).Error; err != nil {
				return err
			}

			// Send SMS confirmation for BNPL / Store Credit sale
			if s.smsService != nil {
				var cust models.Customer
				if err := tx.Where("id = ?", *input.CustomerID).First(&cust).Error; err == nil && cust.Phone != nil && *cust.Phone != "" {
					msg := fmt.Sprintf("Dear %s, your purchase of GHS %.2f on Store Credit/BNPL (Order #%s) is recorded. Total balance owed: GHS %.2f. Due: %s.", cust.Name, creditAmount, order.OrderNumber, newBalance, dueDate.Format("02 Jan 2006"))
					desc := fmt.Sprintf("BNPL Sale SMS: Order #%s", order.OrderNumber)
					_ = s.smsService.SendTenantSMS(tx, []string{*cust.Phone}, msg, desc)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Preload Branch for the frontend receipt
	s.db.Preload("Branch").First(&order, "id = ?", order.ID)

	return &order, nil
}

// ProcessPOSCheckout handles POS-specific checkout with inventory deduction.
func (s *OrderService) ProcessPOSCheckout(input OrderCreateInput, cashierID *uuid.UUID) (*models.Order, error) {
	order := models.Order{
		OrderNumber:   generateOrderNumber(),
		CustomerID:    input.CustomerID,
		CashierID:     cashierID,
		Subtotal:      input.Subtotal,
		Tax:           input.Tax,
		Discount:      input.Discount,
		Total:         input.Total,
		AmountPaid:    input.AmountPaid,
		PaymentMethod: determinePrimaryPaymentMethod(input.Payments),
		OrderType:     "in_store",
		Status:        "completed",
		PaymentStatus: "paid",
		ReceiptToken:  generateReceiptToken(),
	}

	if input.Notes != "" {
		order.Notes = &input.Notes
	}
	if input.BranchID != nil {
		order.BranchID = input.BranchID
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var movements []models.StockMovement

		for _, itemReq := range input.Items {
			var product models.Product
			if err := tx.Where("id = ?", itemReq.ProductID).First(&product).Error; err != nil {
				return fmt.Errorf("invalid product: %s", itemReq.ProductID)
			}

			orderItem := models.OrderItem{
				ProductID: itemReq.ProductID,
				VariantID: itemReq.VariantID,
				Quantity:  itemReq.Quantity,
				UnitPrice: itemReq.UnitPrice,
				Discount:  itemReq.Discount,
				Total:     itemReq.Total,
				CostPrice: product.CostPrice,
			}
			order.Items = append(order.Items, orderItem)

			// Gap #11: Prevent negative stock with WHERE guard
			if product.TrackInventory {
				result := tx.Model(&models.Product{}).
					Where("id = ? AND current_stock >= ?", product.ID, itemReq.Quantity).
					Update("current_stock", gorm.Expr("current_stock - ?", itemReq.Quantity))
				if result.RowsAffected == 0 {
					return fmt.Errorf("insufficient stock for product %s (available: %.2f, requested: %.2f)",
						product.Name, product.CurrentStock, itemReq.Quantity)
				}

				// Queue stock movement for history
				branchID := uuid.Nil
				if order.BranchID != nil {
					branchID = *order.BranchID
				}
				movements = append(movements, models.StockMovement{
					TenantID:      s.tenantID,
					BranchID:      branchID,
					ProductID:     product.ID,
					VariantID:     itemReq.VariantID,
					Quantity:      -itemReq.Quantity,
					PreviousStock: product.CurrentStock,
					NewStock:      product.CurrentStock - itemReq.Quantity,
					Reason:        "sale",
					UserID:        cashierID,
				})
			}
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// Assign ReferenceID as Order ID and create movements
		orderIDStr := order.ID.String()
		for i := range movements {
			movements[i].ReferenceID = &orderIDStr
		}
		if len(movements) > 0 {
			if err := tx.Create(&movements).Error; err != nil {
				return err
			}
		}

		// Automatically create a KDS Ticket for F&B orders
		if order.OrderType == "dine_in" || order.OrderType == "takeout" || order.OrderType == "kiosk" {
			kdsTicket := models.KDSTicket{
				BranchScoped: models.BranchScoped{
					TenantScoped: models.TenantScoped{},
					BranchID:     order.BranchID,
				},
				OrderID: order.ID,
				Status:  "pending",
			}
			if err := tx.Create(&kdsTicket).Error; err != nil {
				return err
			}
		}

		// Feature 5: Automated Loyalty Program with CRM Settings and Tier Upgrades for POS
		if input.CustomerID != nil {
			var customer models.Customer
			if err := tx.Where("id = ?", input.CustomerID).First(&customer).Error; err == nil {
				// 1. Fetch CRM Settings
				var crmSettings models.CRMSettings
				if err := tx.First(&crmSettings).Error; err != nil {
					crmSettings.PointsPerCurrency = 1.0
					crmSettings.RedemptionRate = 0.01
				}

				// 2. Handle points redemption
				if input.RedeemPoints > 0 && customer.LoyaltyPts >= input.RedeemPoints {
					redeemTx := models.LoyaltyTransaction{
						TenantID:        s.tenantID,
						CustomerID:      *input.CustomerID,
						OrderID:         &order.ID,
						Points:          -input.RedeemPoints,
						TransactionType: "redeemed",
					}
					desc := fmt.Sprintf("Redeemed %.0f points on POS Order #%s", input.RedeemPoints, order.OrderNumber)
					redeemTx.Description = &desc
					_ = tx.Create(&redeemTx)

					_ = tx.Model(&customer).Update("loyalty_pts", gorm.Expr("loyalty_pts - ?", input.RedeemPoints))
				}

				// 3. Earn points based on order total and CRMSettings
				if order.Total > 0 {
					pointsRate := crmSettings.PointsPerCurrency
					if pointsRate <= 0 {
						pointsRate = 1.0
					}
					pointsEarned := order.Total * pointsRate

					earnTx := models.LoyaltyTransaction{
						TenantID:        s.tenantID,
						CustomerID:      *input.CustomerID,
						OrderID:         &order.ID,
						Points:          pointsEarned,
						TransactionType: "earned",
					}
					desc := fmt.Sprintf("Earned points on POS Order #%s", order.OrderNumber)
					earnTx.Description = &desc
					_ = tx.Create(&earnTx)

					newTotalSpend := customer.TotalSpend + order.Total

					// 4. Auto-upgrade customer tier if eligible
					var eligibleTier models.CustomerTier
					var newTierID *uuid.UUID
					if err := tx.Where("min_spend <= ?", newTotalSpend).Order("min_spend DESC").First(&eligibleTier).Error; err == nil {
						newTierID = &eligibleTier.ID
					}

					updates := map[string]interface{}{
						"total_spend": gorm.Expr("total_spend + ?", order.Total),
						"order_count": gorm.Expr("order_count + 1"),
						"loyalty_pts": gorm.Expr("loyalty_pts + ?", pointsEarned),
						"last_visit":  time.Now(),
					}
					if newTierID != nil {
						updates["tier_id"] = newTierID
					}

					if err := tx.Model(&customer).Updates(updates).Error; err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Preload Branch for the frontend receipt
	s.db.Preload("Branch").First(&order, "id = ?", order.ID)

	return &order, nil
}

func generateOrderNumber() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("ORD-%d-%s", time.Now().Unix(), hex.EncodeToString(b))
}

func generateReceiptToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// VoidOrder voids an order and reverses inventory/loyalty if it was completed (Gap #9).
func (s *OrderService) VoidOrder(id string) error {
	var order models.Order
	if err := s.db.Where("id = ?", id).Preload("Items").First(&order).Error; err != nil {
		return errors.New("order not found")
	}

	if order.Status == "voided" {
		return errors.New("order is already voided")
	}

	// Gap #9: Capture the old status BEFORE updating, then use it for reversal logic
	previousStatus := order.Status

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update status to voided
		if err := tx.Model(&order).Update("status", "voided").Error; err != nil {
			return err
		}

		// Reverse inventory and loyalty only if the order was previously completed
		if previousStatus == "completed" {
			for _, item := range order.Items {
				var product models.Product
				if err := tx.Where("id = ?", item.ProductID).First(&product).Error; err == nil && product.TrackInventory {
					tx.Model(&models.Product{}).Where("id = ?", product.ID).Update("current_stock", gorm.Expr("current_stock + ?", item.Quantity))
				}
			}

			// Reverse customer metrics
			if order.CustomerID != nil {
				pointsEarned := order.Total / 10.0
				tx.Model(&models.Customer{}).Where("id = ?", order.CustomerID).Updates(map[string]interface{}{
					"total_spend": gorm.Expr("total_spend - ?", order.Total),
					"order_count": gorm.Expr("order_count - 1"),
					"loyalty_pts": gorm.Expr("loyalty_pts - ?", pointsEarned),
				})
			}
		}
		return nil
	})
}

func (s *OrderService) CompleteOrder(id string) error {
	return s.db.Model(&models.Order{}).Where("id = ?", id).Update("status", "completed").Error
}

func (s *OrderService) GetReceipt(id string) (*models.Order, error) {
	var order models.Order
	// Can query by ID or ReceiptToken
	if err := s.db.Where("id = ? OR receipt_token = ?", id, id).
		Preload("Items.Product").
		Preload("Branch").
		First(&order).Error; err != nil {
		return nil, errors.New("receipt not found")
	}
	return &order, nil
}

func (s *OrderService) RefundOrder(id string, amount float64) error {
	var order models.Order
	if err := s.db.Where("id = ?", id).First(&order).Error; err != nil {
		return errors.New("order not found")
	}

	if order.Status == "voided" || order.Status == "refunded" {
		return errors.New("order is already voided or fully refunded")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if order.CustomerID != nil {
			pointsToDeduct := amount / 10.0
			tx.Model(&models.Customer{}).Where("id = ?", order.CustomerID).Updates(map[string]interface{}{
				"total_spend": gorm.Expr("total_spend - ?", amount),
				"loyalty_pts": gorm.Expr("loyalty_pts - ?", pointsToDeduct),
			})
		}

		if amount >= order.Total {
			return tx.Model(&order).Update("status", "refunded").Error
		}

		return tx.Model(&order).Update("status", "partially_refunded").Error
	})
}

func determinePrimaryPaymentMethod(payments []OrderPaymentInput) string {
	if len(payments) == 0 {
		return "unknown"
	}
	primary := payments[0]
	for _, p := range payments {
		if p.Amount > primary.Amount {
			primary = p
		}
	}
	return primary.Method
}

// Ensure clause import is used (for wallet/transfer row locking pattern reference)
var _ = clause.Locking{}
