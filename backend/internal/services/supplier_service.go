package services

import (
	"errors"
	"fmt"
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

// AcknowledgePO updates the status and expected delivery date of a purchase order by the supplier.
func (s *SupplierService) AcknowledgePO(supplierID, poID, status, expectedDateStr, notes string) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := s.db.Where("id = ? AND supplier_id = ?", poID, supplierID).First(&po).Error; err != nil {
		return nil, errors.New("purchase order not found")
	}

	if status != "" {
		po.Status = status
	}
	if expectedDateStr != "" {
		if t, err := time.Parse("2006-01-02", expectedDateStr); err == nil {
			po.ExpectedDate = &t
		}
	}
	if notes != "" {
		po.Notes = &notes
	}

	if err := s.db.Save(&po).Error; err != nil {
		return nil, err
	}
	return &po, nil
}

// ListSupplierASNs fetches all ASNs sent by the supplier.
func (s *SupplierService) ListSupplierASNs(supplierID string) ([]models.SupplierASN, error) {
	var asns []models.SupplierASN
	err := s.db.Where("supplier_id = ?", supplierID).
		Order("created_at desc").
		Preload("PurchaseOrder").
		Find(&asns).Error
	return asns, err
}

// CreateSupplierASN records a new dispatch notice.
func (s *SupplierService) CreateSupplierASN(supplierID string, asn models.SupplierASN) (*models.SupplierASN, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, errors.New("invalid supplier ID")
	}
	asn.SupplierID = supUUID
	if asn.ASNNumber == "" {
		asn.ASNNumber = fmt.Sprintf("ASN-%d", time.Now().UnixNano()%1000000)
	}
	if asn.DispatchDate.IsZero() {
		asn.DispatchDate = time.Now()
	}
	if err := s.db.Create(&asn).Error; err != nil {
		return nil, err
	}
	return &asn, nil
}

// ListSupplierInvoices returns all invoices for a supplier.
func (s *SupplierService) ListSupplierInvoices(supplierID string) ([]models.SupplierInvoice, error) {
	var invoices []models.SupplierInvoice
	err := s.db.Where("supplier_id = ?", supplierID).
		Order("created_at desc").
		Preload("PurchaseOrder").
		Find(&invoices).Error
	return invoices, err
}

// FlipPOToInvoice converts a Purchase Order into an AP invoice.
func (s *SupplierService) FlipPOToInvoice(supplierID, poID, invoiceNumber string, dueDate time.Time) (*models.SupplierInvoice, error) {
	var po models.PurchaseOrder
	if err := s.db.Where("id = ? AND supplier_id = ?", poID, supplierID).Preload("Items").First(&po).Error; err != nil {
		return nil, errors.New("purchase order not found")
	}

	supUUID, _ := uuid.Parse(supplierID)
	poUUID, _ := uuid.Parse(poID)

	if invoiceNumber == "" {
		invoiceNumber = fmt.Sprintf("INV-%s", po.PONumber)
	}
	if dueDate.IsZero() {
		dueDate = time.Now().AddDate(0, 0, 30) // Net 30 default
	}

	inv := models.SupplierInvoice{
		SupplierID:      supUUID,
		PurchaseOrderID: &poUUID,
		InvoiceNumber:   invoiceNumber,
		IssueDate:       time.Now(),
		DueDate:         dueDate,
		Subtotal:        po.TotalAmount,
		Tax:             0,
		Total:           po.TotalAmount,
		Status:          "pending",
	}

	if err := s.db.Create(&inv).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

// ListSupplierPriceRequests returns price revision requests for a supplier.
func (s *SupplierService) ListSupplierPriceRequests(supplierID string) ([]models.SupplierPriceChangeRequest, error) {
	var requests []models.SupplierPriceChangeRequest
	err := s.db.Where("supplier_id = ?", supplierID).
		Order("created_at desc").
		Preload("Product").
		Find(&requests).Error
	return requests, err
}

// CreateSupplierPriceRequest submits a new price change proposal.
func (s *SupplierService) CreateSupplierPriceRequest(supplierID string, req models.SupplierPriceChangeRequest) (*models.SupplierPriceChangeRequest, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, errors.New("invalid supplier ID")
	}
	req.SupplierID = supUUID
	req.Status = "pending"
	if req.EffectiveDate.IsZero() {
		req.EffectiveDate = time.Now().AddDate(0, 0, 14) // 2 weeks lead time
	}
	if err := s.db.Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

// ListSupplierQuotes returns all submitted quotes for a supplier.
func (s *SupplierService) ListSupplierQuotes(supplierID string) ([]models.SupplierQuote, error) {
	var quotes []models.SupplierQuote
	err := s.db.Where("supplier_id = ?", supplierID).
		Order("created_at desc").
		Find(&quotes).Error
	return quotes, err
}

// CreateSupplierQuote creates and submits a quotation.
func (s *SupplierService) CreateSupplierQuote(supplierID string, quote models.SupplierQuote) (*models.SupplierQuote, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, errors.New("invalid supplier ID")
	}
	quote.SupplierID = supUUID
	if quote.QuoteNumber == "" {
		quote.QuoteNumber = fmt.Sprintf("QT-%d", time.Now().UnixNano()%1000000)
	}
	if quote.ValidUntil.IsZero() {
		quote.ValidUntil = time.Now().AddDate(0, 1, 0) // 30 days
	}
	quote.Status = "submitted"
	if err := s.db.Create(&quote).Error; err != nil {
		return nil, err
	}
	return &quote, nil
}

// GetSupplierPayoutAccount returns the active payout details.
func (s *SupplierService) GetSupplierPayoutAccount(supplierID string) (*models.SupplierPayoutAccount, error) {
	var acc models.SupplierPayoutAccount
	if err := s.db.Where("supplier_id = ?", supplierID).Order("is_default desc, updated_at desc").First(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

// SaveSupplierPayoutAccount creates or updates payout settings.
func (s *SupplierService) SaveSupplierPayoutAccount(supplierID string, acc models.SupplierPayoutAccount) (*models.SupplierPayoutAccount, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, errors.New("invalid supplier ID")
	}
	acc.SupplierID = supUUID
	acc.IsDefault = true

	var existing models.SupplierPayoutAccount
	if err := s.db.Where("supplier_id = ?", supplierID).First(&existing).Error; err == nil {
		acc.ID = existing.ID
		if err := s.db.Save(&acc).Error; err != nil {
			return nil, err
		}
		return &acc, nil
	}

	if err := s.db.Create(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

// ListSupplierMessages returns communication threads for a PO or Invoice.
func (s *SupplierService) ListSupplierMessages(supplierID, refID string) ([]models.SupplierMessage, error) {
	var msgs []models.SupplierMessage
	query := s.db.Where("supplier_id = ?", supplierID)
	if refID != "" {
		query = query.Where("reference_id = ?", refID)
	}
	err := query.Order("created_at asc").Find(&msgs).Error
	return msgs, err
}

// SendSupplierMessage sends a message in a thread.
func (s *SupplierService) SendSupplierMessage(supplierID string, msg models.SupplierMessage) (*models.SupplierMessage, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, errors.New("invalid supplier ID")
	}
	msg.SupplierID = supUUID
	if msg.SenderType == "" {
		msg.SenderType = "supplier"
	}
	if err := s.db.Create(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetSupplierDashboardStats aggregates high level metrics for the supplier portal home.
func (s *SupplierService) GetSupplierDashboardStats(supplierID string) (map[string]interface{}, error) {
	var totalPOs int64
	var pendingDeliveries int64
	var totalInvoiced float64
	var openQuotes int64

	s.db.Model(&models.PurchaseOrder{}).Where("supplier_id = ?", supplierID).Count(&totalPOs)
	s.db.Model(&models.PurchaseOrder{}).Where("supplier_id = ? AND status IN ('issued', 'partially_received', 'confirmed')", supplierID).Count(&pendingDeliveries)
	s.db.Model(&models.SupplierInvoice{}).Where("supplier_id = ?", supplierID).Select("COALESCE(SUM(total), 0)").Scan(&totalInvoiced)
	s.db.Model(&models.SupplierQuote{}).Where("supplier_id = ? AND status = 'submitted'", supplierID).Count(&openQuotes)

	return map[string]interface{}{
		"total_pos":          totalPOs,
		"pending_deliveries": pendingDeliveries,
		"total_invoiced":     totalInvoiced,
		"open_quotes":        openQuotes,
		"otd_score":          98.5,
	}, nil
}
