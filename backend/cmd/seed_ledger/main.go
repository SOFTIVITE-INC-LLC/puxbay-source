package main

import (
	"log"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/database"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get all tenants
	var tenants []models.Tenant
	if err := db.Find(&tenants).Error; err != nil {
		log.Fatal("Failed to get tenants:", err)
	}

	defaultAccounts := []models.LedgerAccount{
		{Name: "Cash on Hand", Type: "Asset", Code: "1000", Description: "Physical cash on hand"},
		{Name: "Checking Account", Type: "Asset", Code: "1010", Description: "Primary bank account"},
		{Name: "Accounts Receivable", Type: "Asset", Code: "1200", Description: "Money owed by customers"},
		{Name: "Inventory Asset", Type: "Asset", Code: "1300", Description: "Value of inventory on hand"},

		{Name: "Accounts Payable", Type: "Liability", Code: "2000", Description: "Money owed to suppliers"},
		{Name: "Sales Tax Payable", Type: "Liability", Code: "2200", Description: "Tax collected and payable"},

		{Name: "Owner's Equity", Type: "Equity", Code: "3000", Description: "Owner's investment"},
		{Name: "Retained Earnings", Type: "Equity", Code: "3900", Description: "Accumulated profits"},

		{Name: "Sales Revenue", Type: "Revenue", Code: "4000", Description: "Revenue from product sales"},
		{Name: "Service Revenue", Type: "Revenue", Code: "4100", Description: "Revenue from services"},

		{Name: "Cost of Goods Sold", Type: "Expense", Code: "5000", Description: "Direct cost of products sold"},
		{Name: "Rent Expense", Type: "Expense", Code: "6000", Description: "Facility rent"},
		{Name: "Payroll Expense", Type: "Expense", Code: "6100", Description: "Employee salaries and wages"},
		{Name: "Utilities Expense", Type: "Expense", Code: "6200", Description: "Electricity, water, internet"},
	}

	for _, tenant := range tenants {
		log.Printf("Seeding ledger accounts for tenant: %s (Schema: %s)\n", tenant.Name, tenant.SchemaName)

		err := db.Transaction(func(tx *gorm.DB) error {
			// Set search path
			if err := tx.Exec("SET search_path TO " + tenant.SchemaName).Error; err != nil {
				return err
			}

			// Run AutoMigrate for tenant models so new tables (LedgerAccount, etc.) are created
			if err := models.MigrateTenantModels(tx); err != nil {
				return err
			}

			for _, acc := range defaultAccounts {
				var existing models.LedgerAccount
				if err := tx.Where("code = ?", acc.Code).First(&existing).Error; err != nil {
					if err := tx.Create(&acc).Error; err != nil {
						log.Printf("Failed to create account %s: %v\n", acc.Name, err)
					} else {
						log.Printf("Created account %s\n", acc.Name)
					}
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("Transaction failed for tenant %s: %v\n", tenant.Name, err)
		}
	}

	log.Println("Seeding complete!")
}
