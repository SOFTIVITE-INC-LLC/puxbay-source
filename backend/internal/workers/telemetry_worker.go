package workers

import (
	"log"
	"time"

	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

// StartTelemetryCleanupWorker starts a background worker that runs periodically to delete
// telemetry logs older than 7 days.
func StartTelemetryCleanupWorker(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			log.Println("🧹 Running telemetry cleanup worker...")
			sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
			
			// Delete logs older than 7 days
			result := db.Where("created_at < ?", sevenDaysAgo).Delete(&models.TelemetryLog{})
			if result.Error != nil {
				log.Printf("❌ Failed to cleanup telemetry logs: %v", result.Error)
			} else if result.RowsAffected > 0 {
				log.Printf("✅ Cleaned up %d old telemetry logs", result.RowsAffected)
			}
		}
	}()
}
