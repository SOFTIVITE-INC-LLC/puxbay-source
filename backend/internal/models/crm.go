package models

import (
	"time"

	"github.com/google/uuid"
)

// CRMSettings configures points and loyalty.
type CRMSettings struct {
	TenantScoped
	PointsPerCurrency float64 `gorm:"type:decimal(5,2);default:1.00" json:"points_per_currency"`
	RedemptionRate    float64 `gorm:"type:decimal(5,2);default:0.01" json:"redemption_rate"`
	MonthlySalesGoal  float64 `gorm:"type:decimal(12,2);default:50000.00" json:"monthly_sales_goal"`
}

// CustomerSegment for targeted marketing.
type CustomerSegment struct {
	TenantScoped
	Name         string `gorm:"size:100;not null" json:"name"`
	Description  string `gorm:"type:text" json:"description,omitempty"`
	CriteriaJSON string `gorm:"type:jsonb" json:"criteria_json"` // JSON rules for the segment
}

// MarketingCampaign for Email/SMS blasts.
type MarketingCampaign struct {
	TenantScoped
	Name         string     `gorm:"size:100;not null" json:"name"`
	CampaignType string     `gorm:"size:10;default:'email'" json:"campaign_type"` // email, sms, both
	Subject      *string    `gorm:"size:255" json:"subject,omitempty"`
	Message      string     `gorm:"type:text;not null" json:"message"` // Encrypted
	CouponCode   *string    `gorm:"size:50" json:"coupon_code,omitempty"`
	Status       string     `gorm:"size:20;default:'draft'" json:"status"` // draft, scheduled, sent, cancelled
	TargetTierID *uuid.UUID `gorm:"type:uuid" json:"target_tier_id,omitempty"`
	SegmentID    *uuid.UUID `gorm:"type:uuid" json:"segment_id,omitempty"`

	IsAutomated  bool   `gorm:"default:false" json:"is_automated"`
	TriggerEvent string `gorm:"size:30;default:'manual'" json:"trigger_event"` // manual, birthday, abandoned_cart, etc.

	// Analytics
	OpenCount        int     `gorm:"default:0" json:"open_count"`
	ClickCount       int     `gorm:"default:0" json:"click_count"`
	ConversionCount  int     `gorm:"default:0" json:"conversion_count"`
	RevenueGenerated float64 `gorm:"type:decimal(12,2);default:0.00" json:"revenue_generated"`

	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
}

// CustomerFeedback ratings and reviews.
type CustomerFeedback struct {
	TenantScoped
	CustomerID uuid.UUID  `gorm:"type:uuid;not null;index" json:"customer_id"`
	OrderID    *uuid.UUID `gorm:"type:uuid" json:"order_id,omitempty"`
	Rating     uint       `gorm:"default:5" json:"rating"`            // 1-5
	Comment    *string    `gorm:"type:text" json:"comment,omitempty"` // Encrypted
	IsPublic   bool       `gorm:"default:false" json:"is_public"`
}

// CustomerCreditTransaction tracks credit purchases and payments received from customers.
type CustomerCreditTransaction struct {
	TenantScoped
	CustomerID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"customer_id"`
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount"` // positive = debt increase, negative = payment
	TransactionType string     `gorm:"size:20;not null" json:"transaction_type"`  // purchase | payment | adjustment
	Reference       string     `gorm:"size:100" json:"reference,omitempty"`
	Notes           string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedByID     *uuid.UUID `gorm:"type:uuid" json:"created_by_id,omitempty"`
	CreatedBy       *User      `gorm:"foreignKey:CreatedByID" json:"createdby,omitempty"`

	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

// SupplierCreditTransaction tracks credit purchases and payments to suppliers.
type SupplierCreditTransaction struct {
	TenantScoped
	SupplierID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"supplier_id"`
	Supplier        *Supplier  `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount"` // positive = balance increase, negative = payment
	TransactionType string     `gorm:"size:20;not null" json:"transaction_type"`  // purchase | payment | adjustment
	Reference       string     `gorm:"size:100" json:"reference,omitempty"`
	Notes           string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedByID     *uuid.UUID `gorm:"type:uuid" json:"created_by_id,omitempty"`
	CreatedBy       *User      `gorm:"foreignKey:CreatedByID" json:"createdby,omitempty"`
}

// ContactMessage stores messages from the contact form.
type ContactMessage struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:200" json:"name"`
	Email     string    `gorm:"size:255" json:"email"`
	Subject   string    `gorm:"size:200" json:"subject"`
	Message   string    `gorm:"type:text" json:"message"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// FeedbackReport stores bug reports and feature requests.
type FeedbackReport struct {
	TenantScoped
	UserID     *uuid.UUID `gorm:"type:uuid" json:"user_id,omitempty"`
	ReportType string     `gorm:"size:20;default:'bug'" json:"report_type"` // bug, recommendation, feature_request, other
	Priority   string     `gorm:"size:10;default:'medium'" json:"priority"` // low, medium, high, urgent
	Subject    string     `gorm:"size:200" json:"subject"`
	Message    string     `gorm:"type:text" json:"message"`
	Status     string     `gorm:"size:20;default:'new'" json:"status"` // new, in_progress, resolved, closed
	AdminNotes *string    `gorm:"type:text" json:"admin_notes,omitempty"`
}

// SupportTicket tracks a customer issue or helpdesk inquiry.
type SupportTicket struct {
	TenantScoped
	CustomerID  *uuid.UUID `gorm:"type:uuid;index" json:"customer_id,omitempty"` // Nullable if created by an anonymous user
	Customer    *Customer  `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Subject     string     `gorm:"size:255;not null" json:"subject"`
	Description string     `gorm:"type:text;not null" json:"description"`
	Status      string     `gorm:"size:20;default:'open'" json:"status"`   // open, in_progress, resolved, closed
	Priority    string     `gorm:"size:20;default:'medium'" json:"priority"` // low, medium, high, urgent
	AssignedTo  *uuid.UUID `gorm:"type:uuid;index" json:"assigned_to,omitempty"` // ID of the staff member
}

// TicketMessage stores replies to a SupportTicket.
type TicketMessage struct {
	Base
	TicketID uuid.UUID `gorm:"type:uuid;not null;index" json:"ticket_id"`
	SenderID uuid.UUID `gorm:"type:uuid;not null;index" json:"sender_id"` // Can be a customer or staff
	Message  string    `gorm:"type:text;not null" json:"message"`
	IsStaff  bool      `gorm:"default:false" json:"is_staff"`
}
