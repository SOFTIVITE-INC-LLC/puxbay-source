package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

// getDB retrieves the tenant-scoped database connection from the gin context.
// The TenantMiddleware sets "db" to a transaction pinned to the correct schema.
// Falls back to the handler's default DB if the context key is missing.
func getDB(c *gin.Context, fallback *gorm.DB) *gorm.DB {
	if db, exists := c.Get("db"); exists {
		if gormDB, ok := db.(*gorm.DB); ok {
			return gormDB
		}
	}
	return fallback
}

// auditAsync fires an audit log entry in a background goroutine so it never delays the HTTP response.
func auditAsync(db *gorm.DB, tenantID, userID uuid.UUID, action, resource, ip string, details map[string]interface{}) {
	go func() {
		var schemaName string
		if err := db.Table("public.tenants").Select("schema_name").Where("id = ?", tenantID).Scan(&schemaName).Error; err == nil && schemaName != "" {
			db.Transaction(func(tx *gorm.DB) error {
				tx.Exec(fmt.Sprintf("SET LOCAL search_path TO %s", pq.QuoteIdentifier(schemaName)))
				_ = services.NewAuditService(tx).LogAction(tenantID, userID, action, resource, details, ip)
				return nil
			})
		} else {
			_ = services.NewAuditService(db).LogAction(tenantID, userID, action, resource, details, ip)
		}
	}()
}
