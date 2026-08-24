package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type SecurityService struct {
	db *gorm.DB
}

func NewSecurityService(db *gorm.DB) *SecurityService {
	return &SecurityService{db: db}
}

type Setup2FAResult struct {
	Secret  string `json:"secret"`
	QRCode  string `json:"qr_code"`
	Message string `json:"message"`
}

// generateTOTPSecret creates a cryptographically random 20-byte base32-encoded secret.
func generateTOTPSecret() (string, error) {
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

func (s *SecurityService) Setup2FA(userID uuid.UUID) (*Setup2FAResult, error) {
	secret, err := generateTOTPSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate 2FA secret: %w", err)
	}

	// Store the secret on the user's profile (not yet enabled until verified)
	if err := s.db.Model(&models.UserProfile{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"otp_secret":     secret,
			"is_2fa_enabled": false,
		}).Error; err != nil {
		return nil, fmt.Errorf("failed to save 2FA secret: %w", err)
	}

	// Build the otpauth URI for QR code generation (frontend generates the QR image)
	var profile models.UserProfile
	s.db.Where("user_id = ?", userID).Preload("User").First(&profile)
	otpauthURL := fmt.Sprintf("otpauth://totp/Puxbay:%s?secret=%s&issuer=Puxbay&digits=6&period=30",
		profile.User.Email, secret)

	return &Setup2FAResult{
		Secret:  secret,
		QRCode:  otpauthURL,
		Message: "Scan this QR code with your authenticator app, then verify with a code",
	}, nil
}

func (s *SecurityService) Verify2FA(userID uuid.UUID, code string) error {
	if len(code) != 6 {
		return errors.New("code must be 6 digits")
	}

	var profile models.UserProfile
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return errors.New("user profile not found")
	}

	if profile.OTPSecret == nil || *profile.OTPSecret == "" {
		return errors.New("2FA has not been set up — call setup first")
	}

	// Validate the TOTP code against the stored secret
	if !validateTOTPCode(*profile.OTPSecret, code) {
		return errors.New("invalid verification code")
	}

	// Mark 2FA as enabled
	return s.db.Model(&profile).Update("is_2fa_enabled", true).Error
}

// validateTOTPCode performs TOTP validation (RFC 6238).
// Checks the code against the current and adjacent 30-second windows for clock skew tolerance.
func validateTOTPCode(secret, code string) bool {
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil || len(secretBytes) == 0 {
		return false
	}

	now := time.Now()
	// Check current window and ±1 window for clock skew tolerance
	for _, offset := range []int64{-30, 0, 30} {
		t := now.Add(time.Duration(offset) * time.Second)
		counter := uint64(t.Unix()) / 30
		expected := hmacBasedOTP(secretBytes, counter)
		if expected == code {
			return true
		}
	}
	return false
}

// hmacBasedOTP implements HOTP (RFC 4226) which is the basis of TOTP.
func hmacBasedOTP(key []byte, counter uint64) string {
	// Encode counter as big-endian 8 bytes
	msg := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		msg[i] = byte(counter & 0xff)
		counter >>= 8
	}

	// HMAC-SHA1
	mac := hmac.New(sha1.New, key)
	mac.Write(msg)
	h := mac.Sum(nil)

	// Dynamic truncation (RFC 4226 Section 5.4)
	offset := h[len(h)-1] & 0x0f
	binCode := (uint32(h[offset])&0x7f)<<24 |
		(uint32(h[offset+1])&0xff)<<16 |
		(uint32(h[offset+2])&0xff)<<8 |
		(uint32(h[offset+3]) & 0xff)
	otp := binCode % 1000000

	return fmt.Sprintf("%06d", otp)
}

func (s *SecurityService) Disable2FA(userID uuid.UUID) error {
	return s.db.Model(&models.UserProfile{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"is_2fa_enabled": false,
			"otp_secret":     nil,
		}).Error
}

// ListAuditLogs returns paginated audit logs for a tenant.
func (s *SecurityService) ListAuditLogs(tenantID uuid.UUID, limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := s.db.Model(&models.AuditLog{}).Where("tenant_id = ?", tenantID)
	query.Count(&total)

	err := query.Preload("User").Order("created_at desc").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

// BackupDashboard generates a JSON backup of the tenant's data.
func (s *SecurityService) BackupDashboard(tenantID uuid.UUID) (string, error) {
	var customers []models.Customer
	var orders []models.Order
	var products []models.Product
	var categories []models.Category

	if err := s.db.Find(&customers).Error; err != nil {
		return "", err
	}
	if err := s.db.Find(&orders).Error; err != nil {
		return "", err
	}
	if err := s.db.Find(&products).Error; err != nil {
		return "", err
	}
	if err := s.db.Find(&categories).Error; err != nil {
		return "", err
	}

	backupData := map[string]interface{}{
		"customers":  customers,
		"orders":     orders,
		"products":   products,
		"categories": categories,
	}

	backupBytes, err := json.Marshal(backupData)
	if err != nil {
		return "", err
	}

	file, err := os.CreateTemp("", fmt.Sprintf("backup_%s_*.json", tenantID))
	if err != nil {
		return "", err
	}
	defer file.Close()
	file.Write(backupBytes)

	return fmt.Sprintf("file://%s", file.Name()), nil
}

// RestoreBackup restores tenant data from a backup JSON payload.
func (s *SecurityService) RestoreBackup(tenantID uuid.UUID, backupURL string) error {
	// For simulation, we assume backupURL is a local file path
	filePath := strings.TrimPrefix(backupURL, "file://")
	backupBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %v", err)
	}

	var backupData struct {
		Customers  []models.Customer `json:"customers"`
		Orders     []models.Order    `json:"orders"`
		Products   []models.Product  `json:"products"`
		Categories []models.Category `json:"categories"`
	}
	if err := json.Unmarshal(backupBytes, &backupData); err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Clear existing data to avoid conflicts
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Order{}).Error; err != nil {
			return err
		}
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Product{}).Error; err != nil {
			return err
		}
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Category{}).Error; err != nil {
			return err
		}
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Customer{}).Error; err != nil {
			return err
		}

		for _, c := range backupData.Categories {
			if err := tx.Create(&c).Error; err != nil {
				return err
			}
		}
		for _, p := range backupData.Products {
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
		}
		for _, c := range backupData.Customers {
			if err := tx.Create(&c).Error; err != nil {
				return err
			}
		}
		for _, o := range backupData.Orders {
			if err := tx.Create(&o).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
