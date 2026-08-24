package models

import (
	"github.com/google/uuid"
)

// Notification represents a system notification for a user.
type Notification struct {
	TenantScoped
	UserID           uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Title            string    `gorm:"size:255;not null" json:"title"`
	Message          string    `gorm:"type:text;not null" json:"message"`
	Link             string    `gorm:"size:255" json:"link,omitempty"`
	IsRead           bool      `gorm:"default:false" json:"is_read"`
	NotificationType string    `gorm:"size:20;default:'info'" json:"notification_type"` // info, success, warning, error
	Category         string    `gorm:"size:20;default:'general'" json:"category"`       // general, inventory, sales, security, system

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// NotificationSetting holds per-user notification preferences.
type NotificationSetting struct {
	TenantScoped
	UserID             uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	EmailNotifications bool      `gorm:"default:true" json:"email_notifications"`
	LowStockAlerts     bool      `gorm:"default:true" json:"low_stock_alerts"`
	SalesReports       bool      `gorm:"default:true" json:"sales_reports"`
	SecurityAlerts     bool      `gorm:"default:true" json:"security_alerts"`
	SystemAlerts       bool      `gorm:"default:true" json:"system_alerts"`
}
