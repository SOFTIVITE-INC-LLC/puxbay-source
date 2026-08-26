package models

import (
	"time"

	"github.com/google/uuid"
)

// ServiceCategory groups services.
type ServiceCategory struct {
	TenantScoped
	Name  string `gorm:"size:100;not null" json:"name"`
	Icon  string `gorm:"size:50;default:'spa'" json:"icon"`
	Color string `gorm:"size:20;default:'purple'" json:"color"`
}

// Service represents a bookable service.
type Service struct {
	BranchScoped
	CategoryID      *uuid.UUID `gorm:"type:uuid;index" json:"category_id,omitempty"`
	Category        *Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Name            string     `gorm:"size:200;not null" json:"name"`
	Description     *string    `gorm:"type:text" json:"description,omitempty"`
	DurationMinutes uint       `gorm:"default:30" json:"duration_minutes"`
	Price           float64    `gorm:"type:decimal(10,2);not null" json:"price"`
	DefaultStaffID  *uuid.UUID `gorm:"type:uuid" json:"default_staff_id,omitempty"`
	Image           *string    `gorm:"size:512" json:"image,omitempty"`
	IsActive        bool       `gorm:"default:true" json:"is_active"`
}

// Appointment represents a booked service.
type Appointment struct {
	BranchScoped
	CustomerID    *uuid.UUID `gorm:"type:uuid;index" json:"customer_id,omitempty"`
	Customer      *Customer  `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	CustomerName  *string    `gorm:"size:200" json:"customer_name,omitempty"`
	CustomerPhone *string    `gorm:"size:30" json:"customer_phone,omitempty"`
	CustomerEmail *string    `gorm:"size:254" json:"customer_email,omitempty"`

	ServiceID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"service_id"`
	Service       *Service   `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	StaffMemberID *uuid.UUID `gorm:"type:uuid;index" json:"staff_member_id,omitempty"`

	StartTime time.Time `gorm:"index" json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	Status string  `gorm:"size:20;default:'scheduled';index" json:"status"`
	Notes  *string `gorm:"type:text" json:"notes,omitempty"`

	OrderID *uuid.UUID `gorm:"type:uuid" json:"order_id,omitempty"`
}

// ServiceCommissionRule sets commission rates.
type ServiceCommissionRule struct {
	TenantScoped
	StaffMemberID  uuid.UUID `gorm:"type:uuid;not null;index" json:"staff_member_id"`
	CommissionType string    `gorm:"size:20;default:'percentage'" json:"commission_type"`
	Value          float64   `gorm:"type:decimal(8,2);not null" json:"value"`
	AppliesTo      string    `gorm:"size:20;default:'all_sales'" json:"applies_to"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
}

// ServiceCommissionRecord logs calculated commissions.
type ServiceCommissionRecord struct {
	TenantScoped
	StaffMemberID uuid.UUID  `gorm:"type:uuid;not null;index" json:"staff_member_id"`
	RuleID        *uuid.UUID `gorm:"type:uuid" json:"rule_id,omitempty"`
	OrderID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"order_id"`
	Amount        float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	IsPaid        bool       `gorm:"default:false;index" json:"is_paid"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	Notes         *string    `gorm:"type:text" json:"notes,omitempty"`
}
