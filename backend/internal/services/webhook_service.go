package services

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type WebhookService struct {
	db *gorm.DB
}

func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{db: db}
}

// DeliverWebhook delivers an event to an endpoint, with automatic retry for failures.
// In a production environment, this should ideally be pushed to an async queue like Asynq or RabbitMQ.
func (s *WebhookService) DeliverWebhook(endpoint models.WebhookEndpoint, event models.WebhookEvent) {
	go func() {
		maxRetries := 5
		backoff := 2 * time.Second

		payload, _ := json.Marshal(event.Payload)

		for attempt := 1; attempt <= maxRetries; attempt++ {
			req, err := http.NewRequest("POST", endpoint.URL, bytes.NewBuffer(payload))
			if err != nil {
				log.Printf("Failed to create webhook request: %v", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			if endpoint.Secret != "" {
				// E.g., signing payload with secret and attaching to header (HMAC)
				req.Header.Set("X-Webhook-Signature", "stub-signature")
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)

			if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				// Success
				resp.Body.Close()
				s.db.Model(&event).Update("status", "delivered")
				return
			}

			if resp != nil {
				resp.Body.Close()
			}

			// Failed, retry
			log.Printf("Webhook delivery to %s failed (attempt %d/%d). Retrying in %v...", endpoint.URL, attempt, maxRetries, backoff)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}

		// Failed permanently
		s.db.Model(&event).Update("status", "failed")
		log.Printf("Webhook delivery to %s permanently failed after %d attempts.", endpoint.URL, maxRetries)
	}()
}
