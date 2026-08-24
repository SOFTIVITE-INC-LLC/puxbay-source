package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/softivite/puxbay/internal/models"
)

// PrintJobPayload represents the data needed to process a print job.
type PrintJobPayload struct {
	PrintJobID string `json:"print_job_id"`
}

// EnqueuePrintJob enqueues a new print job task.
func EnqueuePrintJob(client *asynq.Client, printJobID string) error {
	payload, err := json.Marshal(PrintJobPayload{PrintJobID: printJobID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypePrintJob, payload)
	info, err := client.Enqueue(task)
	if err != nil {
		return err
	}
	log.Printf("Enqueued print job task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// HandlePrintJob processes the print job task.
func (ws *WorkerServer) HandlePrintJob(ctx context.Context, t *asynq.Task) error {
	var p PrintJobPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("Processing print job: %s", p.PrintJobID)

	// Fetch the print job from the database
	var printJob models.PrintJob
	if err := ws.db.Where("id = ?", p.PrintJobID).First(&printJob).Error; err != nil {
		log.Printf("Print job %s not found: %v", p.PrintJobID, err)
		// If it doesn't exist, we skip retry.
		return fmt.Errorf("print job not found: %w", asynq.SkipRetry)
	}

	// Update status to printed
	if err := ws.db.Model(&printJob).Update("status", "printed").Error; err != nil {
		log.Printf("Failed to update print job %s status: %v", p.PrintJobID, err)
		return err
	}

	log.Printf("Successfully processed print job: %s", p.PrintJobID)
	return nil
}
