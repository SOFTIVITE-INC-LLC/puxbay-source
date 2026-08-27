package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceToken stores push notification tokens for connected user devices.
// Supports web (Web Push VAPID), mobile (custom/platform-agnostic).
type DeviceToken struct {
	Base
	UserID   uuid.UUID `gorm:"type:uuid;not null;index"             json:"user_id"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index"             json:"tenant_id"`
	Token    string    `gorm:"type:text;not null;uniqueIndex"       json:"token"`
	Platform string    `gorm:"size:10;default:'web'"                json:"platform"` // web, android, ios
	IsActive bool      `gorm:"default:true"                         json:"is_active"`

	// Relations
	User   *User   `gorm:"foreignKey:UserID"   json:"user,omitempty"`
}

func (dt *DeviceToken) BeforeCreate(tx *gorm.DB) error {
	if dt.ID == uuid.Nil {
		dt.ID = uuid.New()
	}
	return nil
}
