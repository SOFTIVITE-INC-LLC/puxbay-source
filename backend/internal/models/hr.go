package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// PayrollPeriod represents a calendar month for payroll.
type PayrollPeriod struct {
	TenantScoped
	Name        string     `gorm:"size:100;not null" json:"name"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	IsProcessed bool       `gorm:"default:false" json:"is_processed"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// PayrollRecord is a staff member's financial snapshot.
type PayrollRecord struct {
	Base
	PeriodID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"period_id"`
	StaffID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"staff_id"`
	BaseSalarySnapshot float64    `gorm:"type:decimal(12,2);not null" json:"base_salary_snapshot"`
	TotalCommission    float64    `gorm:"type:decimal(12,2);default:0" json:"total_commission"`
	Bonus              float64    `gorm:"type:decimal(12,2);default:0" json:"bonus"`
	Deductions         float64    `gorm:"type:decimal(12,2);default:0" json:"deductions"`
	NetPay             float64    `gorm:"type:decimal(12,2);not null" json:"net_pay"`
	IsPaid             bool       `gorm:"default:false;index" json:"is_paid"`
	PaidAt             *time.Time `json:"paid_at,omitempty"`
	PaymentReference   *string    `gorm:"size:100" json:"payment_reference,omitempty"`
}

// LeaveRequest tracks staff time off.
type LeaveRequest struct {
	TenantScoped
	StaffID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"staff_id"`
	LeaveType    string     `gorm:"size:20;not null" json:"leave_type"` // annual, sick, maternity, unpaid, other
	StartDate    time.Time  `json:"start_date"`
	EndDate      time.Time  `json:"end_date"`
	Reason       *string    `gorm:"type:text" json:"reason,omitempty"`
	Status       string     `gorm:"size:20;default:'pending'" json:"status"` // pending, approved, rejected
	ReviewedByID *uuid.UUID `gorm:"type:uuid" json:"reviewed_by_id,omitempty"`
	ReviewedBy   *User      `gorm:"foreignKey:ReviewedByID" json:"reviewedby,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
}

// Attendance tracks staff clock-ins.
type Attendance struct {
	BranchScoped
	StaffID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"staff_id"`
	ClockIn  time.Time      `gorm:"index" json:"clock_in"`
	ClockOut *time.Time     `json:"clock_out,omitempty"`
	Metadata datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	Status   string         `gorm:"size:20;default:'present'" json:"status"` // present, on_leave, absent
}

// CommissionRule defines tiered commission logic for staff sales.
type CommissionRule struct {
	TenantScoped
	BranchID             *uuid.UUID `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	Branch               *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Name                 string     `gorm:"size:100;not null" json:"name"`
	MinSalesAmount       float64    `gorm:"type:decimal(12,2);default:0" json:"min_sales_amount"`
	CommissionPercentage float64    `gorm:"type:decimal(5,2);default:0" json:"commission_percentage"`
	FlatBonus            float64    `gorm:"type:decimal(10,2);default:0" json:"flat_bonus"`
	IsActive             bool       `gorm:"default:true" json:"is_active"`
}

// StaffAchievement tracks badges and milestones for staff members.
type StaffAchievement struct {
	Base
	StaffID     uuid.UUID `gorm:"type:uuid;not null;index" json:"staff_id"`
	BadgeName   string    `gorm:"size:100;not null" json:"badge_name"`
	BadgeIcon   string    `gorm:"size:50;default:'stars'" json:"badge_icon"` // Material icon name
	Description string    `gorm:"type:text" json:"description,omitempty"`
}

// ShiftSwapRequest handles employee-led shift exchange logic.
type ShiftSwapRequest struct {
	Base
	RequestingStaffID uuid.UUID  `gorm:"type:uuid;not null;index" json:"requesting_staff_id"`
	TargetStaffID     *uuid.UUID `gorm:"type:uuid;index" json:"target_staff_id,omitempty"`
	OriginalShiftID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"original_shift_id"`
	Status            string     `gorm:"size:20;default:'pending'" json:"status"` // pending, approved, rejected, completed
	Notes             string     `gorm:"type:text" json:"notes,omitempty"`
}

// StaffShift represents a scheduled work period for an employee.
type StaffShift struct {
	BranchScoped
	StaffID   uuid.UUID `gorm:"type:uuid;not null;index" json:"staff_id"`
	StartTime time.Time `gorm:"not null;index" json:"start_time"`
	EndTime   time.Time `gorm:"not null" json:"end_time"`
	Role      string    `gorm:"size:50" json:"role,omitempty"` // e.g. "Cashier", "Manager"
	Status    string    `gorm:"size:20;default:'scheduled'" json:"status"` // scheduled, in_progress, completed, missed
	Notes     string    `gorm:"type:text" json:"notes,omitempty"`
}
