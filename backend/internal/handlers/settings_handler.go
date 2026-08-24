package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	db *gorm.DB
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

type GlobalSettingsResponse struct {
	Currency                string  `json:"currency"`
	Timezone                string  `json:"timezone"`
	DateFormat              string  `json:"date_format"`
	EnableEmailReceipts     bool    `json:"enable_email_receipts"`
	HardwareProxyURL        string  `json:"hardware_proxy_url"`
	EnableHardwareProxy     bool    `json:"enable_hardware_proxy"`
	AutoPrintReceipts       bool    `json:"auto_print_receipts"`
	EnableSMSNotifications  bool    `json:"enable_sms_notifications"`
	EnablePushNotifications bool    `json:"enable_push_notifications"`
	AdminNotificationEmail  string  `json:"admin_notification_email"`
	PromoThreshold          float64 `json:"promo_threshold"`
	PromoDiscountPercent    float64 `json:"promo_discount_percent"`
}

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var tenant models.Tenant

	// db is already scoped to the tenant's schema by the middleware,
	// but to get the tenant metadata we fetch the active tenant.
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant id not in context"})
		return
	}

	// Fetch from public.tenants
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant not found"})
		return
	}

	var settings GlobalSettingsResponse
	if len(tenant.Metadata) > 0 {
		_ = json.Unmarshal(tenant.Metadata, &settings)
	}

	if settings.Currency == "" {
		settings.Currency = "GHS"
	}
	if settings.Timezone == "" {
		settings.Timezone = "UTC"
	}

	c.JSON(http.StatusOK, settings)
}

func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var input GlobalSettingsResponse
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant id not in context"})
		return
	}

	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant not found"})
		return
	}

	b, _ := json.Marshal(input)
	tenant.Metadata = b

	if err := db.Exec("UPDATE public.tenants SET metadata = ? WHERE id = ?", b, tenantID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
		return
	}

	c.JSON(http.StatusOK, input)
}
