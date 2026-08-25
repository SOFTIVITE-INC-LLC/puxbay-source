package services

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type B2BService struct {
	db *gorm.DB
}

func NewB2BService(db *gorm.DB) *B2BService {
	return &B2BService{db: db}
}

func (s *B2BService) ListQuotes(tenantID uuid.UUID, role string) ([]models.Quotation, error) {
	var quotations []models.Quotation

	query := s.db.Preload("Customer")

	roleLower := strings.ToLower(role)
	if roleLower != "admin" && roleLower != "manager" && roleLower != "sales" && roleLower != "superadmin" {
		return nil, errors.New("unauthorized to view all quotes")
	}

	if err := query.Find(&quotations).Error; err != nil {
		return nil, err
	}
	return quotations, nil
}

type QuoteItemInput struct {
	ProductID string
	Quantity  int
}

type QuoteCreateInput struct {
	CustomerID string
	Items      []QuoteItemInput
	ProfileID  uuid.UUID
}

func (s *B2BService) CreateQuote(input QuoteCreateInput) (uuid.UUID, error) {
	customerUUID, err := uuid.Parse(input.CustomerID)
	if err != nil {
		return uuid.Nil, errors.New("invalid customer ID")
	}

	quote := models.Quotation{
		CustomerID:  customerUUID,
		Status:      "pending",
		CreatedByID: &input.ProfileID,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		if err := tx.Preload("Tier").First(&customer, "id = ?", quote.CustomerID).Error; err != nil {
			return err
		}

		if err := tx.Create(&quote).Error; err != nil {
			return err
		}

		var subtotal float64 = 0

		for _, itemReq := range input.Items {
			var product models.Product
			if err := tx.First(&product, "id = ?", itemReq.ProductID).Error; err != nil {
				return err
			}

			unitPrice := product.SellingPrice
			if product.WholesalePrice > 0 {
				unitPrice = product.WholesalePrice
			}

			// Apply Tier Discount Override
			if customer.Tier != nil && customer.Tier.DiscountPercentage > 0 {
				unitPrice = unitPrice * (1.0 - (customer.Tier.DiscountPercentage / 100.0))
			}

			qItem := models.QuotationItem{
				QuotationID: quote.ID,
				ProductID:   product.ID,
				Quantity:    uint(itemReq.Quantity),
				UnitPrice:   unitPrice,
				TotalPrice:  unitPrice * float64(itemReq.Quantity),
			}
			if err := tx.Create(&qItem).Error; err != nil {
				return err
			}
			subtotal += qItem.TotalPrice
		}

		quote.Subtotal = subtotal
		quote.TotalAmount = subtotal
		if err := tx.Save(&quote).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return quote.ID, nil
}

type QuoteUpdateInput struct {
	Action        string
	InternalNotes string
}

func (s *B2BService) UpdateQuote(id string, input QuoteUpdateInput) (string, error) {
	var quote models.Quotation
	if err := s.db.First(&quote, "id = ?", id).Error; err != nil {
		return "", errors.New("quote not found")
	}

	if input.Action == "approve" {
		quote.Status = "approved"
		quote.InternalNotes = &input.InternalNotes
	} else if input.Action == "reject" {
		quote.Status = "rejected"
	} else {
		return "", errors.New("invalid action")
	}

	if err := s.db.Save(&quote).Error; err != nil {
		return "", err
	}

	return quote.Status, nil
}

func (s *B2BService) ListClients() ([]models.Customer, error) {
	var customers []models.Customer
	if err := s.db.Where("customer_type = ?", "wholesale").Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}
