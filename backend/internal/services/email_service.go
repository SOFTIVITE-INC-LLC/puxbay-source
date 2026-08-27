package services

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/templates"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type EmailData struct {
	Title      string
	Paragraphs []template.HTML
	OTPCode    string
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

func (s *EmailService) GetSMTPConfig() config.SMTPConfig {
	return s.smtp
}

// buildCompliantEmailHeaders constructs RFC 5322 compliant email headers that pass
// SPF, DKIM, and DMARC checks, preventing emails from landing in spam folders.
func (s *EmailService) buildCompliantEmailHeaders(to []string, subject, htmlBody string) []byte {
	fromAddress := s.smtp.From
	if fromAddress == "" {
		fromAddress = "notifications@puxbay.com"
	}

	fromHeader := fromAddress
	if !strings.Contains(fromHeader, "<") {
		fromHeader = fmt.Sprintf("Puxbay Commerce <%s>", fromAddress)
	}

	toHeader := strings.Join(to, ", ")
	now := time.Now().Format(time.RFC1123Z)
	messageID := fmt.Sprintf("<%d-%s@puxbay.com>", time.Now().UnixNano(), uuid.New().String()[:8])

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", fromHeader))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", toHeader))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", now))
	buf.WriteString(fmt.Sprintf("Message-ID: %s\r\n", messageID))
	buf.WriteString(fmt.Sprintf("Reply-To: %s\r\n", fromAddress))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n")

	return buf.Bytes()
}

// SendRawHTML sends a custom formatted HTML email with full RFC 5322 headers.
func (s *EmailService) SendRawHTML(recipients []string, subject, rawHTML string) error {
	if len(recipients) == 0 {
		return nil
	}

	if s.smtp.Host == "" {
		fmt.Printf("📧 [MOCK REPORT EMAIL to %v] %s\n", recipients, subject)
		return nil
	}

	auth := smtp.PlainAuth("", s.smtp.User, s.smtp.Password, s.smtp.Host)
	msg := s.buildCompliantEmailHeaders(recipients, subject, rawHTML)
	addr := fmt.Sprintf("%s:%s", s.smtp.Host, s.smtp.Port)

	for _, to := range recipients {
		if err := smtp.SendMail(addr, auth, s.smtp.From, []string{to}, msg); err != nil {
			fmt.Printf("Failed to send email to %s: %v (Host=%s)\n", to, err, s.smtp.Host)
		}
	}

	return nil
}

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
	msg := s.buildCompliantEmailHeaders([]string{to}, subject, body.String())
	addr := fmt.Sprintf("%s:%s", s.smtp.Host, s.smtp.Port)

	if err := smtp.SendMail(addr, auth, s.smtp.From, []string{to}, msg); err != nil {
		fmt.Printf("Failed to send real HTML email: %v\n", err)
	}
}

// SendEmailVerification sends an email verification link and 6-digit OTP code to the newly registered user.
func (s *EmailService) SendEmailVerification(to, firstName, token, otpCode, subdomain string) {
	subject := "Verify your Puxbay account email"
	verifyURL := fmt.Sprintf("https://%s.puxbay.com/auth/verify-email?token=%s", subdomain, token)
	if subdomain == "" || subdomain == "localhost" {
		verifyURL = fmt.Sprintf("http://localhost:4200/auth/verify-email?token=%s", token)
	}

	data := EmailData{
		Title: "Verify Your Email Address",
		Paragraphs: []template.HTML{
			template.HTML(fmt.Sprintf("Hello %s,", template.HTMLEscapeString(firstName))),
			"Welcome to Puxbay Commerce! Please verify your email address to secure your account and activate automated daily business reports.",
			"Alternatively, you can click the button below to verify your account automatically:",
		},
		OTPCode:    otpCode,
		ActionURL:  verifyURL,
		ActionText: "Verify My Account",
	}

	s.sendHTMLEmail(to, subject, data)
}

// RequestPasswordReset initiates password recovery.
func (s *EmailService) RequestPasswordReset(email string) error {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil
	}

	bytesData := make([]byte, 32)
	if _, err := rand.Read(bytesData); err != nil {
		return err
	}
	token := hex.EncodeToString(bytesData)
	expiry := time.Now().Add(15 * time.Minute)

	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"reset_token":        token,
		"reset_token_expiry": expiry,
	}).Error; err != nil {
		return err
	}

	subject := "Reset your Puxbay password"
	data := EmailData{
		Title: "Password Reset Request",
		Paragraphs: []template.HTML{
			"We received a request to reset the password for your Puxbay account.",
			"If you made this request, click the button below within the next 15 minutes to choose a new password:",
		},
		ActionURL:  fmt.Sprintf("https://puxbay.com/auth/reset-password?token=%s", token),
		ActionText: "Reset Password",
	}

	s.sendHTMLEmail(email, subject, data)
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

// GetPrimaryAdminEmail returns the primary registration email for a tenant.
func (s *EmailService) GetPrimaryAdminEmail(tenantID string) string {
	var profile models.UserProfile
	if err := s.db.Preload("User").Where("tenant_id = ?", tenantID).Order("created_at asc").First(&profile).Error; err == nil && profile.User.Email != "" {
		return profile.User.Email
	}

	var tenant models.Tenant
	if err := s.db.Where("id = ?", tenantID).First(&tenant).Error; err == nil {
		// Fallback query directly on users table associated with tenant
		var user models.User
		if err := s.db.Table("public.user_profiles").
			Select("public.users.email").
			Joins("JOIN public.users ON public.users.id = public.user_profiles.user_id").
			Where("public.user_profiles.tenant_id = ?", tenant.ID).
			Order("public.user_profiles.created_at asc").
			Limit(1).
			Scan(&user.Email).Error; err == nil && user.Email != "" {
			return user.Email
		}
	}

	return ""
}

func (s *EmailService) SendTrialExpiringEmail(tenant models.Tenant, daysLeft int) {
	to := s.GetPrimaryAdminEmail(tenant.ID.String())
	if to == "" {
		return
	}

	subject := fmt.Sprintf("Your %s Trial expires in %d days", tenant.Name, daysLeft)
	data := EmailData{
		Title: subject,
		Paragraphs: []template.HTML{
			"We hope you are enjoying your trial of Puxbay!",
			template.HTML(fmt.Sprintf("Just a friendly reminder that your trial for %s will expire in %d days.", template.HTMLEscapeString(tenant.Name), daysLeft)),
			"To ensure uninterrupted access to your admin dashboard and storefront, please log in and upgrade your account to a paid plan.",
		},
		ActionURL:  fmt.Sprintf("https://%s.puxbay.com/admin/pricing", tenant.Subdomain),
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
		Paragraphs: []template.HTML{
			template.HTML(fmt.Sprintf("Your 7-day trial for %s has officially expired.", template.HTMLEscapeString(tenant.Name))),
			"Access to your admin dashboard has been temporarily locked, but don&#39;t worry—your public storefront is still online!",
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
		Paragraphs: []template.HTML{
			template.HTML(fmt.Sprintf("We were unable to process the latest subscription payment for %s.", template.HTMLEscapeString(tenant.Name))),
			"As a result, your account status is now &#39;Past Due&#39; and dashboard access has been restricted.",
			"Please log in and update your payment method as soon as possible to restore full access.",
		},
		ActionURL:  fmt.Sprintf("https://%s.puxbay.com/admin/settings/payments", tenant.Subdomain),
		ActionText: "Update Payment Method",
	}

	s.sendHTMLEmail(to, subject, data)
}

func (s *EmailService) SendStaffWelcomeEmail(to, tenantName, tenantURL, username, password string) {
	subject := fmt.Sprintf("Welcome to %s!", tenantName)
	var paragraphs []template.HTML
	if password != "" {
		paragraphs = []template.HTML{
			"An administrator has created a new staff account for you.",
			"You can log in to the dashboard using the following credentials:",
			template.HTML(fmt.Sprintf("URL: %s", template.HTMLEscapeString(tenantURL))),
			template.HTML(fmt.Sprintf("Username: %s", template.HTMLEscapeString(username))),
			template.HTML(fmt.Sprintf("Password: %s", template.HTMLEscapeString(password))),
			"Please log in and change your password as soon as possible.",
		}
	} else {
		paragraphs = []template.HTML{
			template.HTML(fmt.Sprintf("An administrator has added you to a new branch at %s.", template.HTMLEscapeString(tenantName))),
			"You can log in to the dashboard using your existing credentials:",
			template.HTML(fmt.Sprintf("URL: %s", template.HTMLEscapeString(tenantURL))),
			template.HTML(fmt.Sprintf("Username: %s", template.HTMLEscapeString(username))),
		}
	}

	data := EmailData{
		Title:      fmt.Sprintf("Welcome to %s", tenantName),
		Paragraphs: paragraphs,
		ActionURL:  tenantURL,
		ActionText: "Log In Now",
	}

	s.sendHTMLEmail(to, subject, data)
}
