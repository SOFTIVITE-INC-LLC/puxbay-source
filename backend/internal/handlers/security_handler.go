package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type SecurityHandler struct {
	db *gorm.DB
}

func NewSecurityHandler(db *gorm.DB) *SecurityHandler {
	return &SecurityHandler{db: db}
}

func (h *SecurityHandler) service(c *gin.Context) *services.SecurityService {
	return services.NewSecurityService(getDB(c, h.db))
}

// Setup2FA generates a secret and QR code for 2FA
func (h *SecurityHandler) Setup2FA(c *gin.Context) {
	rawUserID, _ := c.Get(middleware.ContextKeyUserID)
	userID, _ := rawUserID.(uuid.UUID)

	result, err := h.service(c).Setup2FA(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup 2FA"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Verify2FA verifies the 2FA code and enables 2FA for the user
func (h *SecurityHandler) Verify2FA(c *gin.Context) {
	rawUserID, _ := c.Get(middleware.ContextKeyUserID)
	userID, _ := rawUserID.(uuid.UUID)

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service(c).Verify2FA(userID, req.Code); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

// ListAuditLogs returns a paginated list of audit logs
func (h *SecurityHandler) ListAuditLogs(c *gin.Context) {
	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, _ := rawTenantID.(uuid.UUID)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, total, err := h.service(c).ListAuditLogs(tenantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant_id": tenantID,
		"logs":      logs,
		"total":     total,
	})
}

// Disable2FA disables 2FA for the user
func (h *SecurityHandler) Disable2FA(c *gin.Context) {
	rawUserID, _ := c.Get(middleware.ContextKeyUserID)
	userID, _ := rawUserID.(uuid.UUID)

	if err := h.service(c).Disable2FA(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable 2FA"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

func (h *SecurityHandler) BackupDashboard(c *gin.Context) {
	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, _ := rawTenantID.(uuid.UUID)

	downloadURL, err := h.service(c).BackupDashboard(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to backup dashboard"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"download_url": downloadURL})
}

func (h *SecurityHandler) RestoreBackup(c *gin.Context) {
	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, _ := rawTenantID.(uuid.UUID)

	var req struct {
		BackupURL string `json:"backup_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service(c).RestoreBackup(tenantID, req.BackupURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore backup"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "restored"})
}
