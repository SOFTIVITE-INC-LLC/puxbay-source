package models

import "github.com/google/uuid"

// DeliveryDriver represents a driver in the fleet.
type DeliveryDriver struct {
	TenantScoped
	Name          string  `gorm:"size:100;not null" json:"name"`
	Phone         string  `gorm:"size:20;not null" json:"phone"`
	VehicleInfo   string  `gorm:"size:100" json:"vehicle_info"`
	CurrentStatus string  `gorm:"size:20;default:'available'" json:"current_status"` // available, busy, offline
	Lat           float64 `gorm:"type:decimal(10,8)" json:"lat"`
	Lng           float64 `gorm:"type:decimal(11,8)" json:"lng"`
}

// DeliveryOrder tracks the delivery lifecycle of an order.
type DeliveryOrder struct {
	TenantScoped
	OrderID       uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"order_id"`
	DriverID      *uuid.UUID `gorm:"type:uuid;index" json:"driver_id,omitempty"`
	Status        string     `gorm:"size:20;default:'pending'" json:"status"` // pending, assigned, picked_up, delivered, failed
	TrackingLink  string     `gorm:"size:255;uniqueIndex" json:"tracking_link"`
	DeliveryNotes string     `gorm:"type:text" json:"delivery_notes"`
	DeliveryFee   float64    `gorm:"type:decimal(12,2);default:0" json:"delivery_fee"`

	// Relations
	Order  *Order          `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Driver *DeliveryDriver `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
}
