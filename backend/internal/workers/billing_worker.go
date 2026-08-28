package workers

import (
	"log"
	"time"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StartBillingWorker starts a background worker that checks for trial expirations
// and billing issues, then sends automated emails.
func StartBillingWorker(db *gorm.DB, cfg config.SMTPConfig) {
	go func() {
		log.Println("Starting Billing & Trial worker...")
		// Run every 24 hours in production. We use 1 hour for dev/demo purposes.
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		emailSvc := services.NewEmailService(db, cfg)

		for {
			<-ticker.C
			processBillingEmails(db, emailSvc)
		}
	}()
}

func processBillingEmails(db *gorm.DB, emailSvc *services.EmailService) {
	var subscriptions []models.Subscription
	// Only fetch subscriptions that are trialing or past_due
	if err := db.Preload("Tenant").Where("status IN ('trialing', 'past_due')").Find(&subscriptions).Error; err != nil {
		log.Printf("[BillingWorker] Error fetching subscriptions: %v", err)
		return
	}

	now := time.Now()

	for _, sub := range subscriptions {
		// Process each subscription in its own transaction to lock the row
		db.Transaction(func(tx *gorm.DB) error {
			var lockedSub models.Subscription
			// Lock the row to prevent other workers from processing it simultaneously
			if err := tx.Preload("Tenant").Where("id = ?", sub.ID).
				Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&lockedSub).Error; err != nil {
				return err
			}

			// Skip if there's no tenant linked or no end date
			if lockedSub.TenantID.String() == "" || lockedSub.CurrentPeriodEnd == nil {
				return nil
			}

			// Anti-spam: ensure we only send one billing/trial email per 24 hours max
			if lockedSub.LastBillingEmailAt != nil && time.Since(*lockedSub.LastBillingEmailAt) < 24*time.Hour {
				return nil
			}

			emailSent := false

			switch lockedSub.Status {
			case "trialing":
				daysLeft := int(lockedSub.CurrentPeriodEnd.Sub(now).Hours() / 24)

				// Trigger 3-day warning
				if daysLeft == 3 {
					emailSvc.SendTrialExpiringEmail(lockedSub.Tenant, 3)
					emailSent = true
				} else if daysLeft == 1 {
					// Trigger 1-day warning
					emailSvc.SendTrialExpiringEmail(lockedSub.Tenant, 1)
					emailSent = true
				} else if daysLeft <= 0 && daysLeft >= -1 {
					// Trigger expired email if passed today
					emailSvc.SendTrialExpiredEmail(lockedSub.Tenant)
					tx.Model(&lockedSub).Update("status", "past_due")
					emailSent = true
				}
			case "past_due":
				emailSvc.SendPaymentFailedEmail(lockedSub.Tenant)
				emailSent = true
			}

			if emailSent {
				tx.Model(&lockedSub).Update("last_billing_email_at", time.Now())
			}
			return nil
		})
	}
}
