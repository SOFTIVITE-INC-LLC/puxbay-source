package services

import (
	"errors"

	customerrors "github.com/softivite/puxbay/internal/errors"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CustomerService struct {
	db *gorm.DB
}

func NewCustomerService(db *gorm.DB) *CustomerService {
	return &CustomerService{db: db}
}

func (s *CustomerService) ListCustomers(search string, limit, offset int) ([]models.Customer, int64, error) {
	var customers []models.Customer
	var total int64

	query := s.db.Model(&models.Customer{})

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&customers).Error; err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

type CustomerInput struct {
	Name    string
	Email   string
	Phone   string
	Address string
}

func (s *CustomerService) CreateCustomer(input CustomerInput) (*models.Customer, error) {
	customer := models.Customer{
		Name: input.Name,
	}

	if input.Email != "" {
		customer.Email = &input.Email
	}
	if input.Phone != "" {
		customer.Phone = &input.Phone
	}
	if input.Address != "" {
		customer.Address = &input.Address
	}

	if err := s.db.Create(&customer).Error; err != nil {
		return nil, err
	}

	return &customer, nil
}

// FindOrCreateCustomer looks up a customer by phone number, creating one if not found.
// If phone is empty, always creates a new customer.
func (s *CustomerService) FindOrCreateCustomer(input CustomerInput) (*models.Customer, error) {
	if input.Phone != "" {
		var existing models.Customer
		err := s.db.Where("phone = ?", input.Phone).First(&existing).Error
		if err == nil {
			// Update name if it changed
			if input.Name != "" && existing.Name != input.Name {
				s.db.Model(&existing).Update("name", input.Name)
				existing.Name = input.Name
			}
			return &existing, nil
		}
	}
	return s.CreateCustomer(input)
}

func (s *CustomerService) GetCustomer(id string) (*models.Customer, error) {
	var customer models.Customer
	if err := s.db.Where("id = ?", id).First(&customer).Error; err != nil {
		return nil, customerrors.ErrNotFound
	}
	return &customer, nil
}

func (s *CustomerService) UpdateCustomer(id string, input CustomerInput) (*models.Customer, error) {
	var customer models.Customer
	if err := s.db.Where("id = ?", id).First(&customer).Error; err != nil {
		return nil, customerrors.ErrNotFound
	}

	customer.Name = input.Name

	if input.Email != "" {
		customer.Email = &input.Email
	} else {
		customer.Email = nil
	}

	if input.Phone != "" {
		customer.Phone = &input.Phone
	} else {
		customer.Phone = nil
	}

	if input.Address != "" {
		customer.Address = &input.Address
	} else {
		customer.Address = nil
	}

	if err := s.db.Save(&customer).Error; err != nil {
		return nil, err
	}

	return &customer, nil
}

func (s *CustomerService) DeleteCustomer(id string) error {
	if err := s.db.Where("id = ?", id).Delete(&models.Customer{}).Error; err != nil {
		return err
	}
	return nil
}
func (s *CustomerService) RecordPayment(id string, amount float64, paymentMethod string, notes string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&customer).Error; err != nil {
			return customerrors.ErrNotFound
		}

		if amount <= 0 {
			return errors.New("amount must be positive")
		}

		remaining := amount
		if customer.DebtBalance > 0 {
			if customer.DebtBalance >= remaining {
				customer.DebtBalance -= remaining
				remaining = 0
			} else {
				remaining -= customer.DebtBalance
				customer.DebtBalance = 0
			}
		}

		// Any remaining amount becomes store credit
		if remaining > 0 {
			customer.StoreCredit += remaining

			// Log store credit transaction
			creditTx := models.StoreCreditTransaction{
				CustomerID:      customer.ID,
				Amount:          remaining,
				TransactionType: "manual",
				Notes:           "Overpayment converted to store credit",
			}
			if err := tx.Create(&creditTx).Error; err != nil {
				return err
			}
		}

		return tx.Save(&customer).Error
	})
}
