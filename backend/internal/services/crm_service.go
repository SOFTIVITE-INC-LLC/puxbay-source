package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type CRMService struct {
	db *gorm.DB
}

func NewCRMService(db *gorm.DB) *CRMService {
	return &CRMService{db: db}
}

func (s *CRMService) ListLoyaltyTransactions(customerID string) ([]models.LoyaltyTransaction, error) {
	query := s.db.Model(&models.LoyaltyTransaction{})
	if customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}

	var transactions []models.LoyaltyTransaction
	if err := query.Order("created_at DESC").Limit(50).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *CRMService) ListGiftCards(status string) ([]models.GiftCard, error) {
	query := s.db.Model(&models.GiftCard{})
	if status != "" {
		query = query.Where("is_active = ?", status == "active")
	}

	var giftCards []models.GiftCard
	if err := query.Order("created_at DESC").Find(&giftCards).Error; err != nil {
		return nil, err
	}
	return giftCards, nil
}

func (s *CRMService) CreateGiftCard(giftCard *models.GiftCard) error {
	giftCard.IsActive = true
	return s.db.Create(giftCard).Error
}

func (s *CRMService) GetCRMSettings() (*models.CRMSettings, error) {
	var settings models.CRMSettings
	if err := s.db.First(&settings).Error; err != nil {
		settings = models.CRMSettings{
			PointsPerCurrency: 1.0,
			RedemptionRate:    0.01,
		}
		if err := s.db.Create(&settings).Error; err != nil {
			return nil, err
		}
	}
	return &settings, nil
}

func (s *CRMService) UpdateCRMSettings(pointsPerCurrency, redemptionRate float64) (*models.CRMSettings, error) {
	var settings models.CRMSettings
	if err := s.db.First(&settings).Error; err != nil {
		settings = models.CRMSettings{}
		s.db.Create(&settings)
	}

	settings.PointsPerCurrency = pointsPerCurrency
	settings.RedemptionRate = redemptionRate
	if err := s.db.Save(&settings).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *CRMService) ListCustomerTiers() ([]models.CustomerTier, error) {
	var tiers []models.CustomerTier
	if err := s.db.Order("min_spend ASC").Find(&tiers).Error; err != nil {
		return nil, err
	}
	return tiers, nil
}

func (s *CRMService) CreateCustomerTier(tier *models.CustomerTier) error {
	return s.db.Create(tier).Error
}

func (s *CRMService) UpdateCustomerTier(id string, input models.CustomerTier) (*models.CustomerTier, error) {
	tierID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid tier ID")
	}

	var tier models.CustomerTier
	if err := s.db.Where("id = ?", tierID).First(&tier).Error; err != nil {
		return nil, errors.New("tier not found")
	}

	tier.Name = input.Name
	tier.MinSpend = input.MinSpend
	tier.DiscountPercentage = input.DiscountPercentage

	if err := s.db.Save(&tier).Error; err != nil {
		return nil, err
	}
	return &tier, nil
}

func (s *CRMService) DeleteCustomerTier(id string) error {
	tierID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid tier ID")
	}

	if err := s.db.Where("id = ?", tierID).Delete(&models.CustomerTier{}).Error; err != nil {
		return err
	}
	return nil
}

type CustomerCreditData struct {
	Transactions    []models.CustomerCreditTransaction
	OutstandingDebt float64
}

func (s *CRMService) ListCustomerCreditTransactions(customerID string) (*CustomerCreditData, error) {
	var transactions []models.CustomerCreditTransaction
	if err := s.db.Where("customer_id = ?", customerID).Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}

	var customer models.Customer
	if err := s.db.Where("id = ?", customerID).First(&customer).Error; err != nil {
		return nil, err
	}

	return &CustomerCreditData{
		Transactions:    transactions,
		OutstandingDebt: customer.DebtBalance,
	}, nil
}

type PaymentInput struct {
	Amount    float64
	Reference string
	Notes     string
}

func (s *CRMService) RecordCustomerPayment(customerIDStr string, input PaymentInput) (*models.Customer, *models.CustomerCreditTransaction, error) {
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, nil, errors.New("invalid customer_id")
	}

	var customer models.Customer
	if err := s.db.Where("id = ?", customerID).First(&customer).Error; err != nil {
		return nil, nil, errors.New("customer not found")
	}

	newDebt := customer.DebtBalance - input.Amount
	if newDebt < 0 {
		newDebt = 0
	}
	customer.DebtBalance = newDebt

	if err := s.db.Save(&customer).Error; err != nil {
		return nil, nil, err
	}

	tx := models.CustomerCreditTransaction{
		CustomerID:      customerID,
		Amount:          -input.Amount,
		TransactionType: "payment",
		Reference:       input.Reference,
		Notes:           input.Notes,
	}
	if err := s.db.Create(&tx).Error; err != nil {
		return nil, nil, err
	}

	return &customer, &tx, nil
}

func (s *CRMService) GetFeedbackList() ([]models.CustomerFeedback, error) {
	var feedback []models.CustomerFeedback
	if err := s.db.Order("created_at DESC").Find(&feedback).Error; err != nil {
		return nil, err
	}
	return feedback, nil
}
