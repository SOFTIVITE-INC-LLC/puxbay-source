package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserProfileService struct {
	db *gorm.DB
}

func NewUserProfileService(db *gorm.DB) *UserProfileService {
	return &UserProfileService{db: db}
}

func (s *UserProfileService) ListProfiles(tenantID uuid.UUID) ([]models.UserProfile, error) {
	var profiles []models.UserProfile
	if err := s.db.Where("tenant_id = ?", tenantID).Preload("User").Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

func (s *UserProfileService) GetProfile(id string, tenantID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Preload("User").First(&profile).Error; err != nil {
		return nil, errors.New("user profile not found")
	}
	return &profile, nil
}

type ProfileUpdateInput struct {
	Role                  string
	CanPerformCreditSales *bool
	BaseSalary            float64
	HourlyRate            float64
}

func (s *UserProfileService) UpdateProfile(id string, tenantID uuid.UUID, req ProfileUpdateInput) (*models.UserProfile, error) {
	profile, err := s.GetProfile(id, tenantID)
	if err != nil {
		return nil, err
	}

	if req.Role != "" {
		var role models.Role
		if err := s.db.Where("name ILIKE ?", req.Role).First(&role).Error; err == nil {
			profile.RoleID = role.ID
		}
	}
	if req.CanPerformCreditSales != nil {
		profile.CanPerformCreditSales = *req.CanPerformCreditSales
	}
	if req.BaseSalary > 0 {
		profile.BaseSalary = req.BaseSalary
	}
	if req.HourlyRate > 0 {
		profile.HourlyRate = req.HourlyRate
	}

	if err := s.db.Save(&profile).Error; err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *UserProfileService) SetPOSPin(userID uuid.UUID, tenantID uuid.UUID, pin string) error {
	var profile models.UserProfile
	if err := s.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&profile).Error; err != nil {
		return errors.New("user profile not found")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashStr := string(hash)

	return s.db.Model(&profile).Update("pos_pin", &hashStr).Error
}
