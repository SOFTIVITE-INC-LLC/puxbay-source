package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ExpenseCategory categorizes business expenses.
type ExpenseCategory struct {
	TenantScoped
	Name          string  `gorm:"size:100;not null" json:"name"`
	Type          string  `gorm:"size:20;default:'variable'" json:"type"` // fixed, variable
	Description   *string `gorm:"type:text" json:"description,omitempty"`
	MonthlyBudget float64 `gorm:"type:decimal(12,2);default:0" json:"monthly_budget"`
}

// Expense tracks operational costs.
type Expense struct {
	BranchScoped
	CategoryID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"category_id"`
	Category           *Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Amount             float64    `gorm:"type:decimal(12,2);not null" json:"amount"`
	Date               time.Time  `json:"date"`
	Description        *string    `gorm:"type:text" json:"description,omitempty"` // Encrypted
	ReceiptURL         *string    `gorm:"size:512" json:"receipt_url,omitempty"`
	IsRecurring        bool       `gorm:"default:false" json:"is_recurring"`
	RecurrenceInterval string     `gorm:"size:20;default:''" json:"recurrence_interval"` // weekly, monthly, yearly
	CreatedByID        *uuid.UUID `gorm:"type:uuid" json:"created_by_id,omitempty"`
	CreatedBy          *User      `gorm:"foreignKey:CreatedByID" json:"createdby,omitempty"`
}

// PaymentMethod configures payment gateways (Stripe, Cash, Paystack Subaccount, etc).
type PaymentMethod struct {
	TenantScoped
	Name                   string  `gorm:"size:100;not null" json:"name"`
	Provider               string  `gorm:"size:40;not null" json:"provider"` // stripe, paystack, paystack_subaccount, cash, card, mobile, bank_transfer
	IsActive               bool    `gorm:"default:true" json:"is_active"`
	APIKeyHint             *string `gorm:"size:50" json:"api_key_hint,omitempty"`
	PaystackSubaccountCode *string `gorm:"size:100" json:"paystack_subaccount_code,omitempty"`
}

// Payment tracks order payments.
type Payment struct {
	TenantScoped
	OrderID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"order_id"`
	PaymentMethodID uuid.UUID      `gorm:"type:uuid;not null;index" json:"payment_method_id"`
	Amount          float64        `gorm:"type:decimal(12,2);not null" json:"amount"`
	Status          string         `gorm:"size:20;default:'pending'" json:"status"` // pending, completed, failed, refunded
	TransactionID   *string        `gorm:"size:255" json:"transaction_id,omitempty"`
	Metadata        datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	ErrorMessage    *string        `gorm:"type:text" json:"error_message,omitempty"` // Encrypted
}

// TaxConfiguration holds VAT/Sales Tax settings.
type TaxConfiguration struct {
	TenantScoped
	TaxType            string  `gorm:"size:20;default:'sales_tax'" json:"tax_type"` // vat, sales_tax, gst, none
	TaxRate            float64 `gorm:"type:decimal(5,2);default:0" json:"tax_rate"`
	TaxNumber          *string `gorm:"size:50" json:"tax_number,omitempty"`
	IncludeTaxInPrices bool    `gorm:"default:false" json:"include_tax_in_prices"`
	IsActive           bool    `gorm:"default:true" json:"is_active"`
}

// Return tracks product refunds.
type Return struct {
	BranchScoped
	OrderID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"order_id"`
	CustomerID    *uuid.UUID `gorm:"type:uuid" json:"customer_id,omitempty"`
	Reason        string     `gorm:"size:50;not null" json:"reason"`                  // defective, wrong_item, not_as_described, changed_mind, other
	ReasonDetail  *string    `gorm:"type:text" json:"reason_detail,omitempty"`        // Encrypted
	Status        string     `gorm:"size:20;default:'pending'" json:"status"`         // pending, approved, rejected, completed
	RefundMethod  string     `gorm:"size:20;default:'original'" json:"refund_method"` // cash, card, store_credit, original
	RefundAmount  float64    `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`
	RestockingFee float64    `gorm:"type:decimal(12,2);default:0" json:"restocking_fee"`
	TotalRefund   float64    `gorm:"-" json:"total_refund"` // computed: RefundAmount - RestockingFee
	CreatedByID   *uuid.UUID `gorm:"type:uuid" json:"created_by_id,omitempty"`
	CreatedBy     *User      `gorm:"foreignKey:CreatedByID" json:"createdby,omitempty"`
	ApprovedByID  *uuid.UUID `gorm:"type:uuid" json:"approved_by_id,omitempty"`
	ApprovedBy    *User      `gorm:"foreignKey:ApprovedByID" json:"approvedby,omitempty"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`

	// Relations
	Customer *Customer    `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Order    *Order       `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Items    []ReturnItem `gorm:"foreignKey:ReturnID" json:"items,omitempty"`
}

// AfterFind populates the computed TotalRefund field after any query.
func (r *Return) AfterFind(tx *gorm.DB) error {
	r.TotalRefund = r.RefundAmount - r.RestockingFee
	return nil
}

// ReturnItem is a line item in a return request.
type ReturnItem struct {
	Base
	ReturnID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"return_id"`
	ProductID *uuid.UUID `gorm:"type:uuid" json:"product_id,omitempty"`
	Quantity  float64    `gorm:"type:decimal(10,3);default:1" json:"quantity"`
	Condition string     `gorm:"size:20;default:'opened'" json:"condition"` // unopened, opened, damaged
	Restock   bool       `gorm:"default:false" json:"restock"`
	UnitPrice float64    `gorm:"type:decimal(12,2);not null" json:"unit_price"`

	// Relations
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// Currency configures additional currencies for a tenant.
type Currency struct {
	TenantScoped
	Code         string  `gorm:"size:3;not null" json:"code"`
	Name         string  `gorm:"size:50;not null" json:"name"`
	Symbol       string  `gorm:"size:5;not null" json:"symbol"`
	ExchangeRate float64 `gorm:"type:decimal(10,4);default:1.0" json:"exchange_rate"`
	IsBase       bool    `gorm:"default:false" json:"is_base"`
	IsActive     bool    `gorm:"default:true" json:"is_active"`
}
