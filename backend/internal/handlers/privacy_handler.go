package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type PrivacyHandler struct {
	db *gorm.DB
}

func NewPrivacyHandler(db *gorm.DB) *PrivacyHandler {
	return &PrivacyHandler{db: db}
}

func (h *PrivacyHandler) service(c *gin.Context) *services.PrivacyService {
	return services.NewPrivacyService(getDB(c, h.db))
}

// ExportData requests a GDPR export of the tenant's data (admin only - Gap #40)
func (h *PrivacyHandler) ExportData(c *gin.Context) {
	// Gap #40: Only tenant admins can trigger data exports
	role, _ := c.Get(middleware.ContextKeyRole)
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only tenant admins can request data exports"})
		return
	}

	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, _ := rawTenantID.(uuid.UUID)

	if err := h.service(c).ExportData(tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to request data export"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Data export request received. You will receive an email with your data shortly.",
		"tenant_id": tenantID,
	})
}

// DeleteAccount requests account deletion (admin only - Gap #40)
func (h *PrivacyHandler) DeleteAccount(c *gin.Context) {
	// Gap #40: Only tenant admins can trigger account deletion
	role, _ := c.Get(middleware.ContextKeyRole)
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only tenant admins can delete accounts"})
		return
	}

	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, _ := rawTenantID.(uuid.UUID)

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req) // optional

	if err := h.service(c).DeleteAccount(tenantID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Account deletion scheduled. Your data will be permanently removed within 30 days.",
		"tenant_id": tenantID,
	})
}

func (h *PrivacyHandler) AnonymizeCustomer(c *gin.Context) {
	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, _ := rawTenantID.(uuid.UUID)
	customerID := c.Param("id")

	if err := h.service(c).AnonymizeCustomer(tenantID, customerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to anonymize customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "anonymized"})
}
