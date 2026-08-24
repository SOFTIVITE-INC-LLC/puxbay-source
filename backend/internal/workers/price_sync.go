package workers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
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

	log.Printf("🔄 Syncing prices for Tenant: %s, Branch: %s", p.TenantID, p.BranchID)
	// TODO: Sync pricing rules

	return nil
}
