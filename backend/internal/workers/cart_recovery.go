package workers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"time"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

// StartCartRecoveryWorker starts a background worker that checks for abandoned carts
// and sends recovery emails.
func StartCartRecoveryWorker(db *gorm.DB, smtpCfg config.SMTPConfig) {
	go func() {
		log.Println("Starting Abandoned Cart Recovery worker...")
		// Run every 10 minutes. For testing/demo purposes.
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			processAbandonedCarts(db, smtpCfg)
		}
	}()
}

func processAbandonedCarts(db *gorm.DB, smtpCfg config.SMTPConfig) {
	// Look for carts that:
	// 1. Have not been recovered (IsRecovered = false)
	// 2. Have not had an email sent (EmailSent = false)
	// 3. Email is an actual email address (contains '@') - not just a session ID
	// 4. Have been updated more than 24 hours ago (or some threshold)

	threshold := time.Now().Add(-24 * time.Hour)

	var tenants []models.Tenant
	if err := db.Find(&tenants).Error; err != nil {
		log.Printf("[CartWorker] Error fetching tenants: %v", err)
		return
	}

	for _, tenant := range tenants {
		if tenant.SchemaName == "" {
			continue
		}

		// Execute within a transaction using SET LOCAL search_path
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(fmt.Sprintf("SET LOCAL search_path TO %s", tenant.SchemaName)).Error; err != nil {
				return err
			}

			var carts []models.AbandonedCart
			// Using LIKE '%@%' to ensure we only send to actual email addresses
			if err := tx.Where("is_recovered = ? AND email_sent = ? AND email LIKE ? AND updated_at < ?",
				false, false, "%@%", threshold).Find(&carts).Error; err != nil {
				return err
			}

			if len(carts) > 0 {
				log.Printf("[CartWorker] Found %d abandoned carts for schema %s", len(carts), tenant.SchemaName)

				for _, cart := range carts {
					log.Printf("[CartWorker] Sending recovery email to %s for Cart ID %s", cart.Email, cart.ID.String())

					// Send the email using the template
					err := sendRecoveryEmail(smtpCfg, cart.Email)
					if err != nil {
						log.Printf("[CartWorker] Failed to send email to %s: %v", cart.Email, err)
						continue
					}

					// Update the cart to mark email as sent
					cart.EmailSent = true
					if err := tx.Save(&cart).Error; err != nil {
						log.Printf("[CartWorker] Failed to update cart %s: %v", cart.ID.String(), err)
					}
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("[CartWorker] Error processing carts for schema %s: %v", tenant.SchemaName, err)
		}
	}
}

func sendRecoveryEmail(smtpCfg config.SMTPConfig, to string) error {
	tmplData := services.EmailData{
		Title: "You left something behind!",
		Paragraphs: []template.HTML{
			"We noticed you left some items in your shopping cart.",
			"Don&#39;t worry, we&#39;ve saved them for you! Click the link below to complete your purchase.",
		},
		ActionURL:  "https://puxbay.com/store/cart",
		ActionText: "Return to Cart",
		Year:       time.Now().Year(),
	}

	tmpl, err := template.ParseFiles("../../internal/templates/email_base.html")
	if err != nil {
		// Fallback path
		tmpl, err = template.ParseFiles("internal/templates/email_base.html")
		if err != nil {
			return err
		}
	}

	var body bytes.Buffer
	fmt.Fprintf(&body, "To: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n", to, tmplData.Title)
	if err := tmpl.Execute(&body, tmplData); err != nil {
		return err
	}

	if smtpCfg.Host == "" || smtpCfg.Port == "" {
		log.Printf("SMTP not configured, skipping sending email to %s", to)
		return nil
	}

	auth := smtp.PlainAuth("", smtpCfg.User, smtpCfg.Password, smtpCfg.Host)
	addr := fmt.Sprintf("%s:%s", smtpCfg.Host, smtpCfg.Port)
	return smtp.SendMail(addr, auth, smtpCfg.From, []string{to}, body.Bytes())
}
