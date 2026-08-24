package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base contains common fields for all models.
// Provides UUID primary key and automatic timestamps.
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Version   int64          `gorm:"default:1;not null;index" json:"version"` // Logical clock for CRDT Sync
}

// BeforeCreate generates a UUID if one isn't set.
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// TenantScoped is a base for models that belong to a tenant.
// With schema isolation, the schema itself defines the tenant context,
// so a TenantID column is no longer needed on every table.
type TenantScoped struct {
	Base
}

// BranchScoped is a base for models scoped to both tenant and branch.
type BranchScoped struct {
	TenantScoped
	BranchID *uuid.UUID `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	Branch   *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}
