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
	CompanyName             string  `json:"company_name"`
	StoreName               string  `json:"store_name"`
	LogoURL                 string  `json:"logo_url"`
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

	if settings.CompanyName == "" && tenant.Name != "" {
		settings.CompanyName = tenant.Name
	}
	if settings.StoreName == "" && settings.CompanyName != "" {
		settings.StoreName = settings.CompanyName
	}
	if settings.LogoURL == "" && tenant.Logo != nil && *tenant.Logo != "" {
		settings.LogoURL = *tenant.Logo
	}

	// Fallback to StorefrontSettings if available
	var sf models.StorefrontSettings
	if err := db.First(&sf).Error; err == nil {
		if settings.StoreName == "" && sf.StoreName != "" {
			settings.StoreName = sf.StoreName
		}
		if settings.LogoURL == "" && sf.LogoImage != "" {
			settings.LogoURL = sf.LogoImage
		}
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

	nameToUpdate := input.CompanyName
	if nameToUpdate == "" {
		nameToUpdate = input.StoreName
	}

	if nameToUpdate != "" {
		_ = db.Exec("UPDATE public.tenants SET name = ? WHERE id = ?", nameToUpdate, tenantID).Error
	}
	if input.LogoURL != "" {
		_ = db.Exec("UPDATE public.tenants SET logo = ? WHERE id = ?", input.LogoURL, tenantID).Error
	}

	if err := db.Exec("UPDATE public.tenants SET metadata = ? WHERE id = ?", b, tenantID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
		return
	}

	// Also sync to StorefrontSettings in the tenant schema
	var sf models.StorefrontSettings
	if err := db.First(&sf).Error; err == nil {
		sfUpdates := map[string]interface{}{}
		if nameToUpdate != "" {
			sfUpdates["store_name"] = nameToUpdate
		}
		if input.LogoURL != "" {
			sfUpdates["logo_image"] = input.LogoURL
		}
		if len(sfUpdates) > 0 {
			_ = db.Model(&sf).Updates(sfUpdates).Error
		}
	}

	c.JSON(http.StatusOK, input)
}
