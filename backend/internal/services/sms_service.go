package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type SMSService struct {
	config config.SMSConfig
	client *http.Client
}

func NewSMSService(cfg config.SMSConfig) *SMSService {
	return &SMSService{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type ArkeselSMSPayload struct {
	Sender     string   `json:"sender"`
	Message    string   `json:"message"`
	Recipients []string `json:"recipients"`
}

type ArkeselWhatsAppPayload struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

// SendRawSMS dispatches an SMS directly via Arkesel with the specified sender ID.
func (s *SMSService) SendRawSMS(sender string, recipients []string, message string) error {
	if s.config.APIKey == "" {
		log.Println("ℹ️ Arkesel SMS API Key not configured. Skipping SMS dispatch.")
		return nil
	}

	if sender == "" {
		sender = s.config.SenderID
		if sender == "" {
			sender = "PUXBAY"
		}
	}

	payload := ArkeselSMSPayload{
		Sender:     sender,
		Message:    message,
		Recipients: recipients,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://sms.arkesel.com/api/v2/sms/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("arkesel sms api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SendSystemSMS sends an administrative / onboarding SMS (e.g. staff invites/adding staff)
// using the platform's system credentials without deducting from tenant SMS balance.
func (s *SMSService) SendSystemSMS(recipients []string, message string) error {
	sender := s.config.SenderID
	if sender == "" {
		sender = "PUXBAY"
	}
	return s.SendRawSMS(sender, recipients, message)
}

// SendSMS is a backward-compatible alias for SendSystemSMS.
func (s *SMSService) SendSMS(recipients []string, message string) error {
	return s.SendSystemSMS(recipients, message)
}

// SendTenantSMS dispatches an SMS on behalf of a tenant.
// STRICT RULE: All tenant SMS (orders, carts, campaigns, alerts) MUST use the tenant's
// funded SMS wallet balance. If the balance is insufficient, SMS sending is paused.
func (s *SMSService) SendTenantSMS(db *gorm.DB, recipients []string, message string, description string) error {
	if len(recipients) == 0 || message == "" {
		return nil
	}

	if db == nil {
		log.Println("⚠️ [SMSService] No database context provided for tenant SMS. Skipping.")
		return nil
	}

	requiredCredits := int64(len(recipients))

	// 1. Fetch Tenant SMS Wallet
	var wallet models.SMSWallet
	if err := db.First(&wallet).Error; err != nil {
		log.Println("⚠️ [SMSService] Tenant SMS wallet not found. SMS paused.")
		return fmt.Errorf("tenant SMS wallet not initialized")
	}

	// 2. Strict Balance Check: Pause SMS if balance is insufficient
	if wallet.CreditsBalance < requiredCredits {
		log.Printf("🛑 [SMSService] SMS PAUSED: Insufficient tenant SMS credits (Balance: %d, Required: %d). Skipping message to %v",
			wallet.CreditsBalance, requiredCredits, recipients)
		return nil // Gracefully pause without crashing the caller
	}

	// 3. Resolve Sender ID: Use tenant's approved custom Sender ID or fallback to default
	sender := s.config.SenderID
	if sender == "" {
		sender = "PUXBAY"
	}

	var approvedSender models.SMSSenderID
	if err := db.Where("status = ?", "approved").Order("updated_at desc").First(&approvedSender).Error; err == nil && approvedSender.SenderID != "" {
		sender = approvedSender.SenderID
	}

	// 4. Deduct SMS Credits & Record Transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		wallet.CreditsBalance -= requiredCredits
		wallet.CreditsUsed += requiredCredits
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		txn := models.SMSTransaction{
			TenantID:    wallet.TenantID,
			Type:        "deduction",
			CreditsUsed: requiredCredits,
			PricePerSMS: wallet.PricePerSMS,
			Status:      "completed",
			Description: description,
		}
		return tx.Create(&txn).Error
	})
	if err != nil {
		log.Printf("⚠️ [SMSService] Failed to record credit deduction: %v", err)
		return err
	}

	// 5. Dispatch SMS via Arkesel Gateway
	if err := s.SendRawSMS(sender, recipients, message); err != nil {
		log.Printf("⚠️ [SMSService] Arkesel dispatch error: %v", err)
		return err
	}

	log.Printf("✅ [SMSService] Dispatched %d SMS from sender '%s' (Remaining balance: %d)",
		requiredCredits, sender, wallet.CreditsBalance)
	return nil
}

// SendOrderConfirmation notifies the customer when an in-store or online order is completed using tenant balance.
func (s *SMSService) SendOrderConfirmation(db *gorm.DB, phone, customerName, orderNumber string, total float64, storeName string) {
	if phone == "" {
		return
	}
	if storeName == "" {
		storeName = "Puxbay Store"
	}
	msg := fmt.Sprintf("Hi %s! Your order #%s of GHS %.2f at %s has been confirmed. Thank you for shopping with us! 🛍️", customerName, orderNumber, total, storeName)
	desc := fmt.Sprintf("Order Confirmation SMS: Order #%s", orderNumber)
	go func() {
		_ = s.SendTenantSMS(db, []string{phone}, msg, desc)
	}()
}

// SendDeliveryDispatch notifies the customer when their order is dispatched with a driver using tenant balance.
func (s *SMSService) SendDeliveryDispatch(db *gorm.DB, phone, customerName, orderNumber, trkCode, driverName, driverPhone string) {
	if phone == "" {
		return
	}
	msg := fmt.Sprintf("Hi %s! Your order #%s is on the way with driver %s (%s). Track live: https://puxbay.com/track/%s 🚚", customerName, orderNumber, driverName, driverPhone, trkCode)
	desc := fmt.Sprintf("Delivery Dispatch SMS: Order #%s", orderNumber)
	go func() {
		_ = s.SendTenantSMS(db, []string{phone}, msg, desc)
	}()
}

// SendLoyaltyPointsEarned alerts the customer to loyalty points accrued using tenant balance.
func (s *SMSService) SendLoyaltyPointsEarned(db *gorm.DB, phone, customerName string, ptsEarned, newBalance float64, storeName string) {
	if phone == "" {
		return
	}
	msg := fmt.Sprintf("Congratulations %s! You just earned %.0f loyalty points at %s. Total balance: %.0f pts (worth GHS %.2f discount). 🌟", customerName, ptsEarned, storeName, newBalance, newBalance*0.1)
	desc := "Loyalty Points SMS"
	go func() {
		_ = s.SendTenantSMS(db, []string{phone}, msg, desc)
	}()
}

// SendStorefrontOrderSMS sends order tracking code and status link via SMS
func (s *SMSService) SendStorefrontOrderSMS(db *gorm.DB, phone, customerName, trackingCode string, total float64, storeName, trackURL string) {
	if phone == "" {
		return
	}
	if storeName == "" {
		storeName = "Puxbay Store"
	}
	if customerName == "" {
		customerName = "Valued Customer"
	}
	msg := fmt.Sprintf("Hi %s! Your order of GHS %.2f at %s has been received. Your 8-character tracking code is: %s. Track your order anytime: %s 📦", customerName, total, storeName, trackingCode, trackURL)
	desc := fmt.Sprintf("Storefront Order Tracking SMS: %s", trackingCode)
	go func() {
		_ = s.SendTenantSMS(db, []string{phone}, msg, desc)
	}()
}

// SendOrderStatusUpdateSMS notifies the customer whenever their order fulfillment status changes.
func (s *SMSService) SendOrderStatusUpdateSMS(db *gorm.DB, phone, customerName, orderNumber, status, storeName, trackURL string) {
	if phone == "" {
		return
	}
	if storeName == "" {
		storeName = "Store"
	}
	if customerName == "" {
		customerName = "Valued Customer"
	}

	var statusText string
	switch strings.ToLower(status) {
	case "preparing", "processing", "packaging":
		statusText = "is now being prepared and packed"
	case "ready_for_pickup", "ready":
		statusText = "is READY FOR PICKUP at the store counter! "
	case "out_for_delivery", "shipped", "dispatched":
		statusText = "is OUT FOR DELIVERY to your destination!"
	case "delivered", "completed":
		statusText = "has been DELIVERED & COMPLETED. Thank you for shopping with us!"
	case "cancelled", "voided":
		statusText = "has been CANCELLED. Please contact support if you need assistance."
	default:
		statusText = fmt.Sprintf("status has been updated to %s", strings.ToUpper(status))
	}

	var msg string
	if trackURL != "" {
		msg = fmt.Sprintf("Hi %s! Your order #%s at %s %s. Track live: %s", customerName, orderNumber, storeName, statusText, trackURL)
	} else {
		msg = fmt.Sprintf("Hi %s! Your order #%s at %s %s.", customerName, orderNumber, storeName, statusText)
	}

	desc := fmt.Sprintf("Order Status SMS (#%s -> %s)", orderNumber, status)
	go func() {
		_ = s.SendTenantSMS(db, []string{phone}, msg, desc)
	}()
}
