package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type KioskHandler struct {
	db *gorm.DB
}

func NewKioskHandler(db *gorm.DB) *KioskHandler {
	return &KioskHandler{db: db}
}

// GetConfig returns the kiosk configuration for a given branch
func (h *KioskHandler) GetConfig(c *gin.Context) {
	branchID := c.Param("branch_id")
	// Mock DB fetching
	c.JSON(http.StatusOK, gin.H{
		"branch_id":       branchID,
		"theme":           "dark",
		"features":        []string{"order_ahead", "loyalty_scan"},
		"welcome_message": "Welcome to Puxbay Kiosk",
	})
}

// GetMenu returns the menu items enabled for the kiosk
func (h *KioskHandler) GetMenu(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"items": []gin.H{
			{"id": "1", "name": "Signature Coffee", "price": 4.50, "category": "Beverages"},
			{"id": "2", "name": "Avocado Toast", "price": 8.50, "category": "Food"},
		},
	})
}

type KioskOrderItemParam struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  float64   `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64   `json:"unit_price" binding:"required,gte=0"`
	Discount  float64   `json:"discount"`
	Total     float64   `json:"total" binding:"required,gte=0"`
}

type KioskOrderRequest struct {
	BranchID      string                `json:"branch_id"`
	CustomerID    string                `json:"customer_id"`
	Subtotal      float64               `json:"subtotal"`
	Tax           float64               `json:"tax"`
	Discount      float64               `json:"discount"`
	Total         float64               `json:"total"`
	AmountPaid    float64               `json:"amount_paid"`
	PaymentMethod string                `json:"payment_method"`
	Items         []KioskOrderItemParam `json:"items" binding:"required,min=1"`
}

// PlaceOrder creates a kiosk order without requiring JWT authentication.
// The tenant is resolved from the request subdomain via TenantMiddleware.
func (h *KioskHandler) PlaceOrder(c *gin.Context) {
	var req KioskOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	var items []services.OrderItemInput
	for _, item := range req.Items {
		items = append(items, services.OrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Discount:  item.Discount,
			Total:     item.Total,
		})
	}

	var branchID *uuid.UUID
	if req.BranchID != "" && req.BranchID != "default" {
		if parsed, err := uuid.Parse(req.BranchID); err == nil {
			branchID = &parsed
		}
	}
	if branchID == nil {
		if ctxBranchID, ok := middleware.GetBranchID(c); ok {
			branchID = ctxBranchID
		}
	}

	payments := []services.OrderPaymentInput{
		{Method: req.PaymentMethod, Amount: req.AmountPaid},
	}

	var customerID *uuid.UUID
	if req.CustomerID != "" {
		if parsed, err := uuid.Parse(req.CustomerID); err == nil {
			customerID = &parsed
		}
	}

	input := services.OrderCreateInput{
		BranchID:   branchID,
		CustomerID: customerID,
		Subtotal:   req.Subtotal,
		Tax:        req.Tax,
		Discount:   req.Discount,
		Total:      req.Total,
		AmountPaid: req.AmountPaid,
		Payments:   payments,
		OrderType:  "kiosk",
		Items:      items,
	}

	// Use the tenant-scoped DB from context (set by TenantMiddleware)
	db, _ := c.Get("db")
	tx, _ := db.(*gorm.DB)
	if tx == nil {
		tx = h.db
	}

	var tenantID uuid.UUID
	if tid, exists := c.Get(middleware.ContextKeyTenantID); exists {
		if id, ok := tid.(uuid.UUID); ok {
			tenantID = id
		}
	}

	order, err := services.NewOrderService(tx, nil, tenantID).CreateOrder(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create kiosk order: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

type KioskCustomerRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone"`
}

// RegisterCustomer finds or creates a customer by phone for the kiosk flow.
// No authentication required — tenant is resolved from subdomain.
func (h *KioskHandler) RegisterCustomer(c *gin.Context) {
	var req KioskCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	db, _ := c.Get("db")
	tx, _ := db.(*gorm.DB)
	if tx == nil {
		tx = h.db
	}

	customer, err := services.NewCustomerService(tx).FindOrCreateCustomer(services.CustomerInput{
		Name:  req.Name,
		Phone: req.Phone,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register customer"})
		return
	}

	c.JSON(http.StatusOK, customer)
}
