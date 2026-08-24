package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AuditLog records business logic changes.
type AuditLog struct {
	Base
	TenantID  *uuid.UUID     `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	Tenant    *Tenant        `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	UserID    *uuid.UUID     `gorm:"type:uuid;index" json:"user_id,omitempty"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action    string         `gorm:"size:50;index" json:"action"` // create, update, delete, login, etc
	ModelName string         `gorm:"size:100" json:"model_name"`
	ObjectID  *string        `gorm:"size:100" json:"object_id,omitempty"`
	Changes   datatypes.JSON `gorm:"type:jsonb" json:"changes,omitempty"`
	IPAddress *string        `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent *string        `gorm:"type:text" json:"user_agent,omitempty"`
}

// APIRequestLog records API activity.
type APIRequestLog struct {
	Base
	TenantID       *uuid.UUID     `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	Tenant         *Tenant        `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	UserID         *uuid.UUID     `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Method         string         `gorm:"size:10" json:"method"`
	Endpoint       string         `gorm:"size:255;index" json:"endpoint"`
	StatusCode     uint           `gorm:"index" json:"status_code"`
	ResponseTimeMs uint           `json:"response_time_ms"`
	IPAddress      *string        `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent      *string        `gorm:"type:text" json:"user_agent,omitempty"`
	RequestBody    datatypes.JSON `gorm:"type:jsonb" json:"request_body,omitempty"`
	ResponseBody   datatypes.JSON `gorm:"type:jsonb" json:"response_body,omitempty"`
}

// HoneypotAttempt logs bot activity.
type HoneypotAttempt struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  *string   `gorm:"size:255" json:"username,omitempty"`
	Password  *string   `gorm:"size:255" json:"password,omitempty"`
	IPAddress *string   `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent *string   `gorm:"type:text" json:"user_agent,omitempty"`
	Path      string    `gorm:"size:255;default:'/admin/'" json:"path"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
}

// CrossTenantAuditLog logs superadmin access.
type CrossTenantAuditLog struct {
	Base
	UserID           *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	AccessedTenantID *uuid.UUID `gorm:"type:uuid;index" json:"accessed_tenant_id,omitempty"`
	UserHomeTenantID *uuid.UUID `gorm:"type:uuid" json:"user_home_tenant_id,omitempty"`
	ActionType       string     `gorm:"size:20;index" json:"action_type"`
	TargetModel      *string    `gorm:"size:100" json:"target_model,omitempty"`
	TargetObjectID   *string    `gorm:"size:100" json:"target_object_id,omitempty"`
	TargetObjectRepr *string    `gorm:"size:200" json:"target_object_repr,omitempty"`
	Description      *string    `gorm:"type:text" json:"description,omitempty"`
	IPAddress        *string    `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent        *string    `gorm:"type:text" json:"user_agent,omitempty"`
}

// ActivityLog logs detailed user actions.
type ActivityLog struct {
	Base
	TenantID       *uuid.UUID     `gorm:"type:uuid;index" json:"tenant_id"`
	Tenant         *Tenant        `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	ActorID        *uuid.UUID     `gorm:"type:uuid;index" json:"actor_id,omitempty"`
	Actor          *User          `gorm:"foreignKey:ActorID" json:"actor,omitempty"`
	ActionType     string         `gorm:"size:20" json:"action_type"`
	TargetModel    *string        `gorm:"size:100" json:"target_model,omitempty"`
	TargetObjectID *string        `gorm:"size:100" json:"target_object_id,omitempty"`
	Description    string         `gorm:"type:text" json:"description"`
	Changes        datatypes.JSON `gorm:"type:jsonb" json:"changes,omitempty"`
	IPAddress      *string        `gorm:"size:45" json:"ip_address,omitempty"`
}

// SystemLog logs system level messages.
type SystemLog struct {
	Base
	Level     string  `gorm:"size:10" json:"level"`
	Module    string  `gorm:"size:255" json:"module"`
	Message   string  `gorm:"type:text" json:"message"`
	Traceback *string `gorm:"type:text" json:"traceback,omitempty"`
	Path      *string `gorm:"size:255" json:"path,omitempty"`
}
