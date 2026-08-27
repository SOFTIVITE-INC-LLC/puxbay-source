package models

import (
	"time"

	"github.com/google/uuid"
)

// CreditAccount represents a customer's revolving store credit or BNPL account.
type CreditAccount struct {
	TenantScoped
	CustomerID  uuid.UUID  `gorm:"type:uuid;uniqueIndex:idx_tenant_customer_credit;not null" json:"customer_id"`
	CreditLimit float64    `gorm:"type:decimal(12,2);default:0" json:"credit_limit"`
	Balance     float64    `gorm:"type:decimal(12,2);default:0" json:"balance"` // positive = amount owed by customer
	Status      string     `gorm:"size:20;default:'active'" json:"status"`      // active, suspended, blocked, closed
	DaysToRepay int        `gorm:"default:30" json:"days_to_repay"`             // standard repayment cycle in days
	LastDrawdownAt *time.Time `json:"last_drawdown_at,omitempty"`
	LastRepaymentAt *time.Time `json:"last_repayment_at,omitempty"`
	Notes       string     `gorm:"type:text" json:"notes,omitempty"`

	// Relations
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

// CreditTransaction records every draw-down, repayment, or credit adjustment.
type CreditTransaction struct {
	TenantScoped
	CreditAccountID uuid.UUID  `gorm:"type:uuid;index;not null" json:"credit_account_id"`
	CustomerID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"customer_id"`
	OrderID         *uuid.UUID `gorm:"type:uuid;index" json:"order_id,omitempty"`
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount"` // positive = drawdown (owes more), negative = repayment (owes less)
	BalanceAfter    float64    `gorm:"type:decimal(12,2);not null" json:"balance_after"`
	TransactionType string     `gorm:"size:30;not null" json:"transaction_type"` // drawdown, repayment, adjustment, write_off
	PaymentMethod   string     `gorm:"size:30" json:"payment_method,omitempty"`  // cash, momo, card, bank_transfer (for repayments)
	Reference       string     `gorm:"size:100;index" json:"reference,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	Status          string     `gorm:"size:20;default:'completed'" json:"status"` // completed, pending, overdue
	Notes           string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedByID     *uuid.UUID `gorm:"type:uuid" json:"created_by_id,omitempty"`
	CreatedBy       *User      `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	// Relations
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Order    *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// BNPLInstalment represents a scheduled instalment for split Buy-Now-Pay-Later purchases.
type BNPLInstalment struct {
	TenantScoped
	CreditTransactionID uuid.UUID  `gorm:"type:uuid;index;not null" json:"credit_transaction_id"`
	CustomerID          uuid.UUID  `gorm:"type:uuid;index;not null" json:"customer_id"`
	OrderID             *uuid.UUID `gorm:"type:uuid;index" json:"order_id,omitempty"`
	InstalmentNumber    int        `gorm:"not null" json:"instalment_number"`
	TotalInstalments    int        `gorm:"not null" json:"total_instalments"`
	Amount              float64    `gorm:"type:decimal(12,2);not null" json:"amount"`
	AmountPaid          float64    `gorm:"type:decimal(12,2);default:0" json:"amount_paid"`
	DueDate             time.Time  `gorm:"not null" json:"due_date"`
	Status              string     `gorm:"size:20;default:'pending'" json:"status"` // pending, partial, paid, overdue
	PaidAt              *time.Time `json:"paid_at,omitempty"`

	// Relations
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

// ReportSchedule configures automated End-of-Day, Weekly, and Monthly PDF/Email report dispatch.
type ReportSchedule struct {
	TenantScoped
	TenantID      uuid.UUID  `gorm:"type:uuid;index" json:"tenant_id"`
	ReportType    string     `gorm:"size:30;not null" json:"report_type"` // daily_z, weekly_pl, monthly_pl
	Recipients    string     `gorm:"type:text;not null" json:"recipients"` // comma-separated emails
	IsEnabled     bool       `gorm:"default:true" json:"is_enabled"`
	SendTime      string     `gorm:"size:10;default:'23:59'" json:"send_time"` // HH:MM in UTC or local
	IncludePDF    bool       `gorm:"default:true" json:"include_pdf"`
	LastSentAt    *time.Time `json:"last_sent_at,omitempty"`
	LastStatus    string     `gorm:"size:20;default:'ready'" json:"last_status"`
	LastError     string     `gorm:"type:text" json:"last_error,omitempty"`
}
