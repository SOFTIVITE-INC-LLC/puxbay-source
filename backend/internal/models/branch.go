package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Branch represents a physical store location for a tenant.
// Maps from Django: accounts.models.Branch
type Branch struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	UniqueID  *string   `gorm:"type:text;uniqueIndex" json:"unique_id,omitempty"` // Encrypted: BR-0001
	Address   *string   `gorm:"type:text" json:"address,omitempty"`               // Encrypted
	Latitude  *float64  `gorm:"type:decimal(9,6)" json:"latitude,omitempty"`
	Longitude *float64  `gorm:"type:decimal(9,6)" json:"longitude,omitempty"`
	Phone     *string   `gorm:"type:text" json:"phone,omitempty"` // Encrypted

	// Settings
	PrimaryColor      string  `gorm:"size:7;default:'#4f46e5'" json:"primary_color"`
	Logo              *string `gorm:"size:512" json:"logo,omitempty"`
	LowStockThreshold uint    `gorm:"default:10" json:"low_stock_threshold"`
	CurrencySymbol    string  `gorm:"size:5;default:'GH₵'" json:"currency_symbol"`
	CurrencyCode      string  `gorm:"size:3;default:'GHS'" json:"currency_code"`
	ReceiptHeader     *string `gorm:"type:text" json:"receipt_header,omitempty"`
	ReceiptFooter     *string `gorm:"type:text" json:"receipt_footer,omitempty"`
	BranchType        string  `gorm:"size:20;default:'retail'" json:"branch_type"` // retail, wholesale

	// Sync Health
	LastSyncAt       *time.Time `json:"last_sync_at,omitempty"`
	SyncStatus       string     `gorm:"size:20;default:'healthy'" json:"sync_status"`
	PendingSyncCount uint       `gorm:"default:0" json:"pending_sync_count"`
	SyncErrorMessage *string    `gorm:"type:text" json:"sync_error_message,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

func (b *Branch) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ValidCurrencies returns supported currency codes.
func ValidCurrencies() []string {
	return []string{"USD", "NGN", "GHS", "ZAR", "EUR", "GBP"}
}

// ValidBranchTypes returns supported branch types.
func ValidBranchTypes() []string {
	return []string{"retail", "wholesale"}
}

// CashDrawerSession tracks till openings/closings.
type CashDrawerSession struct {
	Base
	BranchID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"branch_id"`
	Branch         *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	OpeningBalance float64    `gorm:"type:decimal(12,2);not null" json:"opening_balance"`
	ClosingBalance float64    `gorm:"type:decimal(12,2)" json:"closing_balance"`
	OpenedAt       time.Time  `gorm:"autoCreateTime" json:"opened_at"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	Notes          *string    `gorm:"type:text" json:"notes,omitempty"`
}

// Shift tracks employee work periods.
type Shift struct {
	Base
	BranchID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"branch_id"`
	Branch    *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
}

// PrintJob queues documents/receipts for remote printing.
type PrintJob struct {
	Base
	BranchID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"branch_id"`
	Branch       *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	DocumentType string     `gorm:"size:50;not null" json:"document_type"` // receipt, report, label
	Content      string     `gorm:"type:text;not null" json:"content"`
	Status       string     `gorm:"size:20;default:'pending'" json:"status"` // pending, printed, failed
	PrintedAt    *time.Time `json:"printed_at,omitempty"`
}
