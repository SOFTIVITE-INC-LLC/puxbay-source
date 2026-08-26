package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

const ContextKeyAdminPermissions = "admin_permissions"
const ContextKeyIsSuperuser = "is_superuser"

// AdminAuthMiddleware validates the JWT for admin routes, then fetches and merges
// the user's AdminRole permissions + direct AdminUser permissions into context.
// Superusers bypass all permission checks.
func AdminAuthMiddleware(db *gorm.DB, authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Reuse standard JWT validation
		AuthMiddleware(authService)(c)
		if c.IsAborted() {
			return
		}

		// Get the user ID from context (set by AuthMiddleware)
		userIDVal, exists := c.Get(ContextKeyUserID)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
			return
		}

		// Check if the user is a superuser
		var user models.User
		if err := db.Select("id, is_superuser").Where("id = ?", userID).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		c.Set(ContextKeyIsSuperuser, user.IsSuperuser)

		if user.IsSuperuser {
			// Superusers have all permissions — set a wildcard sentinel
			c.Set(ContextKeyAdminPermissions, []string{"*"})
			c.Next()
			return
		}

		// Fetch the AdminUser record (role + direct permissions)
		var adminUser models.AdminUser
		if err := db.Preload("Role").Where("user_id = ?", userID).First(&adminUser).Error; err != nil {
			// Not in admin_users table — deny access
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access not granted for this account"})
			return
		}

		// Merge permissions: direct permissions + role permissions
		merged := make(map[string]bool)

		// 1. Direct user permissions
		if adminUser.Permissions != "" && adminUser.Permissions != "{}" {
			var directPerms []string
			if err := json.Unmarshal([]byte(adminUser.Permissions), &directPerms); err == nil {
				for _, p := range directPerms {
					merged[p] = true
				}
			}
		}

		// 2. Role permissions
		if adminUser.Role != nil && adminUser.Role.Permissions != "" {
			var rolePerms []string
			if err := json.Unmarshal([]byte(adminUser.Role.Permissions), &rolePerms); err == nil {
				for _, p := range rolePerms {
					merged[p] = true
				}
			}
		}

		// Convert to slice
		perms := make([]string, 0, len(merged))
		for p := range merged {
			perms = append(perms, p)
		}

		c.Set(ContextKeyAdminPermissions, perms)
		c.Next()
	}
}

// RequireAdminPermission gates an admin route behind a specific permission.
// Superusers (wildcard "*") always pass.
func RequireAdminPermission(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permsVal, exists := c.Get(ContextKeyAdminPermissions)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No admin permissions found"})
			return
		}

		perms, ok := permsVal.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Malformed permissions"})
			return
		}

		for _, p := range perms {
			if p == "*" || p == required {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":    "Insufficient permissions",
			"required": required,
		})
	}
}
