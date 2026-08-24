package models

import (
	"time"

	"github.com/google/uuid"
)

// CashRegisterShift tracks opening and closing of a till.
type CashRegisterShift struct {
	Base
	BranchID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"branch_id"`
	Branch       *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	OpenedByID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"opened_by_id"`
	OpenedBy     *User      `gorm:"foreignKey:OpenedByID" json:"openedby,omitempty"`
	ClosedByID   *uuid.UUID `gorm:"type:uuid;index" json:"closed_by_id,omitempty"`
	ClosedBy     *User      `gorm:"foreignKey:ClosedByID" json:"closedby,omitempty"`
	OpenedAt     time.Time  `json:"opened_at"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	OpeningFloat float64    `gorm:"type:decimal(12,2);not null" json:"opening_float"`
	ExpectedCash float64    `gorm:"type:decimal(12,2);default:0" json:"expected_cash"`
	ActualCash   float64    `gorm:"type:decimal(12,2);default:0" json:"actual_cash"`
	Variance     float64    `gorm:"type:decimal(12,2);default:0" json:"variance"`
	Status       string     `gorm:"size:20;default:'open';index" json:"status"` // open, closed
	Notes        string     `gorm:"type:text" json:"notes"`
}
