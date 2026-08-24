package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission defines a specific action that can be performed in the system.
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code        string    `gorm:"size:100;uniqueIndex;not null" json:"code"` // e.g., "orders:void"
	Description string    `gorm:"size:255" json:"description"`
	Category    string    `gorm:"size:100" json:"category"` // e.g., "inventory", "sales"
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Role represents a collection of permissions.
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    *uuid.UUID     `gorm:"type:uuid;index" json:"tenant_id,omitempty"` // Null for system default roles
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"` // If true, cannot be deleted/modified by users
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Permissions []Permission `gorm:"many2many:public.role_permissions;" json:"permissions,omitempty"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (Role) TableName() string {
	return "public.roles"
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (Permission) TableName() string {
	return "public.permissions"
}
