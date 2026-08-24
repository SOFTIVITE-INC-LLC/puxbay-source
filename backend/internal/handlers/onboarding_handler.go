package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type OnboardingHandler struct {
	db *gorm.DB
}

func NewOnboardingHandler(db *gorm.DB) *OnboardingHandler {
	return &OnboardingHandler{db: db}
}

// OnboardingStatus calculates how far along the merchant is in setting up their store.
func (h *OnboardingHandler) OnboardingStatus(c *gin.Context) {
	tenantRaw, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tenantID, ok := tenantRaw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// 1. Get Public Tenant & Subscription
	var tenant models.Tenant
	if err := h.db.Preload("Subscription").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenant"})
		return
	}

	// Switch to Tenant Schema for isolated data
	tenantDB := getDB(c, h.db)

	var hasProducts bool
	var hasTheme bool
	var hasPayment bool
	var hasSubscription bool

	// Step 1: Check Products
	var productCount int64
	tenantDB.Model(&models.Product{}).Count(&productCount)
	hasProducts = productCount > 0

	// Step 2 & 3: Check Storefront Settings
	var settings models.StorefrontSettings
	if err := tenantDB.First(&settings).Error; err == nil {
		hasTheme = settings.LogoImage != "" || settings.BannerImage != ""
		hasPayment = settings.EnablePaystack || settings.EnableStripe || settings.EnableMobileMoney
	}

	// Step 4: Check Subscription
	if tenant.Subscription != nil && tenant.Subscription.Status == "active" {
		hasSubscription = true
	}

	// Calculate Progress %
	completed := 0
	if hasProducts {
		completed++
	}
	if hasTheme {
		completed++
	}
	if hasPayment {
		completed++
	}
	if hasSubscription {
		completed++
	}

	progressPercent := (completed * 100) / 4

	c.JSON(http.StatusOK, gin.H{
		"progress_percent": progressPercent,
		"steps": []map[string]interface{}{
			{
				"id":          "add_product",
				"title":       "Add your first product",
				"description": "Create a product to start selling.",
				"completed":   hasProducts,
				"action_url":  "/inventory",
				"action_text": "Add Product",
			},
			{
				"id":          "customize_theme",
				"title":       "Customize your storefront",
				"description": "Upload a logo or banner to make it yours.",
				"completed":   hasTheme,
				"action_url":  "/settings/storefront",
				"action_text": "Customize",
			},
			{
				"id":          "setup_payments",
				"title":       "Set up payments",
				"description": "Connect Paystack or Mobile Money to get paid.",
				"completed":   hasPayment,
				"action_url":  "/settings/payments",
				"action_text": "Configure Payments",
			},
			{
				"id":          "pick_plan",
				"title":       "Pick a pricing plan",
				"description": "Upgrade from your 7-day trial to stay online.",
				"completed":   hasSubscription,
				"action_url":  "/pricing",
				"action_text": "View Plans",
			},
		},
	})
}
