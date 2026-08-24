package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type PublicPortalHandler struct {
	db *gorm.DB
}

func NewPublicPortalHandler(db *gorm.DB) *PublicPortalHandler {
	return &PublicPortalHandler{db: db}
}

func (h *PublicPortalHandler) service(c *gin.Context) *services.PublicPortalService {
	return services.NewPublicPortalService(getDB(c, h.db))
}

// GetTenantInfo returns basic info about a tenant for their public portal
func (h *PublicPortalHandler) GetTenantInfo(c *gin.Context) {
	domain := c.Query("domain")

	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
		return
	}

	info, _ := h.service(c).GetTenantInfo(domain)
	c.JSON(http.StatusOK, info)
}

// ListProducts returns active products for the public portal
func (h *PublicPortalHandler) ListProducts(c *gin.Context) {
	tenantID := c.Query("tenant_id")

	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	products, err := h.service(c).ListProducts(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})
}

// TrackOrder allows a customer to check their order status
func (h *PublicPortalHandler) TrackOrder(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	orderID := c.Query("order_id")

	if tenantID == "" || orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id and order_id are required"})
		return
	}

	result, err := h.service(c).TrackOrder(tenantID, orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

type SubmitPublicFeedbackRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email" binding:"required,email"`
	Rating  uint   `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

func (h *PublicPortalHandler) SubmitFeedback(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	var req SubmitPublicFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.PublicFeedbackInput{
		Name:    req.Name,
		Email:   req.Email,
		Rating:  req.Rating,
		Comment: req.Comment,
	}

	feedback, err := h.service(c).SubmitFeedback(tenantID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	c.JSON(http.StatusCreated, feedback)
}
