package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/softivite/puxbay/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/database"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type BillingHandler struct {
	db          *gorm.DB
	paystackCfg *config.PaystackConfig
}

func NewBillingHandler(db *gorm.DB, paystackCfg *config.PaystackConfig) *BillingHandler {
	return &BillingHandler{db: db, paystackCfg: paystackCfg}
}

func (h *BillingHandler) service(c *gin.Context) *services.BillingService {
	return services.NewBillingService(h.db)
}

func (h *BillingHandler) ListInvoices(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	invoices, err := h.service(c).ListInvoices(tenantID.(uuid.UUID))
	if err != nil {
		c.Error(middleware.NewAppError(http.StatusInternalServerError, "Failed to fetch invoices", err))
		return
	}

	c.JSON(http.StatusOK, invoices)
}

func (h *BillingHandler) GetSubscription(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	subscription, err := h.service(c).GetSubscription(tenantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "inactive",
			"plan": gin.H{
				"name":  "No Active Plan",
				"price": 0,
			},
		})
		return
	}

	// Check if the plan is a PricingPlan and map it
	if subscription.PlanID != nil && subscription.Plan == nil {
		var pricingPlan models.PricingPlan
		if err := h.db.Where("id = ?", *subscription.PlanID).First(&pricingPlan).Error; err == nil {
			subscription.Plan = &models.Plan{
				Name:        pricingPlan.Name,
				Price:       pricingPlan.PriceMonthly,
				MaxBranches: uint(pricingPlan.MaxBranches),
			}
		}
	}

	c.JSON(http.StatusOK, subscription)
}

func (h *BillingHandler) ValidatePromo(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var promo models.PromoCode
	if err := h.db.Where("code = ? AND is_active = ?", req.Code, true).First(&promo).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid or expired promo code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"discount_type": promo.DiscountType,
		"discount":      promo.DiscountValue,
		"message":       "Promo applied",
	})
}

func (h *BillingHandler) ListPlans(c *gin.Context) {
	var plans []models.PricingPlan
	if err := h.db.Order("order_index asc").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load plans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
	})
}

func fetchGhsToNgnRate() float64 {
	// Let the rate fetch from the internet.
	// We'll use a public API. If it fails, fallback to a default rate.
	resp, err := http.Get("https://open.er-api.com/v6/latest/GHS")
	if err == nil {
		defer resp.Body.Close()
		var data struct {
			Rates map[string]float64 `json:"rates"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
			if rate, ok := data.Rates["NGN"]; ok && rate > 0 {
				return rate
			}
		}
	}
	return 115.50 // Fallback if API fails
}

func (h *BillingHandler) ProcessPayment(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	planID := c.Param("plan_id")

	var req struct {
		Currency     string `json:"currency"`
		BillingCycle string `json:"billing_cycle"`
		PromoCode    string `json:"promo_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil && c.Request.Method == "POST" {
		// Ignore if it's GET or something else
	}

	// Fallbacks
	if req.Currency == "" {
		req.Currency = c.Query("currency")
		if req.Currency == "" {
			req.Currency = "GHS"
		}
	}
	if req.BillingCycle == "" {
		req.BillingCycle = c.Query("billing_cycle")
		if req.BillingCycle == "" {
			req.BillingCycle = "monthly"
		}
	}

	var plan models.PricingPlan
	if err := h.db.Where("id = ?", planID).First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	price := plan.PriceMonthly
	if req.BillingCycle == "yearly" {
		price = plan.PriceYearly
	}

	// Apply Promo Code if provided
	if req.PromoCode != "" {
		var promo models.PromoCode
		if err := h.db.Where("code = ? AND is_active = ?", req.PromoCode, true).First(&promo).Error; err == nil {
			switch promo.DiscountType {
			case "percentage":
				discount := price * (promo.DiscountValue / 100)
				price = price - discount
			case "flat":
				price = price - promo.DiscountValue
			}
			if price < 0 {
				price = 0
			}
		}
	}

	if req.Currency == "NGN" {
		rate := fetchGhsToNgnRate()
		price = price * rate
	}

	amountInKoboOrPesewas := int64(price * 100)

	// If the plan is free, bypass Paystack entirely
	if amountInKoboOrPesewas == 0 {
		var sub models.Subscription
		if err := h.db.Where("tenant_id = ?", tenantID).First(&sub).Error; err == nil {
			sub.Status = "active"
			// Set expiration far in the future, or based on logic
			newEnd := time.Now().AddDate(10, 0, 0)
			sub.CurrentPeriodEnd = &newEnd
			h.db.Save(&sub)
		}

		c.JSON(http.StatusOK, gin.H{
			"url":     h.paystackCfg.CallbackURL,
			"status":  "success",
			"message": "Free plan activated successfully",
		})
		return
	}

	// Fetch the real admin email for Paystack
	userIDVal, _ := c.Get(middleware.ContextKeyUserID)
	var adminEmail string
	if userIDVal != nil {
		var user models.User
		if err := h.db.Table("public.users").Where("id = ?", userIDVal.(uuid.UUID)).First(&user).Error; err == nil {
			adminEmail = user.Email
		}
	}
	if adminEmail == "" {
		// Fallback: look up admin in public user_profiles linked to this tenant
		var profile struct {
			Email string
		}
		h.db.Table("public.users u").
			Joins("JOIN public.user_profiles p ON p.user_id = u.id").
			Where("p.tenant_id = ? AND p.role IN ('admin', 'superadmin')", tenantID).
			Select("u.email").
			First(&profile)
		adminEmail = profile.Email
	}
	if adminEmail == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find admin email for this tenant"})
		return
	}

	// Build the callback URL using the tenant's subdomain
	var tenantSubdomain string
	var tenantRow struct{ Subdomain string }
	if err := h.db.Table("public.tenants").Select("subdomain").Where("id = ?", tenantID).Scan(&tenantRow).Error; err == nil {
		tenantSubdomain = tenantRow.Subdomain
	}

	callbackURL := h.paystackCfg.CallbackURL
	if tenantSubdomain != "" {
		// Replace the base callback URL scheme+host with the tenant subdomain
		// e.g. http://localhost:4200/billing → http://thinkce.localhost:4200/billing
		//      https://puxbay.com/billing    → https://thinkce.puxbay.com/billing
		parsedBase, err := url.Parse(callbackURL)
		if err == nil {
			parsedBase.Host = tenantSubdomain + "." + parsedBase.Host
			callbackURL = parsedBase.String()
		}
	}

	// Initialize Paystack transaction
	paystackURL := "https://api.paystack.co/transaction/initialize"

	paystackReq := map[string]interface{}{
		"amount":       amountInKoboOrPesewas,
		"email":        adminEmail,
		"currency":     req.Currency,
		"callback_url": callbackURL,
		"metadata": map[string]string{
			"tenant_id": tenantID.(uuid.UUID).String(),
			"plan_id":   planID,
		},
	}

	reqBody, _ := json.Marshal(paystackReq)
	logger.Log.Info(fmt.Sprintf("[Paystack] Initiating checkout: email=%s amount=%d currency=%s secretKeySet=%v", adminEmail, amountInKoboOrPesewas, req.Currency, h.paystackCfg.SecretKey != ""))

	httpReq, err := http.NewRequest("POST", paystackURL, bytes.NewBuffer(reqBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	httpReq.Header.Set("Authorization", "Bearer "+h.paystackCfg.SecretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to communicate with Paystack"})
		return
	}
	defer resp.Body.Close()

	// Read raw body for debugging
	rawBody, _ := io.ReadAll(resp.Body)
	logger.Log.Info(fmt.Sprintf("[Paystack] Response status=%d body=%s", resp.StatusCode, string(rawBody)))

	var paystackResp struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
			AccessCode       string `json:"access_code"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}

	if err := json.Unmarshal(rawBody, &paystackResp); err != nil || !paystackResp.Status {
		logger.Log.Error(fmt.Sprintf("[Paystack] Failed: status=%d message=%s", resp.StatusCode, paystackResp.Message))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Paystack initialization failed", "details": paystackResp.Message})
		return
	}

	mockURL := paystackResp.Data.AuthorizationURL

	c.JSON(http.StatusOK, gin.H{
		"url":      mockURL,
		"amount":   amountInKoboOrPesewas,
		"currency": req.Currency,
	})
}

func (h *BillingHandler) StripeWebhook(c *gin.Context) {
	// Removed
}

// VerifyPayment verifies a Paystack transaction by reference and activates the subscription.
// Called by the frontend after Paystack redirects back via the callback URL.
// This is the fallback for when webhooks can't reach localhost in development.
func (h *BillingHandler) VerifyPayment(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing payment reference"})
		return
	}

	tenantIDVal, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, ok := tenantIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify with Paystack
	verifyURL := "https://api.paystack.co/transaction/verify/" + reference
	httpReq, err := http.NewRequest("GET", verifyURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create verify request"})
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+h.paystackCfg.SecretKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify with Paystack"})
		return
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	logger.Log.Info(fmt.Sprintf("[Paystack] Verify response status=%d body=%s", resp.StatusCode, string(rawBody)))

	var verifyResp struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status    string  `json:"status"`
			Amount    float64 `json:"amount"`
			Reference string  `json:"reference"`
			Currency  string  `json:"currency"`
			Channel   string  `json:"channel"`
			Customer  struct {
				Email        string `json:"email"`
				CustomerCode string `json:"customer_code"`
			} `json:"customer"`
			Metadata map[string]interface{} `json:"metadata"`
			Plan     struct {
				PlanCode string `json:"plan_code"`
				Interval string `json:"interval"`
			} `json:"plan"`
		} `json:"data"`
	}

	if err := json.Unmarshal(rawBody, &verifyResp); err != nil || !verifyResp.Status {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to verify payment", "details": verifyResp.Message})
		return
	}

	if verifyResp.Data.Status != "success" {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Payment not completed", "status": verifyResp.Data.Status})
		return
	}

	// Find the tenant's subscription using the global public-schema DB
	// (NOT the tenant-scoped tx from context, which uses tenant search_path)
	var sub models.Subscription
	if err := database.DB.Table("public.subscriptions").Where("tenant_id = ?", tenantID).First(&sub).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Log.Warn(fmt.Sprintf("[Paystack] Subscription not found for tenant=%s, creating new", tenantID))
			sub = models.Subscription{
				TenantID: tenantID,
				Status:   "active",
			}
			if err := database.DB.Table("public.subscriptions").Create(&sub).Error; err != nil {
				logger.Log.Error(fmt.Sprintf("[Paystack] Failed to create subscription for tenant=%s err=%v", tenantID, err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
				return
			}
		} else {
			logger.Log.Error(fmt.Sprintf("[Paystack] Subscription lookup failed for tenant=%s err=%v", tenantID, err))
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found for tenant"})
			return
		}
	}

	// Activate subscription
	sub.Status = "active"
	if verifyResp.Data.Customer.CustomerCode != "" {
		sub.PaystackCustomerCode = &verifyResp.Data.Customer.CustomerCode
	}

	// Set period end based on plan interval
	newEnd := time.Now()
	if verifyResp.Data.Plan.Interval == "annually" {
		newEnd = newEnd.AddDate(1, 0, 0)
	} else {
		newEnd = newEnd.AddDate(0, 1, 0)
	}
	sub.CurrentPeriodEnd = &newEnd

	// Link to PricingPlan if plan_id in metadata
	if planIDStr, ok := verifyResp.Data.Metadata["plan_id"].(string); ok && planIDStr != "" {
		if planID, err := uuid.Parse(planIDStr); err == nil {
			sub.PlanID = &planID
		}
	}

	database.DB.Table("public.subscriptions").Save(&sub)

	// Create invoice / payment record
	amount := verifyResp.Data.Amount / 100
	ref := verifyResp.Data.Reference
	payment := models.BillingPayment{
		SubscriptionID:    sub.ID,
		Amount:            amount,
		Status:            "succeeded",
		PaystackReference: &ref,
	}
	database.DB.Table("public.billing_payments").Create(&payment)

	logger.Log.Info(fmt.Sprintf("[Paystack] Payment verified and subscription activated for tenant=%s reference=%s", tenantID, reference))

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment verified and subscription activated",
	})
}

func (h *BillingHandler) PaystackWebhook(c *gin.Context) {
	// Verify signature
	signature := c.GetHeader("x-paystack-signature")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature header"})
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	mac := hmac.New(sha512.New, []byte(h.paystackCfg.SecretKey))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if signature != expectedSignature {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Restore body for any subsequent reading if needed, although we can just unmarshal the read body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var req struct {
		Event string `json:"event"`
		Data  struct {
			Reference string `json:"reference"`
			Customer  struct {
				Email string `json:"email"`
			} `json:"customer"`
			Metadata struct {
				TenantID string `json:"tenant_id"`
			} `json:"metadata"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if req.Event == "charge.success" && req.Data.Metadata.TenantID != "" {
		logger.Log.Info(fmt.Sprintf("Paystack payment succeeded for tenant: %v", req.Data.Metadata.TenantID))
		h.db.Model(&models.Subscription{}).
			Where("tenant_id = ?", req.Data.Metadata.TenantID).
			Update("status", "active")
	}

	c.Status(http.StatusOK)
}
