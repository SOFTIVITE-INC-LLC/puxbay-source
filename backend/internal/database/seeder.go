package database

import (
	"log"

	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

// SeedRBAC populates the default permissions and roles if they don't exist.
func SeedRBAC(db *gorm.DB) error {
	permissions := []models.Permission{
		{Code: "pos:sale", Description: "Perform POS sales", Category: "pos"},
		{Code: "pos:refund", Description: "Process returns and refunds", Category: "pos"},
		{Code: "pos:discount", Description: "Apply manual discounts", Category: "pos"},
		{Code: "pos:override", Description: "Manager override (voids, restricted items)", Category: "pos"},
		{Code: "cashdrawer:manage", Description: "Manage cash drawer sessions", Category: "pos"},

		{Code: "orders:read", Description: "View orders", Category: "orders"},
		{Code: "orders:create", Description: "Create orders", Category: "orders"},
		{Code: "orders:update", Description: "Update orders", Category: "orders"},
		{Code: "orders:void", Description: "Void orders", Category: "orders"},
		{Code: "orders:delete", Description: "Delete orders", Category: "orders"},
		{Code: "delivery:manage", Description: "Manage dispatch and delivery tracking", Category: "orders"},
		{Code: "omnichannel:manage", Description: "Manage omnichannel sales channels", Category: "orders"},

		{Code: "inventory:read", Description: "View inventory levels", Category: "inventory"},
		{Code: "inventory:receive", Description: "Receive purchase orders", Category: "inventory"},
		{Code: "inventory:transfer", Description: "Transfer stock between branches", Category: "inventory"},
		{Code: "inventory:stocktake", Description: "Perform stocktakes", Category: "inventory"},
		{Code: "inventory:manage", Description: "General inventory management", Category: "inventory"},

		{Code: "products:read", Description: "View products catalog", Category: "catalog"},
		{Code: "products:create", Description: "Create new products", Category: "catalog"},
		{Code: "products:update", Description: "Update existing products", Category: "catalog"},
		{Code: "products:delete", Description: "Delete products", Category: "catalog"},
		{Code: "products:manage", Description: "General products management", Category: "catalog"},
		{Code: "categories:manage", Description: "Manage product categories", Category: "catalog"},
		{Code: "barcode:manage", Description: "Print barcodes and labels", Category: "catalog"},
		{Code: "services:manage", Description: "Manage service-based products", Category: "catalog"},

		{Code: "customers:read", Description: "View customers", Category: "crm"},
		{Code: "customers:write", Description: "Create/edit customers", Category: "crm"},
		{Code: "customers:delete", Description: "Delete customers", Category: "crm"},
		{Code: "customers:manage", Description: "General customer management", Category: "crm"},
		{Code: "loyalty:manage", Description: "Manage loyalty points and programs", Category: "crm"},
		{Code: "crm:manage", Description: "Manage CRM, tickets, and feedback", Category: "crm"},

		{Code: "finance:read", Description: "View financials and P&L", Category: "financial"},
		{Code: "finance:write", Description: "Record expenses and income", Category: "financial"},
		{Code: "finance:manage", Description: "General finance management", Category: "financial"},
		{Code: "giftcards:manage", Description: "Manage gift cards", Category: "financial"},
		{Code: "payment_methods:manage", Description: "Configure payment methods", Category: "financial"},
		{Code: "wallet:manage", Description: "Manage store wallet and balance", Category: "financial"},

		{Code: "reports:read", Description: "View standard reports", Category: "reports"},
		{Code: "analytics:read", Description: "View advanced analytics and intelligence", Category: "reports"},
		{Code: "export:manage", Description: "Export system data", Category: "reports"},

		{Code: "billing:read", Description: "View subscription and billing", Category: "billing"},
		{Code: "billing:update", Description: "Update subscription", Category: "billing"},
		{Code: "billing:manage", Description: "General billing management", Category: "billing"},

		{Code: "staff:read", Description: "View staff accounts", Category: "staff"},
		{Code: "staff:write", Description: "Create and edit staff accounts", Category: "staff"},
		{Code: "staff:delete", Description: "Delete staff accounts", Category: "staff"},
		{Code: "staff:manage", Description: "General staff management", Category: "staff"},
		{Code: "shifts:manage", Description: "Manage shifts and timeclocks", Category: "staff"},

		{Code: "hr:read", Description: "View HR attendance and payroll", Category: "hr"},
		{Code: "hr:write", Description: "Modify HR records", Category: "hr"},
		{Code: "hr:manage", Description: "General HR management", Category: "hr"},
		{Code: "payroll:manage", Description: "Manage payroll", Category: "hr"},

		{Code: "roles:read", Description: "View roles and permissions", Category: "security"},
		{Code: "roles:write", Description: "Create and update roles", Category: "security"},
		{Code: "roles:delete", Description: "Delete roles", Category: "security"},

		{Code: "settings:read", Description: "View system settings", Category: "settings"},
		{Code: "settings:manage", Description: "Manage system settings", Category: "settings"},
		{Code: "branches:manage", Description: "Manage store branches", Category: "settings"},
		{Code: "integrations:manage", Description: "Manage third-party integrations", Category: "settings"},
		{Code: "security:manage", Description: "Manage security and privacy", Category: "settings"},
		{Code: "taxes:manage", Description: "Manage tax rates", Category: "settings"},

		{Code: "marketing:manage", Description: "Manage marketing campaigns and discounts", Category: "marketing"},
		{Code: "promotions:manage", Description: "Manage coupons and promotions", Category: "marketing"},
		{Code: "campaigns:manage", Description: "Manage email and SMS campaigns", Category: "marketing"},
		{Code: "content:manage", Description: "Manage storefront content", Category: "marketing"},

		{Code: "storefront:manage", Description: "Configure e-commerce storefront", Category: "ecommerce"},

		{Code: "fnb:manage", Description: "General Food & Beverage (F&B) operations", Category: "fnb"},
		{Code: "fnb:tables:manage", Description: "Manage floor plans and tables", Category: "fnb"},
		{Code: "fnb:kitchen:manage", Description: "Manage Kitchen Display System (KDS)", Category: "fnb"},

		{Code: "b2b:manage", Description: "Manage wholesale and B2B customers", Category: "b2b"},
		{Code: "suppliers:manage", Description: "Manage suppliers and catalogs", Category: "catalog"},

		{Code: "kiosk:manage", Description: "Manage self-service kiosks", Category: "pos"},
		{Code: "notifications:manage", Description: "Manage global notifications", Category: "settings"},
		{Code: "webhooks:manage", Description: "Manage API webhooks", Category: "settings"},
		{Code: "copilot:manage", Description: "Manage AI copilot settings", Category: "settings"},
		{Code: "sync:manage", Description: "Manage offline sync", Category: "settings"},
	}

	for i := range permissions {
		db.FirstOrCreate(&permissions[i], models.Permission{Code: permissions[i].Code})
	}

	// Create System Roles
	adminRole := models.Role{Name: "Admin", IsSystem: true, Description: "Full access"}
	managerRole := models.Role{Name: "Manager", IsSystem: true, Description: "Manager access"}
	supervisorRole := models.Role{Name: "Supervisor", IsSystem: true, Description: "Floor supervisor access"}
	cashierRole := models.Role{Name: "Cashier", IsSystem: true, Description: "Basic POS access"}
	inventoryRole := models.Role{Name: "Inventory Manager", IsSystem: true, Description: "Stock and supply chain access"}
	accountantRole := models.Role{Name: "Accountant", IsSystem: true, Description: "Financial and billing access"}
	hrRole := models.Role{Name: "HR Admin", IsSystem: true, Description: "Staff and payroll management"}

	db.FirstOrCreate(&adminRole, models.Role{Name: "Admin"})
	db.FirstOrCreate(&managerRole, models.Role{Name: "Manager"})
	db.FirstOrCreate(&supervisorRole, models.Role{Name: "Supervisor"})
	db.FirstOrCreate(&cashierRole, models.Role{Name: "Cashier"})
	db.FirstOrCreate(&inventoryRole, models.Role{Name: "Inventory Manager"})
	db.FirstOrCreate(&accountantRole, models.Role{Name: "Accountant"})
	db.FirstOrCreate(&hrRole, models.Role{Name: "HR Admin"})

	// Assign permissions to Admin
	db.Model(&adminRole).Association("Permissions").Replace(permissions)

	// Assign permissions to Cashier
	var cashierPerms []models.Permission
	cashierCodes := []string{
		"pos:sale", "pos:refund", "cashdrawer:manage",
		"orders:read", "orders:create",
		"inventory:read",
		"customers:read", "customers:write",
		"shifts:manage",
	}
	db.Where("code IN ?", cashierCodes).Find(&cashierPerms)
	db.Model(&cashierRole).Association("Permissions").Replace(cashierPerms)

	// Assign permissions to Manager
	var managerPerms []models.Permission
	managerCodes := []string{
		"pos:sale", "pos:refund", "pos:discount", "pos:override", "cashdrawer:manage", "returns:manage", "kiosk:manage",
		"orders:read", "orders:create", "orders:update", "orders:void", "orders:delete", "delivery:manage", "omnichannel:manage",
		"inventory:read", "inventory:receive", "inventory:transfer", "inventory:stocktake", "inventory:manage",
		"products:read", "products:create", "products:update", "products:delete", "products:manage",
		"categories:manage", "barcode:manage", "services:manage",
		"customers:read", "customers:write", "customers:delete", "customers:manage", "loyalty:manage", "crm:manage",
		"finance:read", "finance:write",
		"reports:read", "analytics:read",
		"staff:read", "staff:write", "shifts:manage",
		"hr:read", "hr:write",
		"marketing:manage", "promotions:manage", "campaigns:manage", "content:manage",
		"fnb:manage", "fnb:tables:manage", "fnb:kitchen:manage",
		"b2b:manage", "suppliers:manage",
	}
	db.Where("code IN ?", managerCodes).Find(&managerPerms)
	db.Model(&managerRole).Association("Permissions").Replace(managerPerms)

	// Assign permissions to Supervisor
	var supervisorPerms []models.Permission
	supervisorCodes := []string{
		"pos:sale", "pos:refund", "pos:discount", "pos:override", "cashdrawer:manage", "returns:manage",
		"orders:read", "orders:create", "orders:update", "orders:void",
		"inventory:read", "inventory:stocktake",
		"customers:read", "customers:write", "loyalty:manage",
		"reports:read",
		"shifts:manage",
	}
	db.Where("code IN ?", supervisorCodes).Find(&supervisorPerms)
	db.Model(&supervisorRole).Association("Permissions").Replace(supervisorPerms)

	// Assign permissions to Inventory Manager
	var inventoryPerms []models.Permission
	inventoryCodes := []string{
		"inventory:read", "inventory:receive", "inventory:transfer", "inventory:stocktake", "inventory:manage",
		"products:read", "products:create", "products:update", "products:delete", "products:manage",
		"categories:manage", "barcode:manage", "suppliers:manage",
		"reports:read",
	}
	db.Where("code IN ?", inventoryCodes).Find(&inventoryPerms)
	db.Model(&inventoryRole).Association("Permissions").Replace(inventoryPerms)

	// Assign permissions to Accountant
	var accountantPerms []models.Permission
	accountantCodes := []string{
		"finance:read", "finance:write", "finance:manage",
		"billing:read",
		"reports:read", "analytics:read", "export:manage",
		"taxes:manage",
	}
	db.Where("code IN ?", accountantCodes).Find(&accountantPerms)
	db.Model(&accountantRole).Association("Permissions").Replace(accountantPerms)

	// Assign permissions to HR Admin
	var hrPerms []models.Permission
	hrCodes := []string{
		"staff:read", "staff:write", "staff:manage", "shifts:manage",
		"hr:read", "hr:write", "hr:manage", "payroll:manage",
		"reports:read",
	}
	db.Where("code IN ?", hrCodes).Find(&hrPerms)
	db.Model(&hrRole).Association("Permissions").Replace(hrPerms)

	log.Println("🌱 Seeding default Plans...")
	plans := []models.Plan{
		{Name: "Free", Description: "Basic features for small businesses", Price: 29.99, Interval: "monthly", MaxBranches: 1, MaxUsers: 1, IsActive: true},
		{Name: "Pro", Description: "Advanced features for growing businesses", Price: 69.99, Interval: "monthly", MaxBranches: 3, MaxUsers: 5, IsActive: true},
		{Name: "Enterprise", Description: "Unlimited everything", Price: 129.99, Interval: "monthly", MaxBranches: 10, MaxUsers: 20, IsActive: true},
	}
	for i := range plans {
		db.FirstOrCreate(&plans[i], models.Plan{Name: plans[i].Name})
	}

	log.Println("🌱 Seeding default Feature Flags...")
	flags := []models.FeatureFlag{
		{Key: "enable_copilot", Value: true},
		{Key: "enable_offline_sync", Value: true},
		{Key: "enable_whatsapp_integration", Value: false},
		{Key: "enable_loyalty", Value: true},
	}
	for i := range flags {
		db.FirstOrCreate(&flags[i], models.FeatureFlag{Key: flags[i].Key})
	}

	return nil
}
