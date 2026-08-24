package models

import (
	"github.com/google/uuid"
)

// LedgerAccount represents a category in the chart of accounts.
type LedgerAccount struct {
	TenantScoped
	Name        string `gorm:"size:100;not null" json:"name"`
	Type        string `gorm:"size:50;not null" json:"type"` // Asset, Liability, Equity, Revenue, Expense
	Code        string `gorm:"size:20;uniqueIndex" json:"code"`
	Description string `gorm:"type:text" json:"description"`
}

// JournalEntry represents a double-entry bookkeeping transaction.
type JournalEntry struct {
	TenantScoped
	ReferenceID   *uuid.UUID   `gorm:"type:uuid;index" json:"reference_id,omitempty"` // E.g., Order ID or Purchase Order ID
	ReferenceType string       `gorm:"size:50" json:"reference_type"`                 // "Order", "Purchase", "Manual"
	Description   string       `gorm:"type:text" json:"description"`
	Lines         []LedgerLine `gorm:"foreignKey:JournalEntryID" json:"lines,omitempty"`
}

// LedgerLine represents a single debit or credit in a journal entry.
type LedgerLine struct {
	Base
	JournalEntryID uuid.UUID `gorm:"type:uuid;not null;index" json:"journal_entry_id"`
	AccountID      uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id"`
	Amount         float64   `gorm:"type:decimal(12,2);not null" json:"amount"`
	IsDebit        bool      `gorm:"not null" json:"is_debit"` // True for Debit, False for Credit

	// Relations
	Account *LedgerAccount `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}
