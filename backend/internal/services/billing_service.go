package services

import (
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type BillingService struct {
	db *gorm.DB
}

func NewBillingService(db *gorm.DB) *BillingService {
	return &BillingService{db: db}
}

func (s *BillingService) ListInvoices(tenantID uuid.UUID) ([]models.BillingPayment, error) {
	var subscription models.Subscription
	if err := s.db.Where("tenant_id = ?", tenantID).First(&subscription).Error; err != nil {
		return []models.BillingPayment{}, nil // No subscription yet, no invoices
	}

	var payments []models.BillingPayment
	if err := s.db.Where("subscription_id = ?", subscription.ID).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (s *BillingService) GetSubscription(tenantID uuid.UUID) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := s.db.Where("tenant_id = ?", tenantID).First(&subscription).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}
