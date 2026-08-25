package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type WalletService struct {
	db *gorm.DB
}

func NewWalletService(db *gorm.DB) *WalletService {
	return &WalletService{db: db}
}

type WalletDashboardData struct {
	Customer       models.Customer
	RecentOrders   []models.Order
	GiftCards      []models.GiftCard
	LoyaltyHistory []models.LoyaltyTransaction
}

func (s *WalletService) GetDashboard(customerIDStr, customerPhone string) (*WalletDashboardData, error) {
	var customer models.Customer
	var err error

	if customerIDStr != "" {
		customerID, parseErr := uuid.Parse(customerIDStr)
		if parseErr != nil {
			return nil, errors.New("invalid customer_id")
		}
		err = s.db.Where("id = ?", customerID).First(&customer).Error
	} else if customerPhone != "" {
		err = s.db.Where("phone = ?", customerPhone).First(&customer).Error
	} else {
		return nil, errors.New("phone or customer_id required")
	}

	if err != nil {
		return nil, errors.New("customer not found")
	}

	var recentOrders []models.Order
	s.db.Where("customer_id = ?", customer.ID).Order("created_at DESC").Limit(5).Find(&recentOrders)

	var giftCards []models.GiftCard
	s.db.Where("purchaser_id = ? AND is_active = ?", customer.ID, true).Find(&giftCards)

	var loyaltyHistory []models.LoyaltyTransaction
	s.db.Where("customer_id = ?", customer.ID).Order("created_at DESC").Limit(10).Find(&loyaltyHistory)

	return &WalletDashboardData{
		Customer:       customer,
		RecentOrders:   recentOrders,
		GiftCards:      giftCards,
		LoyaltyHistory: loyaltyHistory,
	}, nil
}

func (s *WalletService) LookupCustomer(phone string) (*models.Customer, error) {
	var customer models.Customer
	if err := s.db.Where("phone = ?", phone).First(&customer).Error; err != nil {
		return nil, errors.New("customer not found with this phone number")
	}
	return &customer, nil
}

func (s *WalletService) GetGiftCards(customerIDStr string) ([]models.GiftCard, error) {
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, errors.New("invalid customer_id")
	}

	var giftCards []models.GiftCard
	s.db.Where("purchaser_id = ?", customerID).Find(&giftCards)
	return giftCards, nil
}

func (s *WalletService) CreateGiftCard(giftCard *models.GiftCard) error {
	giftCard.IsActive = true
	return s.db.Create(giftCard).Error
}

func (s *WalletService) GetLoyaltyTransactions(customerIDStr string) ([]models.LoyaltyTransaction, error) {
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, errors.New("invalid customer_id")
	}

	var transactions []models.LoyaltyTransaction
	s.db.Where("customer_id = ?", customerID).Order("created_at DESC").Limit(20).Find(&transactions)
	return transactions, nil
}

type LoyaltyAdjustmentInput struct {
	Points      float64
	Description string
}

func (s *WalletService) AdjustLoyaltyPoints(customerIDStr string, input LoyaltyAdjustmentInput) (*models.Customer, *models.LoyaltyTransaction, error) {
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, nil, errors.New("invalid customer_id")
	}

	var customer models.Customer
	if err := s.db.Where("id = ?", customerID).First(&customer).Error; err != nil {
		return nil, nil, errors.New("customer not found")
	}

	customer.LoyaltyPts += input.Points
	s.db.Save(&customer)

	txType := "adjustment"
	transaction := models.LoyaltyTransaction{
		CustomerID:      customerID,
		Points:          input.Points,
		TransactionType: txType,
		Description:     &input.Description,
	}
	s.db.Create(&transaction)

	return &customer, &transaction, nil
}

func (s *WalletService) GetBalance(customerIDStr string) (*models.Customer, error) {
	if customerIDStr == "" {
		return nil, errors.New("customer_id required")
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, errors.New("invalid customer_id")
	}

	var customer models.Customer
	if err := s.db.Where("id = ?", customerID).First(&customer).Error; err != nil {
		return nil, errors.New("customer not found")
	}

	return &customer, nil
}

type StoreCreditAdjustmentInput struct {
	Amount float64
	Note   string
}

func (s *WalletService) AdjustStoreCredit(customerIDStr string, input StoreCreditAdjustmentInput) (*models.Customer, error) {
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, errors.New("invalid customer_id")
	}

	var customer models.Customer
	if err := s.db.Where("id = ?", customerID).First(&customer).Error; err != nil {
		return nil, errors.New("customer not found")
	}

	customer.StoreCredit += input.Amount
	if err := s.db.Save(&customer).Error; err != nil {
		return nil, err
	}

	txType := "adjustment"
	if input.Amount > 0 {
		txType = "credit"
	} else {
		txType = "debit"
	}

	creditTx := models.StoreCreditTransaction{
		CustomerID:      customerID,
		Amount:          input.Amount,
		TransactionType: txType,
		Notes:           input.Note,
	}
	s.db.Create(&creditTx)

	return &customer, nil
}
