package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SupplierService struct {
	db *gorm.DB
}

func NewSupplierService(db *gorm.DB) *SupplierService {
	return &SupplierService{db: db}
}

func (s *SupplierService) ListSuppliers() ([]models.Supplier, error) {
	var suppliers []models.Supplier
	if err := s.db.Find(&suppliers).Error; err != nil {
		return nil, err
	}
	return suppliers, nil
}

func (s *SupplierService) GetSupplier(id string) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := s.db.First(&supplier, "id = ?", id).Error; err != nil {
		return nil, errors.New("supplier not found")
	}
	return &supplier, nil
}

type SupplierInput struct {
	Name          string
	ContactPerson *string
	Email         *string
	Phone         *string
	Address       *string
	TaxNumber     *string
	PaymentTerms  *string
	Notes         *string
}

func (s *SupplierService) CreateSupplier(input SupplierInput) (*models.Supplier, error) {
	supplier := models.Supplier{
		Name:          input.Name,
		ContactPerson: input.ContactPerson,
		Email:         input.Email,
		Phone:         input.Phone,
		Address:       input.Address,
		TaxNumber:     input.TaxNumber,
		PaymentTerms:  input.PaymentTerms,
		Notes:         input.Notes,
		IsActive:      true,
	}

	if err := s.db.Create(&supplier).Error; err != nil {
		return nil, err
	}

	return &supplier, nil
}

func (s *SupplierService) UpdateSupplier(id string, input SupplierInput) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := s.db.First(&supplier, "id = ?", id).Error; err != nil {
		return nil, errors.New("supplier not found")
	}

	supplier.Name = input.Name
	supplier.ContactPerson = input.ContactPerson
	supplier.Email = input.Email
	supplier.Phone = input.Phone
	supplier.Address = input.Address
	supplier.TaxNumber = input.TaxNumber
	supplier.PaymentTerms = input.PaymentTerms
	supplier.Notes = input.Notes

	if err := s.db.Save(&supplier).Error; err != nil {
		return nil, err
	}

	return &supplier, nil
}

func (s *SupplierService) DeleteSupplier(id string) error {
	if err := s.db.Delete(&models.Supplier{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (s *SupplierService) ListSupplierProducts(supplierID string) ([]models.SupplierProduct, error) {
	var products []models.SupplierProduct
	if err := s.db.Preload("Product").Where("supplier_id = ?", supplierID).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

type SupplierProductInput struct {
	ProductID   string
	SupplierSKU string
	UnitCost    float64
	MinOrderQty float64
}

func (s *SupplierService) AddSupplierProduct(supplierID string, input SupplierProductInput) (*models.SupplierProduct, error) {
	minQty := input.MinOrderQty
	if minQty <= 0 {
		minQty = 1
	}
	supplierProduct := models.SupplierProduct{
		SupplierID:  uuid.MustParse(supplierID),
		ProductID:   uuid.MustParse(input.ProductID),
		SupplierSKU: input.SupplierSKU,
		UnitCost:    input.UnitCost,
		MinOrderQty: minQty,
	}

	if err := s.db.Create(&supplierProduct).Error; err != nil {
		return nil, err
	}

	return &supplierProduct, nil
}

func (s *SupplierService) ListLedgerEntries(supplierID string) ([]models.SupplierLedgerEntry, error) {
	var entries []models.SupplierLedgerEntry
	if err := s.db.Where("supplier_id = ?", supplierID).Order("created_at desc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

type LedgerEntryInput struct {
	EntryType   string
	Amount      float64
	ReferenceID *string
	Notes       *string
}

func (s *SupplierService) AddLedgerEntry(supplierID string, input LedgerEntryInput) (*models.SupplierLedgerEntry, error) {
	var entry *models.SupplierLedgerEntry

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var supplier models.Supplier
		if err := tx.Where("id = ?", supplierID).First(&supplier).Error; err != nil {
			return err
		}

		if input.EntryType == "payment" {
			supplier.CreditBalance -= input.Amount
		} else if input.EntryType == "invoice" {
			supplier.CreditBalance += input.Amount
		}

		if err := tx.Save(&supplier).Error; err != nil {
			return err
		}

		entry = &models.SupplierLedgerEntry{
			SupplierID:  supplier.ID,
			EntryType:   input.EntryType,
			Amount:      input.Amount,
			Balance:     supplier.CreditBalance,
			ReferenceID: input.ReferenceID,
			Notes:       input.Notes,
		}

		if err := tx.Create(entry).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return entry, nil
}

// LedgerEntryInput with TransactionDate support
// (already defined above — we extend it here for the date field via the existing struct)
// NOTE: TransactionDate is accepted via AddLedgerEntry below

// RemoveSupplierProduct removes a product link from a supplier's catalog.
func (s *SupplierService) RemoveSupplierProduct(supplierID, productID string) error {
	return s.db.Where("supplier_id = ? AND product_id = ?", supplierID, productID).
		Delete(&models.SupplierProduct{}).Error
}

// InvitePortalAccess sets portal login credentials on a supplier record.
func (s *SupplierService) InvitePortalAccess(supplierID, email, password string) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := s.db.First(&supplier, "id = ?", supplierID).Error; err != nil {
		return nil, errors.New("supplier not found")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashStr := string(hash)
	supplier.PortalEmail = &email
	supplier.PortalPassword = &hashStr
	if err := s.db.Save(&supplier).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

// AuthenticateSupplierPortal validates portal credentials and returns the supplier.
// tenantID is determined from the subdomain (already scoped in s.db).
func (s *SupplierService) AuthenticateSupplierPortal(email, password string) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := s.db.Where("portal_email = ?", email).First(&supplier).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}
	if supplier.PortalPassword == nil {
		return nil, errors.New("portal access not configured")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*supplier.PortalPassword), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &supplier, nil
}

// ListSupplierPurchaseOrders lists all purchase orders addressed to a supplier.
func (s *SupplierService) ListSupplierPurchaseOrders(supplierID string) ([]models.PurchaseOrder, error) {
	var orders []models.PurchaseOrder
	if err := s.db.
		Where("supplier_id = ?", supplierID).
		Order("created_at desc").
		Preload("Items").
		Preload("Items.Product").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// UpdateLedgerEntryInput extends the base with TransactionDate.
type UpdatedLedgerEntryInput struct {
	EntryType       string
	Amount          float64
	ReferenceID     *string
	Notes           *string
	TransactionDate *time.Time
}

// AddLedgerEntryFull records a ledger entry including an optional transaction date.
func (s *SupplierService) AddLedgerEntryFull(supplierID string, input UpdatedLedgerEntryInput) (*models.SupplierLedgerEntry, error) {
	var entry *models.SupplierLedgerEntry

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var supplier models.Supplier
		if err := tx.Where("id = ?", supplierID).First(&supplier).Error; err != nil {
			return err
		}

		if input.EntryType == "payment" {
			supplier.CreditBalance -= input.Amount
		} else if input.EntryType == "invoice" {
			supplier.CreditBalance += input.Amount
		}

		if err := tx.Save(&supplier).Error; err != nil {
			return err
		}

		entry = &models.SupplierLedgerEntry{
			SupplierID:      supplier.ID,
			EntryType:       input.EntryType,
			Amount:          input.Amount,
			Balance:         supplier.CreditBalance,
			ReferenceID:     input.ReferenceID,
			Notes:           input.Notes,
			TransactionDate: input.TransactionDate,
		}

		return tx.Create(entry).Error
	})

	return entry, err
}
