package admin

import "github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"

// CreateGenerators maps table names to GoAdmin table generation functions.
func CreateGenerators() table.GeneratorList {
	return map[string]table.Generator{
		// Generate tables here using GoAdmin's adm cli tool later
		// "tenants": GetTenantsTable,
		// "users": GetUsersTable,
	}
}
