package models

import (
	"time"

	"github.com/google/uuid"
)

// Order represents a sales transaction.
type Order struct {
	BranchScoped
	OrderNumber string     `gorm:"size:50;uniqueIndex;not null" json:"order_number"`
	CustomerID  *uuid.UUID `gorm:"type:uuid;index" json:"customer_id,omitempty"`
	CashierID   *uuid.UUID `gorm:"type:uuid;index" json:"cashier_id,omitempty"`
	Cashier     *User      `gorm:"foreignKey:CashierID" json:"cashier,omitempty"`

	Subtotal   float64 `gorm:"type:decimal(12,2);default:0" json:"subtotal"`
	Tax        float64 `gorm:"type:decimal(12,2);default:0" json:"tax"`
	Discount   float64 `gorm:"type:decimal(12,2);default:0" json:"discount"`
	Total      float64 `gorm:"type:decimal(12,2);not null" json:"total"`
	AmountPaid float64 `gorm:"type:decimal(12,2);default:0" json:"amount_paid"`

	Status        string `gorm:"size:20;default:'completed';index" json:"status"` // pending, completed, cancelled, refunded
	PaymentStatus string `gorm:"size:20;default:'paid'" json:"payment_status"`    // unpaid, partial, paid, refunded
	PaymentMethod string `gorm:"size:20" json:"payment_method"`                   // cash, card, mobile, split, credit

	OrderType    string  `gorm:"size:20;default:'in_store'" json:"order_type"` // in_store, online, delivery, kiosk
	Notes        *string `gorm:"type:text" json:"notes,omitempty"`
	ReceiptToken string  `gorm:"size:64;uniqueIndex" json:"receipt_token"` // For public receipt URL

	// Contact & Pickup Verification (Online/Kiosk orders)
	CustomerName       *string    `gorm:"size:255" json:"customer_name,omitempty"`
	CustomerPhone      *string    `gorm:"size:50" json:"customer_phone,omitempty"`
	DeliveryAddress    *string    `gorm:"type:text" json:"delivery_address,omitempty"`
	PickupOTP          *string    `gorm:"size:20" json:"pickup_otp,omitempty"`
	PickupOTPExpiresAt *time.Time `json:"pickup_otp_expires_at,omitempty"`
	IsOTPVerified      bool       `gorm:"default:false" json:"is_otp_verified"`

	// Relations
	Customer *Customer   `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Items    []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`

	Version int `gorm:"default:1" json:"version"` // Optimistic Locking
}

// OrderItem represents a single product line in an order.
type OrderItem struct {
	Base
	OrderID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	VariantID *uuid.UUID `gorm:"type:uuid;index" json:"variant_id,omitempty"`

	Quantity  float64 `gorm:"type:decimal(10,3);not null" json:"quantity"`
	UnitPrice float64 `gorm:"type:decimal(12,2);not null" json:"unit_price"`
	Discount  float64 `gorm:"type:decimal(12,2);default:0" json:"discount"`
	Total     float64 `gorm:"type:decimal(12,2);not null" json:"total"` // (Quantity * UnitPrice) - Discount

	CostPrice float64 `gorm:"type:decimal(12,2);default:0" json:"cost_price"` // For profit calculation at time of sale

	// Relations
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// GlobalOrder is an aggregated read-only copy of orders for the Superuser dashboard.
type GlobalOrder struct {
	Base
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant       *Tenant   `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	CustomerName *string   `gorm:"type:text" json:"customer_name,omitempty"` // Encrypted
	TotalAmount  float64   `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status       string    `gorm:"size:20;not null" json:"status"`
	OrderDate    time.Time `json:"order_date"` // the created_at from the original order
	SyncedAt     time.Time `gorm:"autoUpdateTime" json:"synced_at"`
}
