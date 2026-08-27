package workers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

// StartReportWorker starts the background cron worker that evaluates active ReportSchedules
// and sends Daily Z-Reports, Weekly P&L, and Monthly Financial Summaries.
func StartReportWorker(db *gorm.DB, smtpCfg config.SMTPConfig) {
	go func() {
		log.Println("📊 Starting Automated Report Dispatch Worker...")
		reportSvc := services.NewReportService(db, smtpCfg)
		emailSvc := services.NewEmailService(db, smtpCfg)

		// Ticker checks every 30 minutes
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			processReportSchedules(db, reportSvc, emailSvc)
		}
	}()
}

func processReportSchedules(db *gorm.DB, reportSvc *services.ReportService, emailSvc *services.EmailService) {
	var schedules []models.ReportSchedule
	if err := db.Where("is_enabled = ?", true).Find(&schedules).Error; err != nil {
		log.Printf("[ReportWorker] Error fetching report schedules: %v", err)
		return
	}

	now := time.Now()

	for _, sched := range schedules {
		// Anti-spam & frequency check
		if sched.LastSentAt != nil && time.Since(*sched.LastSentAt) < 12*time.Hour {
			continue
		}

		recipients := parseRecipients(sched.Recipients)
		if len(recipients) == 0 {
			// Fall back to the account owner's registration email
			ownerEmail := emailSvc.GetPrimaryAdminEmail(sched.TenantID.String())
			if ownerEmail == "" {
				continue
			}
			recipients = []string{ownerEmail}
		}

		switch sched.ReportType {
		case "daily_z":
			// Run Z-Report for today
			zData, err := reportSvc.GenerateDailyZReport(sched.TenantID, now)
			if err == nil {
				html, renderErr := reportSvc.RenderDailyZReportHTML(zData)
				if renderErr == nil {
					subject := fmt.Sprintf("📊 [%s] Daily Z-Report & Sales Audit (%s)", zData.TenantName, zData.Date)
					_ = emailSvc.SendRawHTML(recipients, subject, html)
					db.Model(&sched).Updates(map[string]interface{}{
						"last_sent_at": now,
						"last_status":  "sent",
					})
				}
			}

		case "weekly_pl":
			// Run only on Sundays
			if now.Weekday() == time.Sunday {
				start := now.AddDate(0, 0, -7)
				plData, err := reportSvc.GeneratePLReport(sched.TenantID, "Weekly", start, now)
				if err == nil {
					html, renderErr := reportSvc.RenderPLReportHTML(plData)
					if renderErr == nil {
						subject := fmt.Sprintf("📈 [%s] Weekly Profit & Loss Summary (%s - %s)", plData.TenantName, plData.StartDate, plData.EndDate)
						_ = emailSvc.SendRawHTML(recipients, subject, html)
						db.Model(&sched).Updates(map[string]interface{}{
							"last_sent_at": now,
							"last_status":  "sent",
						})
					}
				}
			}

		case "monthly_pl":
			// Run on the 1st of the month
			if now.Day() == 1 {
				start := now.AddDate(0, -1, 0)
				plData, err := reportSvc.GeneratePLReport(sched.TenantID, "Monthly", start, now)
				if err == nil {
					html, renderErr := reportSvc.RenderPLReportHTML(plData)
					if renderErr == nil {
						subject := fmt.Sprintf("📑 [%s] Monthly Financial P&L Statement (%s - %s)", plData.TenantName, plData.StartDate, plData.EndDate)
						_ = emailSvc.SendRawHTML(recipients, subject, html)
						db.Model(&sched).Updates(map[string]interface{}{
							"last_sent_at": now,
							"last_status":  "sent",
						})
					}
				}
			}
		}
	}
}

func parseRecipients(raw string) []string {
	parts := strings.Split(raw, ",")
	var cleaned []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" && strings.Contains(trimmed, "@") {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}
