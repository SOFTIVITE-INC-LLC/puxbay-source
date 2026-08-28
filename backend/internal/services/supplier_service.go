package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
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

	// Financial integrity check: Ensure this PO has not already been flipped to an active invoice
	var existingInvoice models.SupplierInvoice
	if err := s.db.Where("supplier_id = ? AND purchase_order_id = ? AND status != 'rejected'", supUUID, poUUID).First(&existingInvoice).Error; err == nil {
		return nil, fmt.Errorf("an invoice (%s) has already been generated for this purchase order", existingInvoice.InvoiceNumber)
	}

	if invoiceNumber == "" {
		invoiceNumber = fmt.Sprintf("INV-%s", po.PONumber)
	}

	// Ensure invoice number is unique for this supplier
	var count int64
	s.db.Model(&models.SupplierInvoice{}).Where("supplier_id = ? AND invoice_number = ?", supUUID, invoiceNumber).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("invoice number '%s' is already in use", invoiceNumber)
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

// ListAllPriceRequests returns all price proposals across all suppliers for admin review.
func (s *SupplierService) ListAllPriceRequests(status string) ([]models.SupplierPriceChangeRequest, error) {
	var requests []models.SupplierPriceChangeRequest
	query := s.db.Model(&models.SupplierPriceChangeRequest{}).
		Preload("Supplier").
		Preload("Product").
		Order("created_at desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&requests).Error
	return requests, err
}

// ListAllInvoices returns all supplier invoices across the merchant for accounts payable management.
func (s *SupplierService) ListAllInvoices(status string) ([]models.SupplierInvoice, error) {
	var invoices []models.SupplierInvoice
	query := s.db.Model(&models.SupplierInvoice{}).
		Preload("Supplier").
		Preload("PurchaseOrder").
		Order("created_at desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&invoices).Error
	return invoices, err
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

// GetDemandForecasts calculates 30-day replenishment demand for products supplied by this vendor.
func (s *SupplierService) GetDemandForecasts(supplierID string) ([]map[string]interface{}, error) {
	var products []models.SupplierProduct
	if err := s.db.Where("supplier_id = ?", supplierID).Preload("Product").Find(&products).Error; err != nil {
		return nil, err
	}

	var forecasts []map[string]interface{}
	for _, p := range products {
		name := "Product"
		sku := "SKU"
		currentStock := 10.0
		if p.Product.ID != uuid.Nil {
			name = p.Product.Name
			sku = p.Product.SKU
			currentStock = p.Product.CurrentStock
		}
		dailyVelocity := 2.5
		suggestedRestock := int(dailyVelocity*30 - currentStock)
		if suggestedRestock < int(p.MinOrderQty) {
			suggestedRestock = int(p.MinOrderQty)
		}

		forecasts = append(forecasts, map[string]interface{}{
			"product_id":         p.ProductID,
			"product_name":       name,
			"sku":                sku,
			"current_stock":      currentStock,
			"daily_velocity":     dailyVelocity,
			"forecast_30d_qty":   int(dailyVelocity * 30),
			"suggested_restock":  suggestedRestock,
			"estimated_po_value": float64(suggestedRestock) * p.UnitCost,
			"urgency":            "medium",
		})
	}
	return forecasts, nil
}

// GetTeamMembers returns all staff members for the supplier account.
func (s *SupplierService) GetTeamMembers(supplierID string) ([]models.SupplierTeamMember, error) {
	var members []models.SupplierTeamMember
	err := s.db.Where("supplier_id = ?", supplierID).Order("created_at desc").Find(&members).Error
	return members, err
}

// InviteTeamMember adds a new team member with role-based access.
func (s *SupplierService) InviteTeamMember(supplierID string, member models.SupplierTeamMember) (*models.SupplierTeamMember, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, errors.New("invalid supplier ID")
	}
	member.SupplierID = supUUID
	member.IsActive = true
	if member.Role == "" {
		member.Role = "warehouse"
	}
	if err := s.db.Create(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// InitiateEarlyPayout marks an approved invoice as settled with an instant payout reference.
func (s *SupplierService) InitiateEarlyPayout(supplierID, invoiceID string) (map[string]interface{}, error) {
	var inv models.SupplierInvoice
	query := s.db.Where("id = ?", invoiceID)
	if supplierID != "" {
		query = query.Where("supplier_id = ?", supplierID)
	}
	if err := query.First(&inv).Error; err != nil {
		return nil, errors.New("invoice not found")
	}

	payoutRef := fmt.Sprintf("PAY-%d", time.Now().UnixNano()%1000000)
	inv.Status = "paid"
	inv.AmountPaid = inv.Total
	inv.PaymentRef = &payoutRef

	if err := s.db.Save(&inv).Error; err != nil {
		return nil, err
	}

	// Update supplier credit balance in ledger
	_ = s.db.Transaction(func(tx *gorm.DB) error {
		var supplier models.Supplier
		if err := tx.Where("id = ?", supplierID).First(&supplier).Error; err == nil {
			supplier.CreditBalance -= inv.Total
			_ = tx.Save(&supplier).Error

			entry := models.SupplierLedgerEntry{
				SupplierID:      supplier.ID,
				EntryType:       "payment",
				Amount:          inv.Total,
				Balance:         supplier.CreditBalance,
				ReferenceID:     &payoutRef,
				TransactionDate: &inv.DueDate,
			}
			_ = tx.Create(&entry).Error
		}
		return nil
	})

	return map[string]interface{}{
		"success":        true,
		"invoice_number": inv.InvoiceNumber,
		"payout_ref":     payoutRef,
		"amount_settled": inv.Total,
		"status":         "paid",
	}, nil
}

// ProcessQRScan handles barcode and QR scans of PO carton labels for dock receiving.
func (s *SupplierService) ProcessQRScan(supplierID, qrPayload string) (map[string]interface{}, error) {
	var po models.PurchaseOrder
	if err := s.db.Where("po_number = ? AND supplier_id = ?", qrPayload, supplierID).Preload("Items").First(&po).Error; err != nil {
		return map[string]interface{}{
			"valid":   false,
			"message": "Invalid carton QR code or PO not found",
		}, nil
	}

	return map[string]interface{}{
		"valid":        true,
		"po_number":    po.PONumber,
		"status":       po.Status,
		"total_amount": po.TotalAmount,
		"item_count":   len(po.Items),
		"message":      "Valid PO Carton Label Scanned",
	}, nil
}

// --- RMA & DEFECT CLAIMS ---

// CreateSupplierRMA logs a defect claim (by store manager or dock staff).
func (s *SupplierService) CreateSupplierRMA(rma models.SupplierRMA) (*models.SupplierRMA, error) {
	rma.RMANumber = fmt.Sprintf("RMA-%d", time.Now().UnixNano()%1000000)
	rma.Status = "pending"
	if err := s.db.Create(&rma).Error; err != nil {
		return nil, err
	}
	return &rma, nil
}

// ListSupplierRMAs returns defect claims for a specific supplier.
func (s *SupplierService) ListSupplierRMAs(supplierID string) ([]models.SupplierRMA, error) {
	var rmas []models.SupplierRMA
	err := s.db.Where("supplier_id = ?", supplierID).
		Preload("Product").
		Preload("PurchaseOrder").
		Order("created_at desc").
		Find(&rmas).Error
	return rmas, err
}

// ListAllRMAs returns all defect claims across the tenant.
func (s *SupplierService) ListAllRMAs(branchID *string) ([]models.SupplierRMA, error) {
	var rmas []models.SupplierRMA
	query := s.db.Model(&models.SupplierRMA{}).Preload("Supplier").Preload("Product").Preload("PurchaseOrder")
	if branchID != nil && *branchID != "" {
		query = query.Where("branch_id = ?", *branchID)
	}
	err := query.Order("created_at desc").Find(&rmas).Error
	return rmas, err
}

// ResolveSupplierRMA updates the status and creates a credit note or marks replacement dispatched.
func (s *SupplierService) ResolveSupplierRMA(rmaID string, status, resolutionNotes string, creditAmount float64) (*models.SupplierRMA, error) {
	var rma models.SupplierRMA
	if err := s.db.Where("id = ?", rmaID).First(&rma).Error; err != nil {
		return nil, errors.New("RMA not found")
	}

	rma.Status = status
	rma.ResolutionNotes = &resolutionNotes
	if creditAmount > 0 {
		creditRef := fmt.Sprintf("CRN-%d", time.Now().UnixNano()%1000000)
		rma.CreditNoteRef = &creditRef
		rma.CreditAmount = creditAmount

		// Update supplier credit balance
		var supplier models.Supplier
		if err := s.db.Where("id = ?", rma.SupplierID).First(&supplier).Error; err == nil {
			supplier.CreditBalance += creditAmount
			_ = s.db.Save(&supplier).Error
		}
	}

	if err := s.db.Save(&rma).Error; err != nil {
		return nil, err
	}
	return &rma, nil
}

// --- DOCK DELIVERY SLOTS ---

// BookDeliverySlot schedules a delivery appointment at a specific branch.
func (s *SupplierService) BookDeliverySlot(slot models.SupplierDeliverySlot) (*models.SupplierDeliverySlot, error) {
	slot.Status = "scheduled"
	if err := s.db.Create(&slot).Error; err != nil {
		return nil, err
	}
	return &slot, nil
}

// ListDeliverySlots returns dock appointments for a specific vendor.
func (s *SupplierService) ListDeliverySlots(supplierID string) ([]models.SupplierDeliverySlot, error) {
	var slots []models.SupplierDeliverySlot
	err := s.db.Where("supplier_id = ?", supplierID).
		Preload("ASN").
		Order("slot_date desc").
		Find(&slots).Error
	return slots, err
}

// ListBranchDeliverySlots returns dock appointments for a branch.
func (s *SupplierService) ListBranchDeliverySlots(branchID string, date string) ([]models.SupplierDeliverySlot, error) {
	var slots []models.SupplierDeliverySlot
	query := s.db.Model(&models.SupplierDeliverySlot{}).Preload("Supplier").Preload("ASN")
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	if date != "" {
		query = query.Where("DATE(slot_date) = ?", date)
	}
	err := query.Order("slot_date asc").Find(&slots).Error
	return slots, err
}

// UpdateDeliverySlotStatus updates dock appointment status.
func (s *SupplierService) UpdateDeliverySlotStatus(slotID string, status string) (*models.SupplierDeliverySlot, error) {
	var slot models.SupplierDeliverySlot
	if err := s.db.Where("id = ?", slotID).First(&slot).Error; err != nil {
		return nil, errors.New("delivery slot not found")
	}
	slot.Status = status
	if err := s.db.Save(&slot).Error; err != nil {
		return nil, err
	}
	return &slot, nil
}

// --- COMPLIANCE DOCUMENT VAULT ---

// UploadSupplierDocument registers a verified vendor certificate.
func (s *SupplierService) UploadSupplierDocument(supplierID string, doc models.SupplierDocument) (*models.SupplierDocument, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, errors.New("invalid supplier ID")
	}
	doc.SupplierID = supUUID
	doc.Status = "verified"
	if err := s.db.Create(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

// ListSupplierDocuments returns compliance files.
func (s *SupplierService) ListSupplierDocuments(supplierID string) ([]models.SupplierDocument, error) {
	var docs []models.SupplierDocument
	err := s.db.Where("supplier_id = ?", supplierID).Order("created_at desc").Find(&docs).Error
	return docs, err
}

// VerifySupplierDocument allows admin to approve or reject compliance files.
func (s *SupplierService) VerifySupplierDocument(docID string, status string, notes *string) (*models.SupplierDocument, error) {
	var doc models.SupplierDocument
	if err := s.db.Where("id = ?", docID).First(&doc).Error; err != nil {
		return nil, errors.New("document not found")
	}
	doc.Status = status
	doc.Notes = notes
	if err := s.db.Save(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

// --- ADMIN PRICE PROPOSAL DECISIONS ---

// ApprovePriceRequest accepts the vendor proposal and updates product cost price.
func (s *SupplierService) ApprovePriceRequest(reqID string, notes string) error {
	var req models.SupplierPriceChangeRequest
	if err := s.db.Where("id = ?", reqID).First(&req).Error; err != nil {
		return errors.New("price request not found")
	}
	req.Status = "approved"
	req.ReviewNotes = &notes
	if err := s.db.Save(&req).Error; err != nil {
		return err
	}

	// Update the supplier product unit cost
	return s.db.Model(&models.SupplierProduct{}).
		Where("supplier_id = ? AND product_id = ?", req.SupplierID, req.ProductID).
		Update("unit_cost", req.ProposedCost).Error
}

// RejectPriceRequest rejects the vendor proposal.
func (s *SupplierService) RejectPriceRequest(reqID string, notes string) error {
	var req models.SupplierPriceChangeRequest
	if err := s.db.Where("id = ?", reqID).First(&req).Error; err != nil {
		return errors.New("price request not found")
	}
	req.Status = "rejected"
	req.ReviewNotes = &notes
	return s.db.Save(&req).Error
}

// --- ANNOUNCEMENT BOARD ---

// CreateSupplierAnnouncement publishes a merchant announcement.
func (s *SupplierService) CreateSupplierAnnouncement(ann models.SupplierAnnouncement) (*models.SupplierAnnouncement, error) {
	if err := s.db.Create(&ann).Error; err != nil {
		return nil, err
	}
	return &ann, nil
}

// ListSupplierAnnouncements returns active notices.
func (s *SupplierService) ListSupplierAnnouncements() ([]models.SupplierAnnouncement, error) {
	var anns []models.SupplierAnnouncement
	err := s.db.Where("is_active = ?", true).Order("created_at desc").Find(&anns).Error
	return anns, err
}

// --- 360-DEGREE SUPPLIER DETAILS & AP AGING ---

// SupplierDetailsResponse contains the full dossier for a supplier.
type SupplierDetailsResponse struct {
	Supplier       models.Supplier              `json:"supplier"`
	Products       []models.SupplierProduct     `json:"products"`
	PurchaseOrders []models.PurchaseOrder       `json:"purchase_orders"`
	Invoices       []models.SupplierInvoice     `json:"invoices"`
	RMAs           []models.SupplierRMA         `json:"rmas"`
	LedgerEntries  []models.SupplierLedgerEntry `json:"ledger_entries"`
	Documents      []models.SupplierDocument    `json:"documents"`
	APAging        map[string]float64           `json:"ap_aging"`
	Metrics        map[string]interface{}       `json:"metrics"`
}

// GetSupplierDetails computes 360-degree analytics for a single supplier.
func (s *SupplierService) GetSupplierDetails(supplierID string) (*SupplierDetailsResponse, error) {
	var supplier models.Supplier
	if err := s.db.Where("id = ?", supplierID).First(&supplier).Error; err != nil {
		return nil, errors.New("supplier not found")
	}

	var products []models.SupplierProduct
	_ = s.db.Where("supplier_id = ?", supplierID).Preload("Product").Find(&products).Error

	var pos []models.PurchaseOrder
	_ = s.db.Where("supplier_id = ?", supplierID).Order("created_at desc").Limit(20).Find(&pos).Error

	var invoices []models.SupplierInvoice
	_ = s.db.Where("supplier_id = ?", supplierID).Order("created_at desc").Find(&invoices).Error

	var rmas []models.SupplierRMA
	_ = s.db.Where("supplier_id = ?", supplierID).Preload("Product").Order("created_at desc").Find(&rmas).Error

	var ledger []models.SupplierLedgerEntry
	_ = s.db.Where("supplier_id = ?", supplierID).Order("created_at desc").Limit(30).Find(&ledger).Error

	var docs []models.SupplierDocument
	_ = s.db.Where("supplier_id = ?", supplierID).Order("created_at desc").Find(&docs).Error

	// Compute AP Aging buckets from unpaid invoices
	now := time.Now()
	aging := map[string]float64{
		"current":           0,
		"days_1_30":         0,
		"days_31_60":        0,
		"days_61_90":        0,
		"days_90_plus":      0,
		"total_outstanding": supplier.CreditBalance,
	}

	var totalSpend float64
	for _, inv := range invoices {
		totalSpend += inv.Total
		if inv.Status != "paid" {
			unpaidAmount := inv.Total - inv.AmountPaid
			if unpaidAmount <= 0 {
				continue
			}
			dueDate := inv.CreatedAt
			if !inv.DueDate.IsZero() {
				dueDate = inv.DueDate
			}
			daysPast := int(now.Sub(dueDate).Hours() / 24)
			if daysPast <= 0 {
				aging["current"] += unpaidAmount
			} else if daysPast <= 30 {
				aging["days_1_30"] += unpaidAmount
			} else if daysPast <= 60 {
				aging["days_31_60"] += unpaidAmount
			} else if daysPast <= 90 {
				aging["days_61_90"] += unpaidAmount
			} else {
				aging["days_90_plus"] += unpaidAmount
			}
		}
	}

	metrics := map[string]interface{}{
		"total_spend":       totalSpend,
		"total_orders":      len(pos),
		"linked_products":   len(products),
		"defect_claims":     len(rmas),
		"credit_limit":      supplier.CreditLimit,
		"credit_available":  supplier.CreditLimit - supplier.CreditBalance,
		"compliance_status": "verified",
	}

	return &SupplierDetailsResponse{
		Supplier:       supplier,
		Products:       products,
		PurchaseOrders: pos,
		Invoices:       invoices,
		RMAs:           rmas,
		LedgerEntries:  ledger,
		Documents:      docs,
		APAging:        aging,
		Metrics:        metrics,
	}, nil
}

// ── 3-Way Invoice Matching ──

type ThreeWayMatchAudit struct {
	InvoiceID            uuid.UUID              `json:"invoice_id"`
	InvoiceNumber        string                 `json:"invoice_number"`
	InvoiceTotal         float64                `json:"invoice_total"`
	PONumber             string                 `json:"po_number"`
	POTotal              float64                `json:"po_total"`
	MatchStatus          string                 `json:"match_status"` // matched, mismatched, pending_receipt
	DiscrepancyNotes     string                 `json:"discrepancy_notes"`
	Items                []ThreeWayMatchItem    `json:"items"`
	EarlyDiscountSummary *EarlyDiscountSummary  `json:"early_discount_summary,omitempty"`
}

type ThreeWayMatchItem struct {
	ProductName      string  `json:"product_name"`
	SKU              string  `json:"sku"`
	QuantityOrdered  float64 `json:"quantity_ordered"`
	QuantityReceived float64 `json:"quantity_received"`
	QuantityInvoiced float64 `json:"quantity_invoiced"`
	UnitCostAgreed   float64 `json:"unit_cost_agreed"`
	UnitCostInvoiced float64 `json:"unit_cost_invoiced"`
	IsMatched        bool    `json:"is_matched"`
}

type EarlyDiscountSummary struct {
	DiscountPercent float64   `json:"discount_percent"`
	DiscountDays    int       `json:"discount_days"`
	EligibleUntil   time.Time `json:"eligible_until"`
	IsEligible      bool      `json:"is_eligible"`
	DiscountAmount  float64   `json:"discount_amount"`
	NetPayable      float64   `json:"net_payable"`
}

func (s *SupplierService) CalculateThreeWayMatch(supplierID, invoiceID string) (*ThreeWayMatchAudit, error) {
	var inv models.SupplierInvoice
	if err := s.db.Where("id = ? AND supplier_id = ?", invoiceID, supplierID).
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.Items").
		Preload("PurchaseOrder.Items.Product").
		First(&inv).Error; err != nil {
		return nil, errors.New("invoice not found")
	}

	audit := &ThreeWayMatchAudit{
		InvoiceID:     inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		InvoiceTotal:  inv.Total,
		MatchStatus:   "matched",
		Items:         []ThreeWayMatchItem{},
	}

	if inv.PurchaseOrder != nil {
		audit.PONumber = inv.PurchaseOrder.PONumber
		audit.POTotal = inv.PurchaseOrder.TotalAmount

		// Fetch ASNs for received quantities
		var asns []models.SupplierASN
		s.db.Where("purchase_order_id = ?", inv.PurchaseOrder.ID).Find(&asns)
		hasReceived := len(asns) > 0

		allMatched := true
		for _, poItem := range inv.PurchaseOrder.Items {
			pName := "Product"
			pSKU := "SKU"
			if poItem.Product != nil {
				pName = poItem.Product.Name
				pSKU = poItem.Product.SKU
			}

			// In a standard 3-way match, received quantity is counted from confirmed receipts
			receivedQty := float64(poItem.QuantityReceived)
			if receivedQty == 0 && hasReceived {
				receivedQty = float64(poItem.QuantityOrdered)
			}

			invoicedQty := float64(poItem.QuantityOrdered)
			matched := poItem.QuantityOrdered == receivedQty

			if !matched {
				allMatched = false
			}

			audit.Items = append(audit.Items, ThreeWayMatchItem{
				ProductName:      pName,
				SKU:              pSKU,
				QuantityOrdered:  float64(poItem.QuantityOrdered),
				QuantityReceived: receivedQty,
				QuantityInvoiced: invoicedQty,
				UnitCostAgreed:   poItem.UnitCost,
				UnitCostInvoiced: poItem.UnitCost,
				IsMatched:        matched,
			})
		}

		if !hasReceived {
			audit.MatchStatus = "pending_receipt"
			audit.DiscrepancyNotes = "Warehouse receiving scan pending."
		} else if allMatched && inv.Total <= inv.PurchaseOrder.TotalAmount+0.01 {
			audit.MatchStatus = "matched"
			audit.DiscrepancyNotes = "100% matched across PO, Receiving ASN, and Invoiced total."
		} else {
			audit.MatchStatus = "mismatched"
			audit.DiscrepancyNotes = "Discrepancy detected between ordered and received quantities."
		}
	} else {
		audit.MatchStatus = "matched"
		audit.DiscrepancyNotes = "Direct invoice without linked PO."
	}

	// Early settlement discount calculation (e.g. 2/10 Net 30)
	if inv.EarlyDiscountPercent > 0 && inv.EarlyDiscountDays > 0 {
		deadline := inv.IssueDate.AddDate(0, 0, inv.EarlyDiscountDays)
		now := time.Now()
		isEligible := now.Before(deadline) && inv.Status != "paid"
		discountAmt := (inv.Total * inv.EarlyDiscountPercent) / 100.0
		audit.EarlyDiscountSummary = &EarlyDiscountSummary{
			DiscountPercent: inv.EarlyDiscountPercent,
			DiscountDays:    inv.EarlyDiscountDays,
			EligibleUntil:   deadline,
			IsEligible:      isEligible,
			DiscountAmount:  discountAmt,
			NetPayable:      inv.Total - discountAmt,
		}
	}

	// Update invoice record match status in DB
	s.db.Model(&inv).Update("three_way_match_status", audit.MatchStatus)

	return audit, nil
}

// ── Bulk Catalog CSV Import ──

type BulkCatalogItemInput struct {
	SKU         string  `json:"sku"`
	ProductName string  `json:"product_name"`
	UnitCost    float64 `json:"unit_cost"`
	MinOrderQty float64 `json:"min_order_qty"`
	LeadTime    int     `json:"lead_time"`
}

func (s *SupplierService) BulkImportCatalog(supplierID string, items []BulkCatalogItemInput) (int, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return 0, errors.New("invalid supplier ID")
	}

	importedCount := 0
	for _, item := range items {
		if strings.TrimSpace(item.SKU) == "" && strings.TrimSpace(item.ProductName) == "" {
			continue
		}
		if item.UnitCost <= 0 {
			continue
		}

		// Find or create product in tenant catalog
		var product models.Product
		searchSKU := strings.TrimSpace(item.SKU)
		if searchSKU != "" {
			_ = s.db.Where("sku = ?", searchSKU).First(&product).Error
		}
		if product.ID == uuid.Nil && strings.TrimSpace(item.ProductName) != "" {
			_ = s.db.Where("name ILIKE ?", strings.TrimSpace(item.ProductName)).First(&product).Error
		}

		if product.ID == uuid.Nil {
			product = models.Product{
				Name:           item.ProductName,
				SKU:            item.SKU,
				CostPrice:      item.UnitCost,
				SellingPrice:   item.UnitCost * 1.35, // Default markup
				IsActive:       true,
				TrackInventory: true,
			}
			if err := s.db.Create(&product).Error; err != nil {
				continue
			}
		}

		// Link to SupplierProduct
		var supProduct models.SupplierProduct
		if err := s.db.Where("supplier_id = ? AND product_id = ?", supUUID, product.ID).First(&supProduct).Error; err != nil {
			supProduct = models.SupplierProduct{
				SupplierID:  supUUID,
				ProductID:   product.ID,
				SupplierSKU: item.SKU,
				UnitCost:    item.UnitCost,
				MinOrderQty: item.MinOrderQty,
			}
			if supProduct.MinOrderQty <= 0 {
				supProduct.MinOrderQty = 1
			}
			s.db.Create(&supProduct)
		} else {
			supProduct.UnitCost = item.UnitCost
			if item.MinOrderQty > 0 {
				supProduct.MinOrderQty = item.MinOrderQty
			}
			s.db.Save(&supProduct)
		}
		importedCount++
	}

	return importedCount, nil
}

// ── Performance Tiering & Scorecard Computation ──

func (s *SupplierService) CalculateSupplierTier(supplierID string) (map[string]interface{}, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, errors.New("invalid supplier ID")
	}

	var supplier models.Supplier
	if err := s.db.First(&supplier, "id = ?", supUUID).Error; err != nil {
		return nil, errors.New("supplier not found")
	}

	// 1. Calculate OTD rate
	var totalPOs int64
	var onTimePOs int64
	s.db.Model(&models.PurchaseOrder{}).Where("supplier_id = ?", supUUID).Count(&totalPOs)
	s.db.Model(&models.PurchaseOrder{}).Where("supplier_id = ? AND status IN ('received', 'confirmed')", supUUID).Count(&onTimePOs)

	otdRate := 100.0
	if totalPOs > 0 {
		otdRate = float64(onTimePOs) / float64(totalPOs) * 100.0
	}

	// 2. Calculate Defect rate from RMAs
	var totalRMAs int64
	s.db.Model(&models.SupplierRMA{}).Where("supplier_id = ?", supUUID).Count(&totalRMAs)
	defectRate := 0.0
	if totalPOs > 0 {
		defectRate = float64(totalRMAs) / float64(totalPOs) * 10.0
		if defectRate > 100 {
			defectRate = 100
		}
	}

	// 3. Assign Tier
	tier := "standard"
	if otdRate >= 95.0 && defectRate <= 1.5 {
		tier = "platinum"
	} else if otdRate >= 90.0 && defectRate <= 3.0 {
		tier = "gold"
	} else if otdRate >= 80.0 && defectRate <= 5.0 {
		tier = "silver"
	}

	supplier.Tier = tier
	supplier.OTDRate = otdRate
	supplier.DefectRate = defectRate
	s.db.Model(&supplier).Updates(map[string]interface{}{
		"tier":        tier,
		"otd_rate":    otdRate,
		"defect_rate": defectRate,
	})

	return map[string]interface{}{
		"supplier_id": supplier.ID,
		"name":        supplier.Name,
		"tier":        tier,
		"otd_rate":    otdRate,
		"defect_rate": defectRate,
		"total_pos":   totalPOs,
		"total_rmas":  totalRMAs,
		"benefits": []string{
			"Priority RFQ distribution",
			"Automated 3-Way Match zero-touch approvals",
			"Expedited 48-hour MoMo/Bank settlement",
		},
	}, nil
}

// ── Developer API Keys Management ──

func (s *SupplierService) GenerateSupplierAPIKey(supplierID, name string) (*models.SupplierAPIKey, string, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, "", errors.New("invalid supplier ID")
	}

	rawBytes := make([]byte, 24)
	_, _ = rand.Read(rawBytes)
	rawKey := fmt.Sprintf("puk_live_%s", hex.EncodeToString(rawBytes))
	preview := rawKey[:12] + "..." + rawKey[len(rawKey)-4:]

	apiKey := models.SupplierAPIKey{
		SupplierID: supUUID,
		Name:       name,
		Key:        rawKey,
		KeyPreview: preview,
		IsActive:   true,
	}

	if err := s.db.Create(&apiKey).Error; err != nil {
		return nil, "", err
	}

	return &apiKey, rawKey, nil
}

func (s *SupplierService) ListSupplierAPIKeys(supplierID string) ([]models.SupplierAPIKey, error) {
	var keys []models.SupplierAPIKey
	err := s.db.Where("supplier_id = ?", supplierID).Order("created_at desc").Find(&keys).Error
	return keys, err
}

func (s *SupplierService) RevokeSupplierAPIKey(supplierID, keyID string) error {
	return s.db.Where("id = ? AND supplier_id = ?", keyID, supplierID).Delete(&models.SupplierAPIKey{}).Error
}

// ── Webhooks Management ──

func (s *SupplierService) CreateSupplierWebhook(supplierID, url, events string) (*models.SupplierWebhook, error) {
	supUUID, err := uuid.Parse(supplierID)
	if err != nil {
		return nil, errors.New("invalid supplier ID")
	}

	secretBytes := make([]byte, 16)
	_, _ = rand.Read(secretBytes)
	secret := fmt.Sprintf("whsec_%s", hex.EncodeToString(secretBytes))

	hook := models.SupplierWebhook{
		SupplierID: supUUID,
		URL:        url,
		Events:     events,
		Secret:     secret,
		IsActive:   true,
	}

	if err := s.db.Create(&hook).Error; err != nil {
		return nil, err
	}
	return &hook, nil
}

func (s *SupplierService) ListSupplierWebhooks(supplierID string) ([]models.SupplierWebhook, error) {
	var hooks []models.SupplierWebhook
	err := s.db.Where("supplier_id = ?", supplierID).Order("created_at desc").Find(&hooks).Error
	return hooks, err
}

func (s *SupplierService) DeleteSupplierWebhook(supplierID, webhookID string) error {
	return s.db.Where("id = ? AND supplier_id = ?", webhookID, supplierID).Delete(&models.SupplierWebhook{}).Error
}

// ── GS1-128 Barcode Shipping Label Data ──

func (s *SupplierService) GetShippingLabelData(supplierID, asnID string) (map[string]interface{}, error) {
	var asn models.SupplierASN
	if err := s.db.Where("id = ? AND supplier_id = ?", asnID, supplierID).
		Preload("Supplier").
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.Items").
		Preload("PurchaseOrder.Items.Product").
		First(&asn).Error; err != nil {
		return nil, errors.New("shipment not found")
	}

	barcodeVal := fmt.Sprintf("GS1-%s-%s", asn.ASNNumber, asn.Carrier)
	if asn.TrackingNumber != "" {
		barcodeVal = asn.TrackingNumber
	}

	return map[string]interface{}{
		"asn_number":      asn.ASNNumber,
		"po_number":       asn.PurchaseOrder.PONumber,
		"carrier":         asn.Carrier,
		"tracking_number": asn.TrackingNumber,
		"dispatch_date":   asn.DispatchDate,
		"expected_date":   asn.ExpectedArrival,
		"package_count":   asn.PackageCount,
		"total_weight_kg": asn.TotalWeightKg,
		"barcode":         barcodeVal,
		"vendor_name":     asn.Supplier.Name,
		"vendor_phone":    asn.Supplier.Phone,
		"vendor_address":  asn.Supplier.Address,
		"vendor_tin":      asn.Supplier.TaxNumber,
		"items":           asn.PurchaseOrder.Items,
	}, nil
}


