package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// PaymentLog represents a platform-wide payment record (storefront order, POS, subscription, manual settlement, disputes).
type PaymentLog struct {
	Base
	TenantID           *uuid.UUID     `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	TenantName         string         `gorm:"size:150" json:"tenant_name"`
	PaymentType        string         `gorm:"size:50;default:'store_order';index" json:"payment_type"` // store_order, pos_order, subscription, manual_entry, dispute_settlement, refund
	Reference          string         `gorm:"size:120;index" json:"reference"`
	OrderID            *uuid.UUID     `gorm:"type:uuid;index" json:"order_id,omitempty"`
	OrderNumber        string         `gorm:"size:60;index" json:"order_number"`
	Amount             float64        `gorm:"type:decimal(12,2);not null" json:"amount"`
	Currency           string         `gorm:"size:10;default:'GHS'" json:"currency"`
	PaymentMethod      string         `gorm:"size:50;default:'paystack'" json:"payment_method"` // paystack, card, mobile_money, cash, bank_transfer, split
	Gateway            string         `gorm:"size:50;default:'paystack'" json:"gateway"`        // paystack, stripe, manual, offline
	SubaccountCode     *string        `gorm:"size:100;index" json:"subaccount_code,omitempty"`
	IsSubaccountRouted bool           `gorm:"default:false;index" json:"is_subaccount_routed"`
	SubaccountShare    float64        `gorm:"type:decimal(12,2);default:0" json:"subaccount_share"`
	PlatformFee        float64        `gorm:"type:decimal(12,2);default:0" json:"platform_fee"`
	CustomerName       string         `gorm:"size:200" json:"customer_name"`
	CustomerEmail      string         `gorm:"size:150" json:"customer_email"`
	CustomerPhone      string         `gorm:"size:50" json:"customer_phone"`
	Status             string         `gorm:"size:30;default:'successful';index" json:"status"` // successful, pending, failed, refunded, disputed
	DisputeStatus      string         `gorm:"size:30;default:'none';index" json:"dispute_status"` // none, under_review, resolved, chargeback
	Notes              string         `gorm:"type:text" json:"notes"`
	RawMetadata        datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"raw_metadata,omitempty"`

	// Relations
	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

func (PaymentLog) TableName() string {
	return "public.payment_logs"
}
