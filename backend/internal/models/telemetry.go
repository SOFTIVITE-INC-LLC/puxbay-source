package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TelemetryLog represents a persisted OpenTelemetry trace span.
type TelemetryLog struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TraceID   string         `gorm:"index;not null" json:"trace_id"`
	SpanID    string         `gorm:"index;not null" json:"span_id"`
	Name      string         `gorm:"index;not null" json:"name"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
	Duration  float64        `json:"duration_ms"` // Duration in milliseconds
	Status    string         `gorm:"index" json:"status"` // e.g. Unset, Ok, Error
	
	// Attributes stores the key-value pairs of the span
	Attributes datatypes.JSON `json:"attributes"`
	// Events stores any log events attached to the span
	Events     datatypes.JSON `json:"events"`
	
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (t *TelemetryLog) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (TelemetryLog) TableName() string {
	return "public.telemetry_logs"
}
