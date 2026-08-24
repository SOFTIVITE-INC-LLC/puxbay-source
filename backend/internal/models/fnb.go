package models

import (
	"time"

	"github.com/google/uuid"
)

// DiningTable for F&B floor plans.
type DiningTable struct {
	BranchScoped
	Name      string  `gorm:"size:50;not null" json:"name"`
	Capacity  uint    `gorm:"default:4" json:"capacity"`
	Status    string  `gorm:"size:20;default:'available';index" json:"status"` // available, occupied, reserved, cleaning
	QRCodeURL *string `gorm:"size:512" json:"qr_code_url,omitempty"`
	PositionX int     `gorm:"default:0" json:"position_x"`
	PositionY int     `gorm:"default:0" json:"position_y"`
	IsActive  bool    `gorm:"default:true" json:"is_active"`
}

// KDSTicket represents a Kitchen Display System order.
type KDSTicket struct {
	BranchScoped
	OrderID      uuid.UUID    `gorm:"type:uuid;uniqueIndex;not null" json:"order_id"`
	Order        *Order       `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	TableID      *uuid.UUID   `gorm:"type:uuid;index" json:"table_id,omitempty"`
	Table        *DiningTable `gorm:"foreignKey:TableID" json:"table,omitempty"`
	Status       string       `gorm:"size:20;default:'pending';index" json:"status"` // pending, preparing, ready, served, cancelled
	KitchenNotes *string      `gorm:"type:text" json:"kitchen_notes,omitempty"`
	IsRush       bool         `gorm:"default:false" json:"is_rush"`
	StartedAt    *time.Time   `json:"started_at,omitempty"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
}

// SplitBillGroup links sub-orders from a single table.
type SplitBillGroup struct {
	BranchScoped
	TableID         *uuid.UUID   `gorm:"type:uuid" json:"table_id,omitempty"`
	Table           *DiningTable `gorm:"foreignKey:TableID" json:"table,omitempty"`
	OriginalOrderID *uuid.UUID   `gorm:"type:uuid" json:"original_order_id,omitempty"`
	OriginalOrder   *Order       `gorm:"foreignKey:OriginalOrderID" json:"original_order,omitempty"`
	Notes           *string      `gorm:"type:text" json:"notes,omitempty"`
}
