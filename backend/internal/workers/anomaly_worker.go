package workers

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

// StartAnomalyDetectionWorker starts a background worker that periodically runs
// anomaly detection across all active tenants and pushes alerts to admins.
func StartAnomalyDetectionWorker(db *gorm.DB, intelligenceSvc *services.IntelligenceService, notifSvc *services.NotificationService) {
	go func() {
		log.Println("🔍 Starting Anomaly Detection worker...")
		// Run once an hour
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			processAnomalyDetection(db, intelligenceSvc, notifSvc)
		}
	}()
}

func processAnomalyDetection(db *gorm.DB, intelligenceSvc *services.IntelligenceService, notifSvc *services.NotificationService) {
	// Fetch all active tenant IDs
	var tenants []models.Tenant
	if err := db.Where("status = ?", "active").Find(&tenants).Error; err != nil {
		log.Printf("[AnomalyWorker] Error fetching tenants: %v", err)
		return
	}

	for _, tenant := range tenants {
		if tenant.SchemaName == "" {
			continue
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(fmt.Sprintf("SET LOCAL search_path TO %s, public", tenant.SchemaName)).Error; err != nil {
				return err
			}

			// Anti-spam: skip if we already sent an anomaly alert in the last hour
			var recentCount int64
			if err := tx.Model(&models.Notification{}).
				Where("category = ? AND created_at > ?", "anomaly", time.Now().Add(-1*time.Hour)).
				Count(&recentCount).Error; err != nil {
				return err
			}
			if recentCount > 0 {
				return nil
			}

			tenantIntelSvc := services.NewIntelligenceService(tx)
			anomalies, err := tenantIntelSvc.DetectAnomalies(tenant.ID.String())
			if err != nil || len(anomalies) == 0 {
				return nil
			}

			// Send one consolidated notification per tenant
			criticalCount := 0
			for _, a := range anomalies {
				if a.Severity == "critical" {
					criticalCount++
				}
			}

			severity := "warning"
			notifType := "warning"
			if criticalCount > 0 {
				severity = "critical"
				notifType = "error"
			}

			title := fmt.Sprintf("%d Anomaly Alert(s) Detected", len(anomalies))
			msg := buildAnomalySummary(anomalies)

			tenantID, err := uuid.Parse(tenant.ID.String())
			if err != nil {
				return nil
			}

			var pushSvc *services.PushService
			if notifSvc != nil {
				pushSvc = notifSvc.GetPushService()
			}
			tenantNotifSvc := services.NewNotificationService(tx, pushSvc)
			tenantNotifSvc.CreateAndPushToAdmins(
				tenantID,
				title,
				msg,
				"anomaly",
				"/intelligence?tab=anomalies",
				notifType,
			)

			log.Printf("[AnomalyWorker] Sent %s anomaly alert for tenant %s (%d anomalies, %d critical)",
				severity, tenant.ID, len(anomalies), criticalCount)
			return nil
		})
		if err != nil {
			log.Printf("[AnomalyWorker] Error processing anomalies for tenant %s: %v", tenant.ID, err)
		}
	}
}

func buildAnomalySummary(anomalies []services.Anomaly) string {
	msg := ""
	for i, a := range anomalies {
		if i >= 3 {
			msg += fmt.Sprintf("...and %d more anomaly/anomalies.\n", len(anomalies)-3)
			break
		}
		prefix := "[WARNING]"
		if a.Severity == "critical" {
			prefix = "[CRITICAL]"
		}
		msg += fmt.Sprintf("%s %s (deviation: %.1f%%)\n", prefix, a.Title, a.Deviation)
	}
	return msg
}
