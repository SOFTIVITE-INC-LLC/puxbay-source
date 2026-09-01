package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type KioskHandler struct {
	db          *gorm.DB
	smsService  *services.SMSService
	pushService *services.PushService
}

func NewKioskHandler(db *gorm.DB, sms *services.SMSService, push *services.PushService) *KioskHandler {
	return &KioskHandler{db: db, smsService: sms, pushService: push}
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
	CustomerName  string                `json:"customer_name"`
	CustomerPhone string                `json:"customer_phone"`
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

	// Use the tenant-scoped DB from context (set by TenantMiddleware)
	db, _ := c.Get("db")
	tx, _ := db.(*gorm.DB)
	if tx == nil {
		tx = h.db
	}

	var customerID *uuid.UUID
	customerName := strings.TrimSpace(req.CustomerName)
	customerPhone := strings.TrimSpace(req.CustomerPhone)

	if req.CustomerID != "" {
		if parsed, err := uuid.Parse(req.CustomerID); err == nil {
			customerID = &parsed
			var cust models.Customer
			if tx.First(&cust, "id = ?", parsed).Error == nil {
				if customerName == "" {
					customerName = cust.Name
				}
				if customerPhone == "" && cust.Phone != nil {
					customerPhone = *cust.Phone
				}
			}
		}
	}

	if customerPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer contact phone number is compulsory for kiosk orders"})
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

	input := services.OrderCreateInput{
		BranchID:      branchID,
		CustomerID:    customerID,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
		Subtotal:      req.Subtotal,
		Tax:           req.Tax,
		Discount:      req.Discount,
		Total:         req.Total,
		AmountPaid:    req.AmountPaid,
		Payments:      payments,
		OrderType:     "kiosk",
		Items:         items,
	}

	var tenantID uuid.UUID
	if tid, exists := c.Get(middleware.ContextKeyTenantID); exists {
		if id, ok := tid.(uuid.UUID); ok {
			tenantID = id
		}
	}

	order, err := services.NewOrderService(tx, h.smsService, tenantID).CreateOrder(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create kiosk order: " + err.Error()})
		return
	}

	// Dispatch SMS to customer with tracking code
	if h.smsService != nil && customerPhone != "" {
		msg := fmt.Sprintf("Thank you for your order! Your Kiosk Order #%s has been placed. Please present this code when called.", order.OrderNumber)
		_ = h.smsService.SendTenantSMS(tx, []string{customerPhone}, msg, "Kiosk Order SMS: #"+order.OrderNumber)
	}

	// Dispatch Real-time Notification & Sound to Admins/Cashiers
	if h.pushService != nil && tenantID != uuid.Nil {
		custDisplay := customerName
		if custDisplay == "" {
			custDisplay = "Customer"
		}
		h.pushService.SendToTenantAdminsWithSound(
			tenantID,
			fmt.Sprintf("New Kiosk Order #%s", order.OrderNumber),
			fmt.Sprintf("Kiosk order for GH₵%.2f placed by %s", order.Total, custDisplay),
			"kiosk_order",
			"/orders/"+order.ID.String(),
			"kiosk_order",
			"kiosk_order",
		)
	}

	c.JSON(http.StatusCreated, order)
}

type KioskCustomerRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

// RegisterCustomer finds or creates a customer by phone for the kiosk flow.
// No authentication required — tenant is resolved from subdomain.
func (h *KioskHandler) RegisterCustomer(c *gin.Context) {
	var req KioskCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both customer name and phone number are compulsory"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register customer: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}
