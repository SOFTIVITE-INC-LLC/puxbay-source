package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Promotion struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name           string         `gorm:"not null" json:"name"`
	Type           string         `gorm:"not null" json:"type"`          // e.g. "bogo", "bundle", "seasonal"
	Status         string         `gorm:"default:'draft'" json:"status"` // draft, active, paused, completed
	StartDate      time.Time      `json:"start_date"`
	EndDate        *time.Time     `json:"end_date"`
	Description    *string        `json:"description"`
	PointsRequired *int           `json:"points_required,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type DiscountCode struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Code           string         `gorm:"unique;not null" json:"code"`
	Type           string         `gorm:"not null" json:"type"` // e.g. "percentage", "fixed"
	Value          float64        `gorm:"not null" json:"value"`
	Status         string         `gorm:"default:'active'" json:"status"` // active, disabled, expired
	MaxUses        *int           `json:"max_uses"`
	CurrentUses    int            `gorm:"default:0" json:"current_uses"`
	ValidFrom      time.Time      `json:"valid_from"`
	ValidUntil     *time.Time     `json:"valid_until"`
	MinOrderValue  *float64       `json:"min_order_value"`
	PointsRequired *int           `json:"points_required,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
