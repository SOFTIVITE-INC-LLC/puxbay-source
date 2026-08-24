package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

// ArchivalJob handles background archival of old records to prevent table bloat.
type ArchivalJob struct {
	db *gorm.DB
}

// NewArchivalJob creates a new ArchivalJob.
func NewArchivalJob(db *gorm.DB) *ArchivalJob {
	return &ArchivalJob{db: db}
}

// Run executes the archival process.
func (j *ArchivalJob) Run(ctx context.Context) error {
	fmt.Println("[ArchivalJob] Starting archival process...")
	if err := j.archiveStockMovements(); err != nil {
		fmt.Printf("[ArchivalJob] Error archiving stock movements: %v\n", err)
		return err
	}
	fmt.Println("[ArchivalJob] Archival process completed.")
	return nil
}

// archiveStockMovements deletes stock movements older than 1 year (or moves them to a cold storage table).
func (j *ArchivalJob) archiveStockMovements() error {
	oneYearAgo := time.Now().AddDate(-1, 0, 0)

	// In a real scenario with very high volume, this might be chunked to avoid large locks
	result := j.db.Where("created_at < ?", oneYearAgo).Delete(&models.StockMovement{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		fmt.Printf("[ArchivalJob] Archived/Deleted %d old stock movements.\n", result.RowsAffected)
	}
	return nil
}
