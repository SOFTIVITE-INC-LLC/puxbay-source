package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// APIKey represents a programmatic access token.
type APIKey struct {
	BranchScoped
	Name       string     `gorm:"size:100" json:"name"`
	KeyPrefix  string     `gorm:"size:8;index" json:"key_prefix"`
	KeyHash    string     `gorm:"size:128" json:"-"` // SHA-256
	IsActive   bool       `gorm:"default:true" json:"is_active"`
	IsSandbox  bool       `gorm:"default:false" json:"is_sandbox"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// ExternalSystem represents an OAuth/webhook app.
type ExternalSystem struct {
	Base
	DeveloperID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"developer_id"` // Links to Tenant
	Name             string         `gorm:"size:150;not null" json:"name"`
	Description      *string        `gorm:"type:text" json:"description,omitempty"`
	ClientID         uuid.UUID      `gorm:"type:uuid;uniqueIndex;default:gen_random_uuid()" json:"client_id"`
	ClientSecretHash string         `gorm:"size:128" json:"-"`
	RedirectURIs     datatypes.JSON `gorm:"type:jsonb" json:"redirect_uris,omitempty"`
	WebhookURL       *string        `gorm:"size:512" json:"webhook_url,omitempty"`
	Icon             string         `gorm:"size:50;default:'rocket_launch'" json:"icon"`
	IsPublic         bool           `gorm:"default:false" json:"is_public"`
	IsActive         bool           `gorm:"default:true" json:"is_active"`
}

// WebhookEndpoint is a destination for events.
type WebhookEndpoint struct {
	TenantScoped
	ExternalSystemID *uuid.UUID     `gorm:"type:uuid;index" json:"external_system_id,omitempty"`
	URL              string         `gorm:"size:512;not null" json:"url"`
	Secret           string         `gorm:"size:64;not null" json:"-"` // Used for HMAC signature
	IsActive         bool           `gorm:"default:true" json:"is_active"`
	Events           datatypes.JSON `gorm:"type:jsonb" json:"events"` // Array of event types
}

// WebhookEvent logs delivery attempts.
type WebhookEvent struct {
	Base
	EndpointID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"endpoint_id"`
	EventType    string         `gorm:"size:50;not null" json:"event_type"`
	Payload      datatypes.JSON `gorm:"type:jsonb" json:"payload"`
	Signature    *string        `gorm:"size:128" json:"signature,omitempty"`
	StatusCode   *uint          `json:"status_code,omitempty"`
	ResponseBody *string        `gorm:"type:text" json:"response_body,omitempty"`
	ErrorMessage *string        `gorm:"type:text" json:"error_message,omitempty"`
}
