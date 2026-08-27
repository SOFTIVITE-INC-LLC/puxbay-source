package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/softivite/puxbay/internal/config"
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

// SendSMS sends an SMS message via Arkesel V2 API.
func (s *SMSService) SendSMS(recipients []string, message string) error {
	if s.config.APIKey == "" {
		log.Println("ℹ️ Arkesel SMS API Key not configured. Skipping SMS dispatch.")
		return nil
	}

	sender := s.config.SenderID
	if sender == "" {
		sender = "PUXBAY"
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

// SendWhatsApp sends a WhatsApp message via Arkesel WhatsApp API with automatic fallback to SMS.
func (s *SMSService) SendWhatsApp(recipientPhone, message string) error {
	if s.config.APIKey == "" {
		log.Println("ℹ️ Arkesel API Key not configured. Skipping WhatsApp dispatch.")
		return nil
	}

	payload := ArkeselWhatsAppPayload{
		Recipient: recipientPhone,
		Message:   message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return s.SendSMS([]string{recipientPhone}, message)
	}

	req, err := http.NewRequest("POST", "https://sms.arkesel.com/api/v2/whatsapp/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return s.SendSMS([]string{recipientPhone}, message)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		// Fallback to regular SMS if WhatsApp endpoint fails or is unconfigured
		if resp != nil {
			resp.Body.Close()
		}
		return s.SendSMS([]string{recipientPhone}, message)
	}
	defer resp.Body.Close()

	return nil
}

// SendOrderConfirmation notifies the customer when an in-store or online order is completed.
func (s *SMSService) SendOrderConfirmation(phone, customerName, orderNumber string, total float64, storeName string) {
	if phone == "" {
		return
	}
	if storeName == "" {
		storeName = "Puxbay Store"
	}
	msg := fmt.Sprintf("Hi %s! Your order #%s of GHS %.2f at %s has been confirmed. Thank you for shopping with us! 🛍️", customerName, orderNumber, total, storeName)
	go func() {
		_ = s.SendWhatsApp(phone, msg)
	}()
}

// SendDeliveryDispatch notifies the customer when their order is dispatched with a driver.
func (s *SMSService) SendDeliveryDispatch(phone, customerName, orderNumber, trkCode, driverName, driverPhone string) {
	if phone == "" {
		return
	}
	msg := fmt.Sprintf("Hi %s! Your order #%s is on the way with driver %s (%s). Track live: https://puxbay.com/track/%s 🚚", customerName, orderNumber, driverName, driverPhone, trkCode)
	go func() {
		_ = s.SendWhatsApp(phone, msg)
	}()
}

// SendLoyaltyPointsEarned alerts the customer to loyalty points accrued.
func (s *SMSService) SendLoyaltyPointsEarned(phone, customerName string, ptsEarned, newBalance float64, storeName string) {
	if phone == "" {
		return
	}
	msg := fmt.Sprintf("Congratulations %s! You just earned %.0f loyalty points at %s. Total balance: %.0f pts (worth GHS %.2f discount). 🌟", customerName, ptsEarned, storeName, newBalance, newBalance*0.1)
	go func() {
		_ = s.SendSMS([]string{phone}, msg)
	}()
}
