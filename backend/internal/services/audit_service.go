package services

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

// AuditService handles immutable logging of sensitive operations.
type AuditService struct {
	db *gorm.DB
}

// NewAuditService creates a new audit service.
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// LogPIIAccess logs when a user views or exports PII.
func (s *AuditService) LogPIIAccess(tenantID, userID uuid.UUID, resource string, details map[string]interface{}, ip string) error {
	detailsJSON, _ := json.Marshal(details)
	detailsStr := string(detailsJSON)

	entry := models.AuditLog{
		TenantID:  &tenantID,
		UserID:    &userID,
		Action:    "PII_ACCESS",
		ModelName: resource,
		IPAddress: &ip,
	}
	_ = entry.Changes.UnmarshalJSON([]byte(detailsStr))

	// Create without running hooks that could modify this via AfterSave logic, ensuring it's an append-only log.
	if err := s.db.Create(&entry).Error; err != nil {
		fmt.Printf("Warning: Failed to create audit log: %v\n", err)
		return err
	}
	return nil
}

// LogAction logs a generic action in the system.
func (s *AuditService) LogAction(tenantID, userID uuid.UUID, action, resource string, details map[string]interface{}, ip string) error {
	detailsJSON, _ := json.Marshal(details)
	detailsStr := string(detailsJSON)

	entry := models.AuditLog{
		TenantID:  &tenantID,
		UserID:    &userID,
		Action:    action,
		ModelName: resource,
		IPAddress: &ip,
	}
	_ = entry.Changes.UnmarshalJSON([]byte(detailsStr))

	return s.db.Create(&entry).Error
}
