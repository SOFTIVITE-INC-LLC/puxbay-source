package services

import (
	"errors"
	"fmt"
	"crypto/rand"
	"math/big"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type StaffService struct {
	db           *gorm.DB
	authService  *AuthService
	smsService   *SMSService
	emailService *EmailService
	rootDomain   string
}

func NewStaffService(db *gorm.DB, auth *AuthService, sms *SMSService, email *EmailService, rootDomain string) *StaffService {
	return &StaffService{db: db, authService: auth, smsService: sms, emailService: email, rootDomain: rootDomain}
}

func (s *StaffService) ListStaff(tenantID uuid.UUID, branchID string) ([]models.UserProfile, error) {
	var profiles []models.UserProfile
	query := s.db.Where("tenant_id = ?", tenantID).Preload("User")

	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}

	if err := query.Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

type StaffCreateInput struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
	Role      string
	BranchID  *uuid.UUID
}

func (s *StaffService) CreateStaff(tenantID uuid.UUID, input StaffCreateInput) (*models.UserProfile, error) {
	// 1. Fetch Subscription and Plan Limits
	var sub models.Subscription
	if err := s.db.Where("tenant_id = ? AND status IN ('active', 'trialing')", tenantID).First(&sub).Error; err != nil {
		return nil, errors.New("active subscription required to create staff")
	}

	var currentStaff int64
	s.db.Model(&models.UserProfile{}).Where("tenant_id = ?", tenantID).Count(&currentStaff)

	maxStaff := uint(1)
	if sub.Status == "trialing" {
		maxStaff = 3
	} else if sub.PlanID != nil {
		var pricingPlan models.PricingPlan
		if err := s.db.Where("id = ?", *sub.PlanID).First(&pricingPlan).Error; err == nil {
			maxStaff = uint(pricingPlan.MaxStaff)
		}
	} else if sub.Plan != nil {
		maxStaff = sub.Plan.MaxUsers
	}

	if sub.CustomUsersCount != nil {
		maxStaff = *sub.CustomUsersCount
	}

	if currentStaff >= int64(maxStaff) {
		return nil, fmt.Errorf("staff limit reached. Your plan allows %d staff members. Please upgrade your plan", maxStaff)
	}

	if input.Password == "" {
		input.Password = generateRandomPassword(10)
	}

	hashedPassword, err := s.authService.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	var profile models.UserProfile
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		// Check if user already exists
		if err := tx.Where("username = ? OR email = ?", input.Username, input.Email).First(&user).Error; err == nil {
			// User exists, we just use this user.ID
			// Optionally update fields if needed, but for now just link it
			input.Password = "" // Clear password so we don't send the generated one in email
		} else {
			// Create new user
			user = models.User{
				Username:              input.Username,
				Email:                 input.Email,
				Password:              hashedPassword,
				FirstName:             input.FirstName,
				LastName:              input.LastName,
				Phone:                 input.Phone,
				IsActive:              true,
				RequirePasswordChange: true,
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		}

		var role models.Role
		if err := tx.Where("name ILIKE ?", input.Role).First(&role).Error; err != nil {
			return fmt.Errorf("role not found: %s", input.Role)
		}

		profile = models.UserProfile{
			UserID:   user.ID,
			TenantID: tenantID,
			BranchID: input.BranchID,
			RoleID:   role.ID,
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}

		profile.User = user
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch tenant details for notifications
	var tenant models.Tenant
	if err := s.db.Where("id = ?", tenantID).First(&tenant).Error; err == nil {
		// Build Tenant URL
		protocol := "https"
		if s.rootDomain == "localhost" || s.rootDomain == "localhost:8080" || s.rootDomain == "localhost:5000" {
			protocol = "http"
		}
		tenantURL := fmt.Sprintf("%s://%s.%s", protocol, tenant.Subdomain, s.rootDomain)

		// Send Email Notification
		if input.Email != "" && s.emailService != nil {
			go s.emailService.SendStaffWelcomeEmail(input.Email, tenant.Name, tenantURL, input.Username, input.Password)
		}

		// Send SMS Notification
		if input.Phone != "" && s.smsService != nil {
			var message string
			if input.Password != "" {
				message = fmt.Sprintf("Welcome to %s! Login at %s | User: %s | Pass: %s", tenant.Name, tenantURL, input.Username, input.Password)
			} else {
				message = fmt.Sprintf("Welcome to %s! Login at %s | User: %s. Use your existing password.", tenant.Name, tenantURL, input.Username)
			}
			go s.smsService.SendSMS([]string{input.Phone}, message)
		}
	}

	return &profile, nil
}

func (s *StaffService) GetStaff(id string, tenantID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Preload("User").First(&profile).Error; err != nil {
		return nil, errors.New("staff not found")
	}
	return &profile, nil
}

type StaffUpdateInput struct {
	FirstName string
	LastName  string
	Role      string
	BranchID  *uuid.UUID
}

func (s *StaffService) UpdateStaff(id string, tenantID uuid.UUID, input StaffUpdateInput) (*models.UserProfile, error) {
	profile, err := s.GetStaff(id, tenantID)
	if err != nil {
		return nil, err
	}

	var role models.Role
	if err := s.db.Where("name ILIKE ?", input.Role).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role not found: %s", input.Role)
	}
	profile.RoleID = role.ID
	profile.BranchID = input.BranchID

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(profile).Error; err != nil {
			return err
		}
		profile.User.FirstName = input.FirstName
		profile.User.LastName = input.LastName
		if err := tx.Save(&profile.User).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *StaffService) DeleteStaff(id string, tenantID uuid.UUID) error {
	return s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.UserProfile{}).Error
}


func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return string(b)
}
