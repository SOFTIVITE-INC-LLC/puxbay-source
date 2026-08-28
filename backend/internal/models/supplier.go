package models

import (
	"time"

	"github.com/google/uuid"
)

// Supplier represents a vendor/supplier.
type Supplier struct {
	TenantScoped
	Name          string  `gorm:"size:200;not null" json:"name"`
	ContactPerson *string `gorm:"size:150" json:"contact_person,omitempty"`
	Email         *string `gorm:"type:text" json:"email,omitempty"`   // Encrypted
	Phone         *string `gorm:"type:text" json:"phone,omitempty"`   // Encrypted
	Address       *string `gorm:"type:text" json:"address,omitempty"` // Encrypted
	Website       *string `gorm:"size:255" json:"website,omitempty"`
	TaxNumber     *string `gorm:"size:50" json:"tax_number,omitempty"`

	PaymentTerms *string `gorm:"type:text" json:"payment_terms,omitempty"`
	Notes        *string `gorm:"type:text" json:"notes,omitempty"`

	CreditBalance float64 `gorm:"type:decimal(12,2);default:0" json:"credit_balance"`
	IsActive      bool    `gorm:"default:true" json:"is_active"`

	// Portal access — set when supplier is invited to the portal
	PortalEmail    *string `gorm:"size:254;index" json:"portal_email,omitempty"`
	PortalPassword *string `gorm:"size:255" json:"-"` // bcrypt hash
}

// SupplierProfile links a user to a supplier for the B2B portal.
type SupplierProfile struct {
	Base
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	SupplierID uuid.UUID `gorm:"type:uuid;not null;index" json:"supplier_id"`
	Role       string    `gorm:"size:50;default:'contact'" json:"role"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`

	// Relations
	User     User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Supplier Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}

// SupplierProduct links a supplier to a specific product they sell
type SupplierProduct struct {
	Base
	SupplierID  uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_supplier_product" json:"supplier_id"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_supplier_product" json:"product_id"`
	SupplierSKU string    `gorm:"size:100" json:"supplier_sku,omitempty"`
	UnitCost    float64   `gorm:"type:decimal(12,2);not null;default:0" json:"unit_cost"`
	MinOrderQty float64   `gorm:"type:decimal(10,4);default:1" json:"min_order_qty"`

	// Relations
	Supplier Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Product  Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// SupplierLedgerEntry tracks financial transactions with a supplier (Invoices and Payments)
type SupplierLedgerEntry struct {
	Base
	SupplierID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"supplier_id"`
	EntryType       string     `gorm:"size:50;not null" json:"entry_type"`     // "invoice", "payment", "adjustment"
	ReferenceID     *string    `gorm:"size:100" json:"reference_id,omitempty"` // PO Number or Payment Ref
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount"`
	Balance         float64    `gorm:"type:decimal(12,2);not null" json:"balance"` // Snapshot of balance after this entry
	Notes           *string    `gorm:"type:text" json:"notes,omitempty"`
	TransactionDate *time.Time `gorm:"index" json:"transaction_date,omitempty"`

	// Relations
	Supplier Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}

// SupplierASN represents an Advanced Shipping Notice sent by the supplier before goods arrive.
type SupplierASN struct {
	Base
	SupplierID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"supplier_id"`
	PurchaseOrderID *uuid.UUID `gorm:"type:uuid;index" json:"purchase_order_id,omitempty"`
	ASNNumber       string     `gorm:"size:50;uniqueIndex;not null" json:"asn_number"`
	Carrier         string     `gorm:"size:100;not null" json:"carrier"` // e.g. DHL, FedEx, Local Fleet
	TrackingNumber  string     `gorm:"size:100" json:"tracking_number,omitempty"`
	DispatchDate    time.Time  `json:"dispatch_date"`
	ExpectedArrival *time.Time `json:"expected_arrival,omitempty"`
	PackageCount    int        `gorm:"default:1" json:"package_count"`
	TotalWeightKg   float64    `gorm:"type:decimal(10,2);default:0" json:"total_weight_kg"`
	Status          string     `gorm:"size:30;default:'dispatched'" json:"status"` // dispatched, in_transit, delivered, rejected
	Notes           *string    `gorm:"type:text" json:"notes,omitempty"`
	WaybillURL      *string    `gorm:"size:512" json:"waybill_url,omitempty"`

	// Relations
	Supplier      Supplier       `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	PurchaseOrder *PurchaseOrder `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`
}

// SupplierInvoice represents an AP invoice submitted by the supplier for fulfilled purchase orders.
type SupplierInvoice struct {
	Base
	SupplierID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"supplier_id"`
	PurchaseOrderID *uuid.UUID `gorm:"type:uuid;index" json:"purchase_order_id,omitempty"`
	InvoiceNumber   string     `gorm:"size:50;not null" json:"invoice_number"`
	IssueDate       time.Time  `json:"issue_date"`
	DueDate         time.Time  `json:"due_date"`
	Subtotal        float64    `gorm:"type:decimal(12,2);not null" json:"subtotal"`
	Tax             float64    `gorm:"type:decimal(12,2);default:0" json:"tax"`
	Total           float64    `gorm:"type:decimal(12,2);not null" json:"total"`
	AmountPaid      float64    `gorm:"type:decimal(12,2);default:0" json:"amount_paid"`
	Status          string     `gorm:"size:30;default:'pending'" json:"status"` // pending, approved, partially_paid, paid, rejected
	PaymentRef      *string    `gorm:"size:100" json:"payment_ref,omitempty"`
	Notes           *string    `gorm:"type:text" json:"notes,omitempty"`

	// Relations
	Supplier      Supplier       `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	PurchaseOrder *PurchaseOrder `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`
}

// SupplierPriceChangeRequest tracks supplier proposals to update unit cost on catalog items.
type SupplierPriceChangeRequest struct {
	Base
	SupplierID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"supplier_id"`
	ProductID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	CurrentCost       float64    `gorm:"type:decimal(12,2);not null" json:"current_cost"`
	ProposedCost      float64    `gorm:"type:decimal(12,2);not null" json:"proposed_cost"`
	EffectiveDate     time.Time  `json:"effective_date"`
	Reason            string     `gorm:"type:text;not null" json:"reason"`
	Status            string     `gorm:"size:30;default:'pending'" json:"status"` // pending, approved, rejected
	ReviewedByID      *uuid.UUID `gorm:"type:uuid" json:"reviewed_by_id,omitempty"`
	ReviewNotes       *string    `gorm:"type:text" json:"review_notes,omitempty"`

	// Relations
	Supplier Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Product  Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// SupplierQuote represents a quotation submitted by a supplier in response to a RFQ.
type SupplierQuote struct {
	Base
	SupplierID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"supplier_id"`
	QuoteNumber     string     `gorm:"size:50;uniqueIndex;not null" json:"quote_number"`
	Title           string     `gorm:"size:200;not null" json:"title"`
	TotalAmount     float64    `gorm:"type:decimal(12,2);not null" json:"total_amount"`
	Currency        string     `gorm:"size:10;default:'USD'" json:"currency"`
	ValidUntil      time.Time  `json:"valid_until"`
	LeadTimeDays    int        `gorm:"default:7" json:"lead_time_days"`
	PaymentTerms    string     `gorm:"size:100;default:'Net 30'" json:"payment_terms"`
	Status          string     `gorm:"size:30;default:'submitted'" json:"status"` // draft, submitted, accepted, rejected, expired
	Notes           *string    `gorm:"type:text" json:"notes,omitempty"`

	// Relations
	Supplier Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}

// SupplierPayoutAccount stores bank and Mobile Money disbursement information.
type SupplierPayoutAccount struct {
	Base
	SupplierID    uuid.UUID `gorm:"type:uuid;not null;index" json:"supplier_id"`
	AccountType   string    `gorm:"size:30;default:'bank'" json:"account_type"` // bank, momo, stripe
	BankName      *string   `gorm:"size:150" json:"bank_name,omitempty"`
	AccountNumber *string   `gorm:"size:100" json:"account_number,omitempty"`
	AccountName   *string   `gorm:"size:150" json:"account_name,omitempty"`
	RoutingCode   *string   `gorm:"size:50" json:"routing_code,omitempty"`
	MoMoNetwork   *string   `gorm:"size:50" json:"momo_network,omitempty"` // MTN, Telecel, AirtelTigo, M-Pesa
	MoMoNumber    *string   `gorm:"size:50" json:"momo_number,omitempty"`
	IsDefault     bool      `gorm:"default:true" json:"is_default"`

	// Relations
	Supplier Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}

// SupplierMessage represents contextual threaded communications regarding POs or Invoices.
type SupplierMessage struct {
	Base
	SupplierID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"supplier_id"`
	SenderType  string     `gorm:"size:20;not null" json:"sender_type"` // supplier, merchant
	SenderName  string     `gorm:"size:150;not null" json:"sender_name"`
	ReferenceID *string    `gorm:"size:100;index" json:"reference_id,omitempty"` // PO-123 or INV-456
	Message     string     `gorm:"type:text;not null" json:"message"`
	IsRead      bool       `gorm:"default:false" json:"is_read"`

	// Relations
	Supplier Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}

// SupplierTeamMember represents a staff user within the supplier's organization with role-based access.
type SupplierTeamMember struct {
	Base
	SupplierID uuid.UUID `gorm:"type:uuid;not null;index" json:"supplier_id"`
	FullName   string    `gorm:"size:150;not null" json:"full_name"`
	Email      string    `gorm:"size:254;not null;index" json:"email"`
	Role       string    `gorm:"size:50;default:'warehouse'" json:"role"` // admin, finance, warehouse, sales
	Phone      *string   `gorm:"size:50" json:"phone,omitempty"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`

	// Relations
	Supplier Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}
