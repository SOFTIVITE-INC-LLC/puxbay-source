package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

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

// HandlePrintJob processes the print job task by fetching the target device and
// dispatching the print content over TCP (ESC/POS RAW, port 9100).
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
		return fmt.Errorf("print job not found: %w", asynq.SkipRetry)
	}

	// Look up the online printer device associated with this branch
	var device models.Device
	err := ws.db.Where(
		"branch_id = ? AND device_type = ? AND status = ?",
		printJob.BranchID, "printer", "online",
	).First(&device).Error

	if err != nil {
		// No online printer — mark as failed
		failReason := "no online printer found for branch"
		log.Printf("Print job %s: %s", p.PrintJobID, failReason)
		ws.db.Model(&printJob).Updates(map[string]interface{}{"status": "failed"})
		return fmt.Errorf("%s: %w", failReason, asynq.SkipRetry)
	}

	if device.IPAddress == "" {
		log.Printf("Print job %s: printer device %s has no IP address", p.PrintJobID, device.ID)
		ws.db.Model(&printJob).Updates(map[string]interface{}{"status": "failed"})
		return fmt.Errorf("printer has no IP address: %w", asynq.SkipRetry)
	}

	// Dispatch print job over TCP RAW / ESC-POS (port 9100)
	addr := fmt.Sprintf("%s:9100", device.IPAddress)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		log.Printf("Print job %s: failed to connect to printer at %s: %v", p.PrintJobID, addr, err)
		// Transient — allow retry
		return fmt.Errorf("printer connection failed: %v", err)
	}
	defer conn.Close()

	// Set a write deadline
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Build the ESC/POS payload:
	// ESC @ — initialize printer
	// content bytes
	// ESC d 4 — feed 4 lines
	// GS V 66 — cut paper (partial cut)
	escInit := []byte{0x1B, 0x40}
	escFeed := []byte{0x1B, 0x64, 0x04}
	escCut := []byte{0x1D, 0x56, 0x42, 0x00}

	payload := append(escInit, []byte(printJob.Content)...)
	payload = append(payload, escFeed...)
	payload = append(payload, escCut...)

	if _, err := conn.Write(payload); err != nil {
		log.Printf("Print job %s: failed to write to printer: %v", p.PrintJobID, err)
		return fmt.Errorf("printer write failed: %v", err)
	}

	// Mark as printed with timestamp
	now := time.Now()
	if err := ws.db.Model(&printJob).Updates(map[string]interface{}{
		"status":     "printed",
		"printed_at": now,
	}).Error; err != nil {
		log.Printf("Failed to update print job %s status: %v", p.PrintJobID, err)
		return err
	}

	log.Printf("Successfully dispatched print job %s to %s", p.PrintJobID, addr)
	return nil
}
