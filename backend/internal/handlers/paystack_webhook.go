package handlers

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type PaystackWebhookHandler struct {
	db *gorm.DB
}

func NewPaystackWebhookHandler(db *gorm.DB) *PaystackWebhookHandler {
	return &PaystackWebhookHandler{db: db}
}

// PaystackEvent represents the structure of an incoming Paystack webhook payload
type PaystackEvent struct {
	Event string `json:"event"`
	Data  struct {
		Amount    float64 `json:"amount"`
		Currency  string  `json:"currency"`
		Status    string  `json:"status"`
		Reference string  `json:"reference"`
		Customer  struct {
			CustomerCode string `json:"customer_code"`
			Email        string `json:"email"`
		} `json:"customer"`
		Plan struct {
			PlanCode string `json:"plan_code"`
			Interval string `json:"interval"`
		} `json:"plan"`
		Subscription struct {
			SubscriptionCode string `json:"subscription_code"`
			NextPaymentDate  string `json:"next_payment_date"`
		} `json:"subscription"`
		Metadata map[string]interface{} `json:"metadata"`
		PaidAt   string                 `json:"paid_at"`
	} `json:"data"`
}

func (h *PaystackWebhookHandler) Handle(c *gin.Context) {
	secretKey := os.Getenv("PAYSTACK_SECRET_KEY")
	if secretKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Paystack secret key not configured"})
		return
	}

	// 1. Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// 2. Validate HMAC SHA512 signature
	signature := c.GetHeader("x-paystack-signature")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature"})
		return
	}

	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if signature != expectedSignature {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// 3. Parse Event
	var event PaystackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
		return
	}

	// 4. Process Event
	switch event.Event {
	case "charge.success", "invoice.update":
		// Paystack often uses charge.success for standard recurring billing
		if event.Data.Status == "success" || event.Data.Status == "paid" {
			h.handlePaymentSuccess(event)
		}
	case "invoice.payment_failed":
		h.handlePaymentFailed(event)
	case "subscription.disable":
		h.handleSubscriptionCanceled(event)
	}

	// Always return 200 OK to Paystack
	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

func (h *PaystackWebhookHandler) handlePaymentSuccess(event PaystackEvent) {
	var sub models.Subscription

	// Try to find subscription by customer code
	if event.Data.Customer.CustomerCode != "" {
		if err := h.db.Where("paystack_customer_code = ?", event.Data.Customer.CustomerCode).First(&sub).Error; err != nil {
			// Fallback: Check if metadata has tenant_id (first payment scenario)
			if tenantIDStr, ok := event.Data.Metadata["tenant_id"].(string); ok && tenantIDStr != "" {
				if err := h.db.Where("tenant_id = ?", tenantIDStr).First(&sub).Error; err != nil {
					return
				}
				// Save the customer code for future webhooks
				sub.PaystackCustomerCode = &event.Data.Customer.CustomerCode
			} else {
				return
			}
		}
	} else {
		return
	}

	// Update the subscription
	sub.Status = "active"

	// Determine next period end
	if event.Data.Subscription.NextPaymentDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, event.Data.Subscription.NextPaymentDate); err == nil {
			sub.CurrentPeriodEnd = &parsedDate
		}
	} else {
		// Fallback calculation
		newEnd := time.Now()
		if event.Data.Plan.Interval == "annually" {
			newEnd = newEnd.AddDate(1, 0, 0)
		} else {
			// default to monthly
			newEnd = newEnd.AddDate(0, 1, 0)
		}
		sub.CurrentPeriodEnd = &newEnd
	}

	h.db.Save(&sub)

	// Record payment
	var amount = event.Data.Amount / 100 // Paystack sends amounts in pesewas
	ref := event.Data.Reference
	payment := models.BillingPayment{
		SubscriptionID:    sub.ID,
		Amount:            amount,
		Status:            "succeeded",
		PaystackReference: &ref,
	}
	h.db.Create(&payment)
}

func (h *PaystackWebhookHandler) handlePaymentFailed(event PaystackEvent) {
	var sub models.Subscription
	if event.Data.Customer.CustomerCode != "" {
		if err := h.db.Where("paystack_customer_code = ?", event.Data.Customer.CustomerCode).First(&sub).Error; err != nil {
			return
		}
		sub.Status = "past_due"
		h.db.Save(&sub)
	}

	// Record failed payment
	var amount = event.Data.Amount / 100
	ref := event.Data.Reference
	payment := models.BillingPayment{
		SubscriptionID:    sub.ID,
		Amount:            amount,
		Status:            "failed",
		PaystackReference: &ref,
	}
	h.db.Create(&payment)
}

func (h *PaystackWebhookHandler) handleSubscriptionCanceled(event PaystackEvent) {
	var sub models.Subscription
	if event.Data.Customer.CustomerCode != "" {
		if err := h.db.Where("paystack_customer_code = ?", event.Data.Customer.CustomerCode).First(&sub).Error; err != nil {
			return
		}
		sub.Status = "canceled"
		h.db.Save(&sub)
	}
}
