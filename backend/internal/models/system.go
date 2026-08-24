package models

import (
	"time"
)

// DatabaseBackup tracks system backups.
type DatabaseBackup struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Filename  string    `gorm:"size:255;uniqueIndex" json:"filename"`
	FilePath  string    `gorm:"size:512" json:"file_path"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// FeatureFlag stores platform-wide toggles.
type FeatureFlag struct {
	Key       string    `gorm:"primaryKey;size:100" json:"key"`
	Value     bool      `gorm:"default:false" json:"value"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
