package models

import (
	"time"

	"github.com/google/uuid"
)

// Quotation represents a B2B quote.
type Quotation struct {
	BranchScoped
	CustomerID  uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`
	Status      string    `gorm:"size:20;default:'draft';index" json:"status"`
	QuoteNumber string    `gorm:"size:30;uniqueIndex;not null" json:"quote_number"`

	Subtotal    float64 `gorm:"type:decimal(12,2);default:0" json:"subtotal"`
	TaxAmount   float64 `gorm:"type:decimal(12,2);default:0" json:"tax_amount"`
	TotalAmount float64 `gorm:"type:decimal(12,2);default:0" json:"total_amount"`

	Notes         *string    `gorm:"type:text" json:"notes,omitempty"`
	InternalNotes *string    `gorm:"type:text" json:"internal_notes,omitempty"`
	ValidUntil    *time.Time `json:"valid_until,omitempty"`

	CreatedByID *uuid.UUID `gorm:"type:uuid" json:"created_by_id,omitempty"`

	CreatedBy    *User      `gorm:"foreignKey:CreatedByID" json:"createdby,omitempty"`
	ReviewedByID *uuid.UUID `gorm:"type:uuid" json:"reviewed_by_id,omitempty"`
	ReviewedBy   *User      `gorm:"foreignKey:ReviewedByID" json:"reviewedby,omitempty"`

	ConvertedOrderID *uuid.UUID `gorm:"type:uuid" json:"converted_order_id,omitempty"`

	// Relations
	Items    []QuotationItem `gorm:"foreignKey:QuotationID" json:"items,omitempty"`
	Customer *Customer       `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

// QuotationItem is a line item in a quote.
type QuotationItem struct {
	Base
	QuotationID uuid.UUID `gorm:"type:uuid;not null;index" json:"quotation_id"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	Quantity    uint      `gorm:"default:1" json:"quantity"`
	UnitPrice   float64   `gorm:"type:decimal(12,2);not null" json:"unit_price"`
	Discount    float64   `gorm:"type:decimal(12,2);default:0" json:"discount"`
	TotalPrice  float64   `gorm:"type:decimal(12,2);not null" json:"total_price"`
}
