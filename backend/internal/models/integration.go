package models

import (
	"time"

	"gorm.io/datatypes"
)

// TenantIntegration stores OAuth tokens and configuration for external apps like Xero or QuickBooks.
type TenantIntegration struct {
	TenantScoped
	Provider     string         `gorm:"size:50;not null;uniqueIndex:idx_tenant_provider" json:"provider"` // e.g. "xero", "quickbooks"
	AccessToken  string         `gorm:"type:text" json:"-"`                                               // Should be encrypted
	RefreshToken string         `gorm:"type:text" json:"-"`                                               // Should be encrypted
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"`
	Config       datatypes.JSON `gorm:"type:jsonb" json:"config,omitempty"` // E.g. account mappings, tax codes
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastSyncAt   *time.Time     `json:"last_sync_at,omitempty"`
}
