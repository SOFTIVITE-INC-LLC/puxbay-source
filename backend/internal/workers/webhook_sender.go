package workers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WebhookPayload defines the data for a webhook delivery job.
type WebhookPayload struct {
	EventID    string                 `json:"event_id"`
	EndpointID uuid.UUID              `json:"endpoint_id"`
	EventType  string                 `json:"event_type"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	Payload    map[string]interface{} `json:"payload"`
}

// NewWebhookDeliveryTask creates an Asynq task for webhook delivery.
func NewWebhookDeliveryTask(tenantID, endpointID uuid.UUID, eventType string, payload map[string]interface{}) (*asynq.Task, error) {
	p, err := json.Marshal(WebhookPayload{
		EventID:    uuid.New().String(),
		EndpointID: endpointID,
		EventType:  eventType,
		TenantID:   tenantID,
		Payload:    payload,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeWebhookDelivery, p,
		asynq.MaxRetry(5),
		asynq.Queue("default"),
		asynq.Timeout(30*time.Second),
	), nil
}

// HandleWebhookDelivery processes the webhook delivery task.
// It resolves the endpoint, checks event filters, signs the payload,
// and delivers with exponential-backoff retries managed by Asynq.
func (ws *WorkerServer) HandleWebhookDelivery(ctx context.Context, t *asynq.Task) error {
	var p WebhookPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	// 1. Load the endpoint
	var endpoint models.WebhookEndpoint
	if err := ws.db.First(&endpoint, "id = ? AND is_active = true", p.EndpointID).Error; err != nil {
		// Endpoint deleted or deactivated — skip silently
		log.Printf("⚠️  Webhook endpoint %s not found or inactive, skipping", p.EndpointID)
		return nil
	}

	// 2. Check that this endpoint subscribes to this event type
	var subscribedEvents []string
	if err := json.Unmarshal([]byte(endpoint.Events), &subscribedEvents); err == nil {
		subscribed := false
		for _, ev := range subscribedEvents {
			if ev == p.EventType || ev == "*" {
				subscribed = true
				break
			}
		}
		if !subscribed {
			log.Printf("ℹ️  Endpoint %s does not subscribe to event %s, skipping", p.EndpointID, p.EventType)
			return nil
		}
	}

	// 3. Build the event envelope
	envelope := map[string]interface{}{
		"id":         p.EventID,
		"event":      p.EventType,
		"tenant_id":  p.TenantID.String(),
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"data":       p.Payload,
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	// 4. Sign with HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(endpoint.Secret))
	mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// 5. Deliver via HTTP POST
	statusCode, responseBody, deliveryErr := sendHTTP(ctx, endpoint.URL, body, signature)

	// 6. Log the delivery attempt
	errMsg := ""
	if deliveryErr != nil {
		errMsg = deliveryErr.Error()
	}
	logEntry := models.WebhookEvent{
		EndpointID:   p.EndpointID,
		EventType:    p.EventType,
		Payload:      datatypes.JSON(body),
		Signature:    &signature,
		StatusCode:   statusCode,
		ResponseBody: responseBody,
		ErrorMessage: nilString(errMsg),
	}
	ws.db.Create(&logEntry)

	// 7. Return error so Asynq's built-in retry handles backoff
	if deliveryErr != nil {
		return fmt.Errorf("webhook delivery to %s failed: %w", endpoint.URL, deliveryErr)
	}
	if statusCode != nil && *statusCode >= 400 {
		return fmt.Errorf("webhook endpoint %s returned HTTP %d", endpoint.URL, *statusCode)
	}

	log.Printf("✅ Webhook %s delivered to %s (HTTP %v)", p.EventType, endpoint.URL, statusCode)
	return nil
}

func sendHTTP(ctx context.Context, url string, body []byte, signature string) (*uint, *string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Puxbay-Signature", signature)
	req.Header.Set("X-Puxbay-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("User-Agent", "Puxbay-Webhooks/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	code := uint(resp.StatusCode)
	respStr := string(respBytes)
	return &code, &respStr, nil
}

func nilString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// DispatchWebhooks queries all active tenant endpoints subscribed to the given event
// and enqueues an Asynq delivery task for each one.
// Call this from handlers after mutating state (order created, product updated, etc.).
func DispatchWebhooks(client *asynq.Client, gormDB *gorm.DB, tenantID uuid.UUID, eventType string, payload map[string]interface{}) {
	var endpoints []models.WebhookEndpoint
	if err := gormDB.Where("tenant_id = ? AND is_active = true", tenantID).Find(&endpoints).Error; err != nil {
		log.Printf("⚠️  DispatchWebhooks: failed to fetch endpoints for tenant %s: %v", tenantID, err)
		return
	}

	for _, ep := range endpoints {
		// Only queue if the endpoint subscribes to this event type or wildcard
		var events []string
		if err := json.Unmarshal([]byte(ep.Events), &events); err != nil {
			continue
		}
		subscribed := false
		for _, ev := range events {
			if ev == eventType || ev == "*" {
				subscribed = true
				break
			}
		}
		if !subscribed {
			continue
		}

		task, err := NewWebhookDeliveryTask(tenantID, ep.ID, eventType, payload)
		if err != nil {
			log.Printf("⚠️  DispatchWebhooks: failed to create task for endpoint %s: %v", ep.ID, err)
			continue
		}
		if _, err := client.Enqueue(task); err != nil {
			log.Printf("⚠️  DispatchWebhooks: failed to enqueue task for endpoint %s: %v", ep.ID, err)
		} else {
			log.Printf("📬 Queued webhook %s → %s", eventType, ep.URL)
		}
	}
}
