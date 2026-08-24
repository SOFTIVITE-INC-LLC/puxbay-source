package services

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"time"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/templates"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type EmailData struct {
	Title      string
	Paragraphs []string
	ActionURL  string
	ActionText string
	Year       int
}

type EmailService struct {
	db   *gorm.DB
	smtp config.SMTPConfig
}

func NewEmailService(db *gorm.DB, smtpCfg config.SMTPConfig) *EmailService {
	return &EmailService{db: db, smtp: smtpCfg}
}

func (s *EmailService) RequestPasswordReset(email string) error {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		// Do not leak that user doesn't exist
		return nil
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	token := hex.EncodeToString(bytes)
	expiry := time.Now().Add(15 * time.Minute)

	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"reset_token":        token,
		"reset_token_expiry": expiry,
	}).Error; err != nil {
		return err
	}

	auth := smtp.PlainAuth("", s.smtp.User, s.smtp.Password, s.smtp.Host)
	msg := []byte("Subject: Password Reset Request\r\n\r\n" +
		"Use this token to reset your password: " + token + "\r\n")

	addr := fmt.Sprintf("%s:%s", s.smtp.Host, s.smtp.Port)
	if s.smtp.Host != "" {
		if err := smtp.SendMail(addr, auth, s.smtp.From, []string{email}, msg); err != nil {
			fmt.Printf("Failed to send real email: %v (Simulating fallback)\n", err)
		}
	} else {
		fmt.Printf("SMTP not configured. Token for %s is %s\n", email, token)
	}

	return nil
}

func (s *EmailService) ResetPassword(token, newPassword string) error {
	var user models.User
	if err := s.db.Where("reset_token = ? AND reset_token_expiry > ?", token, time.Now()).First(&user).Error; err != nil {
		return errors.New("invalid or expired token")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Model(&user).Updates(map[string]interface{}{
		"password":           string(hash),
		"reset_token":        nil,
		"reset_token_expiry": nil,
	}).Error
}

// ---------------------------------------------------------
// Billing & Trial Emails
// ---------------------------------------------------------

func (s *EmailService) sendHTMLEmail(to string, subject string, data EmailData) {
	data.Year = time.Now().Year()

	tmpl, err := template.ParseFS(templates.FS, "email_base.html")
	if err != nil {
		fmt.Printf("Failed to parse email template: %v\n", err)
		return
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		fmt.Printf("Failed to execute email template: %v\n", err)
		return
	}

	if s.smtp.Host == "" {
		fmt.Printf("📧 [MOCK HTML EMAIL to %s] %s\n", to, subject)
		return
	}

	auth := smtp.PlainAuth("", s.smtp.User, s.smtp.Password, s.smtp.Host)

	headers := "MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		fmt.Sprintf("Subject: %s\r\n\r\n", subject)

	msg := append([]byte(headers), body.Bytes()...)

	addr := fmt.Sprintf("%s:%s", s.smtp.Host, s.smtp.Port)
	if err := smtp.SendMail(addr, auth, s.smtp.From, []string{to}, msg); err != nil {
		fmt.Printf("Failed to send real HTML email: %v\n", err)
	}
}

// GetPrimaryAdminEmail tries to find the main admin user for a tenant.
func (s *EmailService) GetPrimaryAdminEmail(tenantID string) string {
	var profile models.UserProfile
	if err := s.db.Preload("User").Where("tenant_id = ? AND role IN ('admin', 'superadmin')", tenantID).First(&profile).Error; err != nil {
		return ""
	}
	return profile.User.Email
}

func (s *EmailService) SendTrialExpiringEmail(tenant models.Tenant, daysLeft int) {
	to := s.GetPrimaryAdminEmail(tenant.ID.String())
	if to == "" {
		return
	}

	subject := fmt.Sprintf("Your %s Trial expires in %d days", tenant.Name, daysLeft)
	data := EmailData{
		Title: subject,
		Paragraphs: []string{
			"We hope you are enjoying your trial of Puxbay!",
			fmt.Sprintf("Just a friendly reminder that your trial for %s will expire in %d days.", tenant.Name, daysLeft),
			"To ensure uninterrupted access to your admin dashboard and storefront, please log in and upgrade your account to a paid plan.",
		},
		ActionURL:  fmt.Sprintf("https://%s.puxbay.com/admin/pricing", tenant.Subdomain), // Assuming domain structure
		ActionText: "Upgrade Now",
	}

	s.sendHTMLEmail(to, subject, data)
}

func (s *EmailService) SendTrialExpiredEmail(tenant models.Tenant) {
	to := s.GetPrimaryAdminEmail(tenant.ID.String())
	if to == "" {
		return
	}

	subject := fmt.Sprintf("Action Required: Your %s Trial has expired", tenant.Name)
	data := EmailData{
		Title: "Trial Expired",
		Paragraphs: []string{
			fmt.Sprintf("Your 7-day trial for %s has officially expired.", tenant.Name),
			"Access to your admin dashboard has been temporarily locked, but don't worry—your public storefront is still online!",
			"To unlock your dashboard and continue managing your business, please log in and choose a pricing plan.",
		},
		ActionURL:  fmt.Sprintf("https://%s.puxbay.com/admin/pricing", tenant.Subdomain),
		ActionText: "Choose a Plan",
	}

	s.sendHTMLEmail(to, subject, data)
}

func (s *EmailService) SendPaymentFailedEmail(tenant models.Tenant) {
	to := s.GetPrimaryAdminEmail(tenant.ID.String())
	if to == "" {
		return
	}

	subject := fmt.Sprintf("Payment Failed: Action required for %s", tenant.Name)
	data := EmailData{
		Title: "Payment Failed",
		Paragraphs: []string{
			fmt.Sprintf("We were unable to process the latest subscription payment for %s.", tenant.Name),
			"As a result, your account status is now 'Past Due' and dashboard access has been restricted.",
			"Please log in and update your payment method as soon as possible to restore full access.",
		},
		ActionURL:  fmt.Sprintf("https://%s.puxbay.com/admin/settings/payments", tenant.Subdomain),
		ActionText: "Update Payment Method",
	}

	s.sendHTMLEmail(to, subject, data)
}

func (s *EmailService) SendStaffWelcomeEmail(to, tenantName, tenantURL, username, password string) {
	subject := fmt.Sprintf("Welcome to %s!", tenantName)
	var paragraphs []string
	if password != "" {
		paragraphs = []string{
			"An administrator has created a new staff account for you.",
			"You can log in to the dashboard using the following credentials:",
			fmt.Sprintf("URL: %s", tenantURL),
			fmt.Sprintf("Username: %s", username),
			fmt.Sprintf("Password: %s", password),
			"Please log in and change your password as soon as possible.",
		}
	} else {
		paragraphs = []string{
			fmt.Sprintf("An administrator has added you to a new branch at %s.", tenantName),
			"You can log in to the dashboard using your existing credentials:",
			fmt.Sprintf("URL: %s", tenantURL),
			fmt.Sprintf("Username: %s", username),
		}
	}

	data := EmailData{
		Title: fmt.Sprintf("Welcome to %s", tenantName),
		Paragraphs: paragraphs,
		ActionURL:  tenantURL,
		ActionText: "Log In Now",
	}

	s.sendHTMLEmail(to, subject, data)
}
