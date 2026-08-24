package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RequirePermissionOrOverride allows access if the user has the required permission,
// OR if a manager at the same branch provides a valid POS override PIN via the X-Manager-Override-PIN header.
func RequirePermissionOrOverride(db *gorm.DB, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Superadmin bypass
		role, exists := c.Get(ContextKeyRole)
		if exists && role.(string) == "superadmin" {
			c.Next()
			return
		}

		// 2. Check if current user has permission naturally
		permsContext, exists := c.Get(ContextKeyPermissions)
		if exists {
			userPerms := permsContext.([]string)
			for _, p := range userPerms {
				if p == requiredPermission {
					c.Next()
					return
				}
			}
		}

		// 3. Current user lacks permission. Check for override PIN header.
		overridePin := c.GetHeader("X-Manager-Override-PIN")
		if overridePin == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions. Manager override required."})
			return
		}
		overridePin = strings.TrimSpace(overridePin)

		tenantID, _ := c.Get(ContextKeyTenantID)
		branchIDContext, exists := c.Get(ContextKeyBranchID)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Override requires an active branch session."})
			return
		}
		branchID := branchIDContext.(uuid.UUID)

		// 4. Find all profiles at this branch with a PIN set
		var profiles []models.UserProfile
		if err := db.Where("tenant_id = ? AND branch_id = ? AND pos_pin IS NOT NULL", tenantID, branchID).Preload("Role").Find(&profiles).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid override PIN"})
			return
		}

		var authorizingProfile *models.UserProfile
		for i := range profiles {
			p := &profiles[i]
			if p.POSPin == nil {
				continue
			}
			// Verify PIN hash
			err := bcrypt.CompareHashAndPassword([]byte(*p.POSPin), []byte(overridePin))
			if err == nil {
				authorizingProfile = p
				break
			}
		}

		if authorizingProfile == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid override PIN"})
			return
		}

		// 5. Verify the authorizing manager actually has the required permission
		var managerPerms []string
		db.Table("permissions").
			Select("permissions.code").
			Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
			Where("role_permissions.role_id = ?", authorizingProfile.RoleID).
			Pluck("permissions.code", &managerPerms)

		hasPerm := false
		for _, p := range managerPerms {
			if p == requiredPermission {
				hasPerm = true
				break
			}
		}

		if authorizingProfile.Role != nil && authorizingProfile.Role.Name == "Admin" {
			hasPerm = true // Admins bypass
		}

		if !hasPerm {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Authorizing manager lacks the required permission."})
			return
		}

		// 6. Success! Inject the authorizing manager's ID into context for audit logging
		c.Set("override_manager_id", authorizingProfile.UserID)
		c.Next()
	}
}
