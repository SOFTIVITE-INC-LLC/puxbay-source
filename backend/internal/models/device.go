package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Device struct {
	TenantScoped
	BranchID   *uuid.UUID     `json:"branch_id" gorm:"type:uuid;index"`
	Branch     *Branch        `json:"branch,omitempty" gorm:"foreignKey:BranchID"`
	Name       string         `json:"name" gorm:"type:varchar(255);not null"`
	DeviceType string         `json:"device_type" gorm:"type:varchar(50);not null"` // e.g., 'printer', 'payment_terminal', 'cash_drawer'
	IPAddress  string         `json:"ip_address" gorm:"type:varchar(255)"`
	MACAddress string         `json:"mac_address" gorm:"type:varchar(255)"`
	Status     string         `json:"status" gorm:"type:varchar(50);default:'offline'"` // online, offline
	Config     datatypes.JSON `json:"config" gorm:"type:jsonb"`                         // Store integration-specific configs like Paystack terminal ID
	LastSeenAt *time.Time     `json:"last_seen_at" gorm:"index"`
}
