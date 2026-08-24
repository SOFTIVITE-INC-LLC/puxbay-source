package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/database"
	"github.com/softivite/puxbay/internal/models"
)

// BillingMiddleware strictly enforces subscription status.
// It checks if the tenant has a valid trial or active subscription.
// If the trial has expired or payment has failed, it returns a 402 Payment Required
// and blocks access to the tenant's admin dashboard endpoints.
func BillingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// Do not block billing portal or storefront routes
		if strings.HasPrefix(path, "/api/v1/billing") || strings.HasPrefix(path, "/api/v1/storefront") {
			c.Next()
			return
		}

		// Retrieve TenantID from context (set by TenantMiddleware or AuthMiddleware)
		tenantIDVal, exists := c.Get(ContextKeyTenantID)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context required for billing validation"})
			c.Abort()
			return
		}

		tenantID, ok := tenantIDVal.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tenant ID context"})
			c.Abort()
			return
		}

		var subscription models.Subscription
		if err := database.DB.Where("tenant_id = ?", tenantID).First(&subscription).Error; err != nil {
			// No subscription record at all
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": "No active subscription found. Please upgrade your account.",
				"code":  "billing_required",
			})
			c.Abort()
			return
		}

		// 1. Check if the trial has expired
		if subscription.Status == "trialing" && subscription.CurrentPeriodEnd != nil {
			if subscription.CurrentPeriodEnd.Before(time.Now()) {
				// Auto-update to past_due
				database.DB.Model(&subscription).Update("status", "past_due")
				subscription.Status = "past_due"
			}
		}

		// 2. Enforce block for invalid statuses
		if subscription.Status == "past_due" || subscription.Status == "canceled" || subscription.Status == "unpaid" || subscription.Status == "incomplete" {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":  "Your subscription is " + subscription.Status + ". Please update your payment details to continue using the platform.",
				"code":   "billing_required",
				"status": subscription.Status,
			})
			c.Abort()
			return
		}

		// If status is active or trialing (and not expired), allow access
		c.Next()
	}
}
