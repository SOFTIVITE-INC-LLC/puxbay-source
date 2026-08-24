package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type PaymentMethodHandler struct {
	db *gorm.DB
}

func NewPaymentMethodHandler(db *gorm.DB) *PaymentMethodHandler {
	return &PaymentMethodHandler{db: db}
}

func (h *PaymentMethodHandler) getDB(c *gin.Context) *gorm.DB {
	return getDB(c, h.db)
}

// List returns all payment methods.
func (h *PaymentMethodHandler) List(c *gin.Context) {
	var methods []models.PaymentMethod
	if err := h.getDB(c).Order("name ASC").Find(&methods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment methods"})
		return
	}
	c.JSON(http.StatusOK, methods)
}

// Create creates a new payment method.
func (h *PaymentMethodHandler) Create(c *gin.Context) {
	var method models.PaymentMethod
	if err := c.ShouldBindJSON(&method); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.getDB(c).Create(&method).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment method"})
		return
	}
	c.JSON(http.StatusCreated, method)
}

// Update updates a payment method.
func (h *PaymentMethodHandler) Update(c *gin.Context) {
	id := c.Param("id")
	methodID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method ID"})
		return
	}

	var existing models.PaymentMethod
	if err := h.getDB(c).First(&existing, "id = ?", methodID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment method not found"})
		return
	}

	var req models.PaymentMethod
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.getDB(c).Model(&existing).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment method"})
		return
	}

	c.JSON(http.StatusOK, existing)
}

// Delete deletes a payment method.
func (h *PaymentMethodHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	methodID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method ID"})
		return
	}
	if err := h.getDB(c).Delete(&models.PaymentMethod{}, "id = ?", methodID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment method not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
