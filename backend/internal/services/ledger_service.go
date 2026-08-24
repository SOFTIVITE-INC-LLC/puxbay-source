package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type LedgerService struct {
	db *gorm.DB
}

func NewLedgerService(db *gorm.DB) *LedgerService {
	return &LedgerService{db: db}
}

// CreateJournalEntry creates a balanced double-entry transaction.
func (s *LedgerService) CreateJournalEntry(tenantID uuid.UUID, referenceID *uuid.UUID, refType, description string, lines []models.LedgerLine) error {
	// Verify debits equal credits
	var debits, credits float64
	for _, line := range lines {
		if line.IsDebit {
			debits += line.Amount
		} else {
			credits += line.Amount
		}
	}

	if debits != credits {
		return fmt.Errorf("journal entry unbalanced: debits (%.2f) != credits (%.2f)", debits, credits)
	}

	entry := models.JournalEntry{
		TenantScoped:  models.TenantScoped{Base: models.Base{ID: uuid.New()}},
		ReferenceID:   referenceID,
		ReferenceType: refType,
		Description:   description,
		Lines:         lines,
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := range entry.Lines {
			entry.Lines[i].ID = uuid.New()
			entry.Lines[i].JournalEntryID = entry.ID
		}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		return nil
	})
}

// SetupDefaultAccounts creates the standard chart of accounts for a new tenant.
func (s *LedgerService) SetupDefaultAccounts(tenantID uuid.UUID) error {
	accounts := []models.LedgerAccount{
		{TenantScoped: models.TenantScoped{Base: models.Base{ID: uuid.New()}}, Name: "Cash", Type: "Asset", Code: "1000"},
		{TenantScoped: models.TenantScoped{Base: models.Base{ID: uuid.New()}}, Name: "Accounts Receivable", Type: "Asset", Code: "1200"},
		{TenantScoped: models.TenantScoped{Base: models.Base{ID: uuid.New()}}, Name: "Inventory", Type: "Asset", Code: "1300"},
		{TenantScoped: models.TenantScoped{Base: models.Base{ID: uuid.New()}}, Name: "Accounts Payable", Type: "Liability", Code: "2000"},
		{TenantScoped: models.TenantScoped{Base: models.Base{ID: uuid.New()}}, Name: "Sales Tax Payable", Type: "Liability", Code: "2100"},
		{TenantScoped: models.TenantScoped{Base: models.Base{ID: uuid.New()}}, Name: "Sales Revenue", Type: "Revenue", Code: "4000"},
		{TenantScoped: models.TenantScoped{Base: models.Base{ID: uuid.New()}}, Name: "Cost of Goods Sold", Type: "Expense", Code: "5000"},
	}

	return s.db.Create(&accounts).Error
}
