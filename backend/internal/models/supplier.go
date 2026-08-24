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
