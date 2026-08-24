package services

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type PrivacyService struct {
	db *gorm.DB
}

func NewPrivacyService(db *gorm.DB) *PrivacyService {
	return &PrivacyService{db: db}
}

func (s *PrivacyService) ExportData(tenantID uuid.UUID) error {
	var customers []models.Customer
	var orders []models.Order
	var products []models.Product

	if err := s.db.Find(&customers).Error; err != nil {
		return err
	}
	if err := s.db.Find(&orders).Error; err != nil {
		return err
	}
	if err := s.db.Find(&products).Error; err != nil {
		return err
	}

	exportData := map[string]interface{}{
		"customers":   customers,
		"orders":      orders,
		"products":    products,
		"exported_at": time.Now(),
		"tenant_id":   tenantID,
	}

	file, err := os.CreateTemp("", fmt.Sprintf("gdpr_export_%s_*.zip", tenantID))
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	f, err := zipWriter.Create("export.json")
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(exportData); err != nil {
		return err
	}

	// Wait for zip writer to flush
	zipWriter.Flush()

	// Real app: trigger async email here
	return nil
}

func (s *PrivacyService) DeleteAccount(tenantID uuid.UUID, reason string) error {
	// Soft delete the tenant account
	return s.db.Model(&models.Tenant{}).Where("id = ?", tenantID).Update("deleted_at", time.Now()).Error
}

func (s *PrivacyService) AnonymizeCustomer(tenantID uuid.UUID, customerID string) error {
	// Replaces PII with anonymous data, but keeps orders intact for reporting
	return s.db.Transaction(func(tx *gorm.DB) error {
		var customer map[string]interface{}
		if err := tx.Table("customers").Where("id = ? AND tenant_id = ?", customerID, tenantID).First(&customer).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"first_name":   "Anonymized",
			"last_name":    "User",
			"email":        customerID + "@anonymized.local",
			"phone":        nil,
			"address_line": nil,
		}

		return tx.Table("customers").Where("id = ? AND tenant_id = ?", customerID, tenantID).Updates(updates).Error
	})
}
