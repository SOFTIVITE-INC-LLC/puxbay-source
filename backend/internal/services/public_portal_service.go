package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type PublicPortalService struct {
	db *gorm.DB
}

func NewPublicPortalService(db *gorm.DB) *PublicPortalService {
	return &PublicPortalService{db: db}
}

type TenantPublicInfo struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	LogoURL      string `json:"logo_url"`
	ContactEmail string `json:"contact_email"`
}

func (s *PublicPortalService) GetTenantInfo(domain string) (*TenantPublicInfo, error) {
	// This is a simplified mock. In a real scenario, you'd lookup the tenant by domain.
	return &TenantPublicInfo{
		Name:         "Demo Store",
		Description:  "Welcome to our store",
		LogoURL:      "https://example.com/logo.png",
		ContactEmail: "hello@example.com",
	}, nil
}

func (s *PublicPortalService) ListProducts(tenantID string) ([]models.Product, error) {
	var products []models.Product
	if err := s.db.Where("is_active = ?", true).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

type TrackOrderResult struct {
	OrderID   uuid.UUID `json:"order_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Total     float64   `json:"total"`
}

func (s *PublicPortalService) TrackOrder(tenantID, orderID string) (*TrackOrderResult, error) {
	var order models.Order
	if err := s.db.Where("id = ?", orderID).First(&order).Error; err != nil {
		return nil, errors.New("order not found")
	}

	return &TrackOrderResult{
		OrderID:   order.ID,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
		Total:     order.Total,
	}, nil
}

type PublicFeedbackInput struct {
	Name    string
	Email   string
	Rating  uint
	Comment string
}

func (s *PublicPortalService) SubmitFeedback(tenantID string, input PublicFeedbackInput) (*models.CustomerFeedback, error) {
	if input.Email == "" {
		return nil, errors.New("email is required")
	}

	var customer models.Customer
	// Attempt to find customer by email
	err := s.db.Where("email = ?", input.Email).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new customer
			customer = models.Customer{
				Name:  input.Name,
				Email: &input.Email,
			}
			if customer.Name == "" {
				customer.Name = "Guest User"
			}
			if err := s.db.Create(&customer).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	feedback := models.CustomerFeedback{
		CustomerID: customer.ID,
		Rating:     input.Rating,
		IsPublic:   true,
	}

	if input.Comment != "" {
		feedback.Comment = &input.Comment
	}

	if err := s.db.Create(&feedback).Error; err != nil {
		return nil, err
	}

	return &feedback, nil
}
