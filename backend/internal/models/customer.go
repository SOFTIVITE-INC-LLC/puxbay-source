package models

import (
	"time"

	"github.com/google/uuid"
)

// CustomerTier represents a loyalty tier.
type CustomerTier struct {
	TenantScoped
	Name               string  `gorm:"size:50;not null" json:"name"`
	MinSpend           float64 `gorm:"type:decimal(12,2);default:0" json:"min_spend"`
	DiscountPercentage float64 `gorm:"type:decimal(5,2);default:0" json:"discount_percentage"`
	Color              string  `gorm:"size:20;default:'blue'" json:"color"`
	Icon               string  `gorm:"size:50;default:'star'" json:"icon"`
}

// Customer represents a buyer.
type Customer struct {
	TenantScoped
	Name         string  `gorm:"size:200;not null;index" json:"name"`
	Phone        *string `gorm:"type:text" json:"phone,omitempty"`       // Encrypted
	Email        *string `gorm:"type:text;index" json:"email,omitempty"` // Encrypted
	Address      *string `gorm:"type:text" json:"address,omitempty"`     // Encrypted
	PasswordHash *string `gorm:"size:255" json:"-"`
	IsRegistered bool    `gorm:"default:false" json:"is_registered"`

	TierID *uuid.UUID `gorm:"type:uuid" json:"tier_id,omitempty"`

	TotalSpend  float64 `gorm:"type:decimal(12,2);default:0" json:"total_spend"`
	OrderCount  uint    `gorm:"default:0" json:"order_count"`
	LoyaltyPts  float64 `gorm:"type:decimal(10,2);default:0" json:"loyalty_points"`
	StoreCredit  float64 `gorm:"type:decimal(12,2);default:0" json:"store_credit"`
	DebtBalance  float64 `gorm:"type:decimal(12,2);default:0" json:"debt_balance"`
	CreditLimit  float64 `gorm:"type:decimal(12,2);default:0" json:"credit_limit"`
	CreditStatus string  `gorm:"size:20;default:'active'" json:"credit_status"`

	AcceptsMarketing bool       `gorm:"default:true" json:"accepts_marketing"`
	LastVisit        *time.Time `json:"last_visit,omitempty"`
	DateOfBirth      *time.Time `json:"date_of_birth,omitempty"`
	Notes            *string    `gorm:"type:text" json:"notes,omitempty"`
	CustomerType     string     `gorm:"size:20;default:'retail'" json:"customer_type"`

	// Relations
	Tier *CustomerTier `gorm:"foreignKey:TierID" json:"tier,omitempty"`
}

// LoyaltyTransaction tracks point changes.
type LoyaltyTransaction struct {
	TenantScoped
	CustomerID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"customer_id"`
	OrderID         *uuid.UUID `gorm:"type:uuid" json:"order_id,omitempty"`
	Points          float64    `gorm:"type:decimal(10,2);not null" json:"points"` // Can be negative
	TransactionType string     `gorm:"size:20;not null" json:"transaction_type"`  // earned, redeemed, manual, expired
	Description     *string    `gorm:"type:text" json:"description,omitempty"`    // Encrypted
}

// GiftCard represents a pre-paid balance.
type GiftCard struct {
	TenantScoped
	Code           string     `gorm:"size:50;uniqueIndex;not null" json:"code"`
	InitialBalance float64    `gorm:"type:decimal(10,2);not null" json:"initial_balance"`
	CurrentBalance float64    `gorm:"type:decimal(10,2);not null" json:"current_balance"`
	PurchaserID    *uuid.UUID `gorm:"type:uuid" json:"purchaser_id,omitempty"`
	RecipientEmail *string    `gorm:"size:254" json:"recipient_email,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	Status         string     `gorm:"size:20;default:'active'" json:"status"`
}

// StoreCreditTransaction tracks changes to a customer's store credit.
type StoreCreditTransaction struct {
	TenantScoped
	CustomerID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"customer_id"`
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount"` // positive = added, negative = used
	TransactionType string     `gorm:"size:20;not null" json:"transaction_type"`  // purchase, refund, manual
	Reference       string     `gorm:"size:100" json:"reference,omitempty"`
	Notes           string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedByID     *uuid.UUID `gorm:"type:uuid" json:"created_by_id,omitempty"`
	CreatedBy       *User      `gorm:"foreignKey:CreatedByID" json:"createdby,omitempty"`
}
