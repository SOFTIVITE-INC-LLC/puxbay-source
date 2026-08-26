package workers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/softivite/puxbay/internal/models"
)

type PriceSyncPayload struct {
	TenantID string
	BranchID string
}

func (ws *WorkerServer) HandlePriceSync(ctx context.Context, t *asynq.Task) error {
	var p PriceSyncPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	log.Printf("🔄 Syncing pricing rules for Tenant: %s, Branch: %s", p.TenantID, p.BranchID)

	// Fetch all active/trialing subscriptions that have an associated plan
	var subscriptions []models.Subscription
	result := ws.db.Where(
		"status IN ? AND plan_id IS NOT NULL",
		[]string{"active", "trialing"},
	).Find(&subscriptions)

	if result.Error != nil {
		log.Printf("PriceSync: failed to fetch subscriptions: %v", result.Error)
		return result.Error
	}

	now := time.Now()
	resetCount := 0

	for i := range subscriptions {
		sub := &subscriptions[i]

		// Reset daily API request counter if the reset date has passed
		if now.After(sub.APILastResetDate.Add(24 * time.Hour)) {
			if err := ws.db.Model(sub).Updates(map[string]interface{}{
				"api_requests_today":  0,
				"api_last_reset_date": now,
			}).Error; err != nil {
				log.Printf("PriceSync: failed to reset daily API counter for tenant %s: %v", sub.TenantID, err)
				continue
			}
			resetCount++
		}

		// Reset monthly API request counter if the month reset date has passed
		if now.After(sub.APIMonthResetDate.AddDate(0, 1, 0)) {
			if err := ws.db.Model(sub).Updates(map[string]interface{}{
				"api_requests_this_month": 0,
				"api_month_reset_date":    now,
			}).Error; err != nil {
				log.Printf("PriceSync: failed to reset monthly API counter for tenant %s: %v", sub.TenantID, err)
				continue
			}
		}
	}

	log.Printf("✅ PriceSync complete: %d subscription(s) processed, %d daily counters reset", len(subscriptions), resetCount)
	return nil
}
