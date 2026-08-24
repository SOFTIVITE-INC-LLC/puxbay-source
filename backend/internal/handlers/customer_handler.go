package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"github.com/softivite/puxbay/internal/utils"
	"gorm.io/gorm"
)

// CustomerHandler handles customer management for CRM.
type CustomerHandler struct {
	db *gorm.DB
}

// NewCustomerHandler creates a new customer handler.
func NewCustomerHandler(db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{db: db}
}

func (h *CustomerHandler) service(c *gin.Context) *services.CustomerService {
	return services.NewCustomerService(getDB(c, h.db))
}

// List returns all customers for the tenant.
// GET /api/v1/customers
func (h *CustomerHandler) List(c *gin.Context) {
	p := utils.GetPagination(c)

	search := c.Query("q")

	customers, total, err := h.service(c).ListCustomers(search, p.Limit, p.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}

	role, _ := c.Get(middleware.ContextKeyRole)
	permissions := c.GetStringSlice(middleware.ContextKeyPermissions)
	maskedCustomers := utils.MaskCollection(customers, role.(string), permissions)

	c.JSON(http.StatusOK, gin.H{
		"data":  maskedCustomers,
		"total": total,
		"page":  p.Page,
		"limit": p.Limit,
	})
}

// CreateRequest defines the payload for creating a customer.
type CustomerCreateRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

// Create creates a new customer.
// POST /api/v1/customers
func (h *CustomerHandler) Create(c *gin.Context) {
	var req CustomerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	input := services.CustomerInput{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
	}

	customer, err := h.service(c).CreateCustomer(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

// Get returns a specific customer.
// GET /api/v1/customers/:id
func (h *CustomerHandler) Get(c *gin.Context) {
	id := c.Param("id")

	customer, err := h.service(c).GetCustomer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	role, _ := c.Get(middleware.ContextKeyRole)
	permissions := c.GetStringSlice(middleware.ContextKeyPermissions)
	maskedCustomer := utils.MaskCollection(customer, role.(string), permissions)

	c.JSON(http.StatusOK, maskedCustomer)
}

// Update modifies an existing customer.
// PUT /api/v1/customers/:id
func (h *CustomerHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req CustomerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	input := services.CustomerInput{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
	}

	customer, err := h.service(c).UpdateCustomer(id, input)
	if err != nil {
		if err.Error() == "customer not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

// Delete removes a customer.
// DELETE /api/v1/customers/:id
func (h *CustomerHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service(c).DeleteCustomer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

type RecordPaymentRequest struct {
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method"`
	Notes         string  `json:"notes"`
}

func (h *CustomerHandler) RecordPayment(c *gin.Context) {
	id := c.Param("id")

	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if err := h.service(c).RecordPayment(id, req.Amount, req.PaymentMethod, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "payment_recorded"})
}
