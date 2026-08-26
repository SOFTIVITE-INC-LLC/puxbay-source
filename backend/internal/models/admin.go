package models

import (
	"time"

	"github.com/google/uuid"
)

// Broadcast represents a system-wide announcement created by super admins.
type Broadcast struct {
	Base
	Title          string    `gorm:"size:255;not null" json:"title"`
	Message        string    `gorm:"type:text;not null" json:"message"`
	Type           string    `gorm:"size:50;default:'info'" json:"type"`           // info, success, alert
	TargetAudience string    `gorm:"size:50;default:'all'" json:"target_audience"` // all, active, trialing, past_due, suspended
	CreatedBy      uuid.UUID `gorm:"type:uuid" json:"created_by"`
}

type AdminRole struct {
	Base
	Name        string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Permissions string `gorm:"type:jsonb;default:'{}'" json:"permissions"`
}

type AdminUser struct {
	Base
	UserID      uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	AdminRoleID *uuid.UUID `gorm:"type:uuid" json:"admin_role_id"`
	Permissions string     `gorm:"type:jsonb;default:'{}'" json:"permissions"` // Per-user overrides/direct permissions

	User User       `gorm:"foreignKey:UserID" json:"-"`
	Role *AdminRole `gorm:"foreignKey:AdminRoleID" json:"-"`
}

type IPAllowlist struct {
	Base
	IPAddress   string `gorm:"size:45;uniqueIndex;not null" json:"ip_address"`
	Description string `gorm:"size:255" json:"description"`
}

type MasterAPIKey struct {
	Base
	Name      string     `gorm:"size:100;not null" json:"name"`
	Key       string     `gorm:"size:255;uniqueIndex;not null" json:"key"` // Hashed ideally, but for MVP plain/UUID
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	LastUsed  *time.Time `json:"last_used"`
	ExpiresAt *time.Time `json:"expires_at"`
}
