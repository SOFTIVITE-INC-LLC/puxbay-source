package models

import (
	"time"

	"github.com/google/uuid"
)

// SMSGatewayConfig stores the global Arkesel gateway settings managed by Superadmin.
// This lives in the public schema and is shared platform-wide.
type SMSGatewayConfig struct {
	Base
	Provider        string     `gorm:"size:50;default:'arkesel'" json:"provider"` // arkesel (system gateway)
	DefaultSenderID string     `gorm:"size:20;default:'PUXBAY'" json:"default_sender_id"`
	PricePerSMS     float64    `gorm:"type:decimal(10,4);default:0.20" json:"price_per_sms"` // in tenant's chosen currency
	PriceCurrency   string     `gorm:"size:10;default:'GHS'" json:"price_currency"`
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	UpdatedByUserID *uuid.UUID `gorm:"type:uuid" json:"updated_by_user_id,omitempty"`
}

// SMSSenderID is a per-tenant request to use a custom Sender ID (e.g. "MYBRAND").
// Lives in the TENANT schema.
type SMSSenderID struct {
	TenantScoped
	SenderID        string     `gorm:"size:20;not null;index" json:"sender_id"`       // max 11 alphanumeric per GSM spec
	Purpose         string     `gorm:"size:500" json:"purpose"`                       // business justification
	Status          string     `gorm:"size:20;default:'pending';index" json:"status"` // pending, approved, rejected
	RejectionReason string     `gorm:"type:text" json:"rejection_reason,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	ApprovedBy      *uuid.UUID `gorm:"type:uuid" json:"approved_by,omitempty"`
	TenantID        string     `gorm:"size:100;not null;index" json:"tenant_id"` // subdomain
}

// SMSWallet holds the balance information for a tenant's SMS credit wallet.
// Lives in the TENANT schema. One wallet per tenant.
type SMSWallet struct {
	TenantScoped
	TenantID       string  `gorm:"size:100;uniqueIndex;not null" json:"tenant_id"`       // subdomain
	BalanceAmount  float64 `gorm:"type:decimal(12,4);default:0" json:"balance_amount"`   // monetary balance remaining
	CreditsTotal   int64   `gorm:"default:0" json:"credits_total"`                       // total SMS credits ever topped up
	CreditsUsed    int64   `gorm:"default:0" json:"credits_used"`                        // total SMS messages sent
	CreditsBalance int64   `gorm:"default:0" json:"credits_balance"`                     // current available credits
	PricePerSMS    float64 `gorm:"type:decimal(10,4);default:0.20" json:"price_per_sms"` // at time of last topup
	Currency       string  `gorm:"size:10;default:'GHS'" json:"currency"`
}

// SMSTransaction records each wallet top-up or deduction event.
// Lives in the TENANT schema.
type SMSTransaction struct {
	TenantScoped
	TenantID      string  `gorm:"size:100;not null;index" json:"tenant_id"`             // subdomain
	Type          string  `gorm:"size:20;not null" json:"type"`                         // topup, deduction
	Amount        float64 `gorm:"type:decimal(12,4);default:0" json:"amount"`           // monetary amount
	CreditsAdded  int64   `gorm:"default:0" json:"credits_added"`                       // credits gained
	CreditsUsed   int64   `gorm:"default:0" json:"credits_used"`                        // credits consumed (for deductions)
	PricePerSMS   float64 `gorm:"type:decimal(10,4);default:0.20" json:"price_per_sms"` // rate at time of transaction
	Reference     string  `gorm:"size:255;index" json:"reference,omitempty"`            // Paystack reference
	PaymentMethod string  `gorm:"size:50" json:"payment_method,omitempty"`              // paystack, manual
	Status        string  `gorm:"size:20;default:'completed'" json:"status"`            // pending, completed, failed
	Description   string  `gorm:"size:500" json:"description,omitempty"`
}

func (SMSGatewayConfig) TableName() string { return "sms_gateway_configs" }
func (SMSSenderID) TableName() string      { return "sms_sender_ids" }
func (SMSWallet) TableName() string        { return "sms_wallets" }
func (SMSTransaction) TableName() string   { return "sms_transactions" }
