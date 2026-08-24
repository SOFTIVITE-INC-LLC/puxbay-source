package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/softivite/puxbay/internal/database"
	"gorm.io/gorm"
)

// extractSubdomain tries to extract the tenant subdomain from the request.
// It prioritizes the X-Tenant-Subdomain header (useful for proxy setups),
// and falls back to parsing the Host header.
func extractSubdomain(c *gin.Context) string {
	if headerSubdomain := c.GetHeader("X-Tenant-Subdomain"); headerSubdomain != "" {
		return headerSubdomain
	}

	if querySubdomain := c.Query("tenant"); querySubdomain != "" {
		return querySubdomain
	}

	host := c.Request.Host
	host = strings.Split(host, ":")[0]

	// Explicitly handle puxbay.com and localhost
	if strings.HasSuffix(host, ".puxbay.com") {
		sub := strings.TrimSuffix(host, ".puxbay.com")
		sub = strings.TrimPrefix(sub, "api.")
		sub = strings.TrimPrefix(sub, "www.")
		if sub != "" && sub != "api" && sub != "www" {
			return sub
		}
	} else if strings.HasSuffix(host, ".localhost") {
		sub := strings.TrimSuffix(host, ".localhost")
		sub = strings.TrimPrefix(sub, "api.")
		sub = strings.TrimPrefix(sub, "www.")
		if sub != "" && sub != "api" && sub != "www" {
			return sub
		}
	}

	return ""
}

// RequireTenantMiddleware ensures that a tenant has been successfully resolved.
// Should be used AFTER TenantMiddleware for routes that absolutely require a tenant.
func RequireTenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get(ContextKeyTenantID); !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant required for this operation"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// TenantMiddleware sets the PostgreSQL Schema for the current request based on the subdomain.
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		subdomain := extractSubdomain(c)

		if subdomain == "" {
			// No subdomain found, fallback to public schema or reject
			// For some public routes or superadmin routes, public schema is fine.
			c.Set("db", database.DB)
			c.Next()
			return
		}

		// Look up the tenant by subdomain
		var tenant struct {
			ID         uuid.UUID
			SchemaName string
		}

		if err := database.DB.Table("public.tenants").Select("id, schema_name").Where("subdomain = ?", subdomain).Scan(&tenant).Error; err != nil || tenant.SchemaName == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found for subdomain: " + subdomain})
			c.Abort()
			return
		}

		// Cross-validate with JWT if authenticated
		if jwtTenantIDVal, exists := c.Get(ContextKeyTenantID); exists {
			jwtTenantID, ok := jwtTenantIDVal.(uuid.UUID)
			if !ok || jwtTenantID != tenant.ID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Cross-tenant access is strictly forbidden"})
				c.Abort()
				return
			}
		} else {
			// Set the tenant ID in context for public routes that rely on it
			c.Set(ContextKeyTenantID, tenant.ID)
		}

		// Start a transaction to pin a database connection for this request
		// This is required to safely use PostgreSQL search_path with a connection pool.
		// Handlers can still use nested transactions (SAVEPOINTs).
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			// Set search_path for this transaction ONLY
			tx.Exec(fmt.Sprintf("SET LOCAL search_path TO %s", pq.QuoteIdentifier(tenant.SchemaName)))

			// Inject the schema name into the GORM context for audit callbacks
			txCtx := context.WithValue(tx.Statement.Context, database.TenantSchemaKey, tenant.SchemaName)
			// Also inject tenant_id and user_id for audit logs
			txCtx = context.WithValue(txCtx, "tenant_id", tenant.ID)
			if userID, exists := c.Get(ContextKeyUserID); exists {
				txCtx = context.WithValue(txCtx, "user_id", userID)
			}
			tx = tx.WithContext(txCtx)

			// Inject the transaction-scoped DB into the context
			c.Set("db", tx)

			// Continue processing the request
			c.Next()

			// If the request had errors or a bad status, we should rollback
			if len(c.Errors) > 0 || c.Writer.Status() >= 400 {
				return fmt.Errorf("request failed with status %d", c.Writer.Status())
			}
			return nil
		})

		if err != nil && c.Writer.Status() < 400 {
			// If it failed and we haven't sent a response yet
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database transaction failed"})
		}
	}
}
