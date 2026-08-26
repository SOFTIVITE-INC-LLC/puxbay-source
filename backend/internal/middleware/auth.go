package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/services"
)

const (
	// ContextKeyUserID is the key for the authenticated user's ID.
	ContextKeyUserID = "user_id"
	// ContextKeyTenantID is the key for the current tenant ID.
	ContextKeyTenantID = "tenant_id"
	// ContextKeyBranchID is the key for the current branch ID.
	ContextKeyBranchID = "branch_id"
	// ContextKeyRole is the key for the user's role.
	ContextKeyRole        = "role"
	ContextKeyPermissions = "permissions"
	// ContextKeyClaims is the key for the full JWT claims.
	ContextKeyClaims = "claims"
)

// AuthMiddleware validates JWT tokens and sets user context.
// It first checks for the HttpOnly `pux_session` cookie (browser sessions),
// then falls back to the `Authorization: Bearer` header (API clients / mobile).
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Prefer the HttpOnly session cookie (XSS-safe)
		if cookie, err := c.Cookie("pux_session"); err == nil && cookie != "" {
			tokenString = cookie
		}

		// 2. Fallback: Authorization: Bearer header (mobile apps / API clients)
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString = parts[1]
				}
			}
		}

		// 3. Fallback to query parameter (mainly for WebSockets)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			return
		}

		// Validate token signature and expiry
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Ensure it's an access token, not a refresh token
		if claims.TokenType != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Access token required",
			})
			return
		}

		// Check the denylist — tokens are added here on logout
		if authService.IsTokenDenied(c.Request.Context(), claims.ID) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token has been revoked. Please log in again.",
			})
			return
		}

		// Set user context for downstream handlers
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyTenantID, claims.TenantID)
		c.Set(ContextKeyRole, claims.Role)
		// Fetch permissions using RoleID
		var permissions []string
		if claims.RoleID != nil {
			permissions, _ = authService.GetRolePermissions(claims.RoleID)
		} else {
			permissions = []string{}
		}

		c.Set(ContextKeyPermissions, permissions)
		c.Set(ContextKeyClaims, claims)

		if claims.BranchID != nil {
			c.Set(ContextKeyBranchID, *claims.BranchID)
		}

		c.Next()
	}
}

// RequirePermission restricts access to users with specific permissions.
func RequirePermission(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Superadmin bypass
		if role, exists := c.Get(ContextKeyRole); exists && (role.(string) == "superadmin" || role.(string) == "admin") {
			c.Next()
			return
		}
		if isSuper, exists := c.Get(ContextKeyIsSuperuser); exists && isSuper.(bool) {
			c.Next()
			return
		}

		permsContext, exists := c.Get(ContextKeyPermissions)
		if !exists {
			slog.Warn("Access Denied: No permissions found in token",
				slog.String("path", c.Request.URL.Path),
				slog.String("method", c.Request.Method),
				slog.Any("user_id", c.Value(ContextKeyUserID)),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No permissions found in token"})
			return
		}

		userPerms := permsContext.([]string)
		
		// Check if user has at least one of the required permissions
		for _, required := range requiredPermissions {
			for _, p := range userPerms {
				if p == required {
					c.Next()
					return
				}
			}
		}

		slog.Warn("Access Denied: Insufficient permissions",
			slog.String("path", c.Request.URL.Path),
			slog.String("method", c.Request.Method),
			slog.Any("user_id", c.Value(ContextKeyUserID)),
			slog.Any("required", requiredPermissions),
		)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
	}
}

// RoleMiddleware restricts access to specific roles (Legacy).
// Deprecated: Use RequirePermission instead.
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "No role found in token",
			})
			return
		}

		userRole := role.(string)
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
	}
}

// SuperAdminMiddleware restricts access to superusers only.
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(ContextKeyClaims)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Authentication required",
			})
			return
		}

		tokenClaims := claims.(*services.Claims)
		// Superadmin check — role must be "superadmin" or user must be marked as superuser
		if tokenClaims.Role != "superadmin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Superadmin access required",
			})
			return
		}

		c.Next()
	}
}

// FieldRBAC provides granular field-level masking based on user permissions.
// E.g. masking PII, salaries, cost_price, etc.
func FieldRBAC(role string, permissions []string, data map[string]interface{}) map[string]interface{} {
	if role == "admin" || role == "superadmin" {
		return data // Admins see everything
	}

	hasPerm := func(p string) bool {
		for _, perm := range permissions {
			if perm == p {
				return true
			}
		}
		return false
	}

	// 1. HR Data
	if !hasPerm("hr:manage") {
		delete(data, "base_salary")
		delete(data, "hourly_rate")
		delete(data, "bank_details")
	}

	// 2. Financial Data
	if !hasPerm("financial:manage") && !hasPerm("inventory:manage") {
		delete(data, "cost_price")
	}
	if !hasPerm("financial:manage") && !hasPerm("products:manage") {
		delete(data, "wholesale_price")
	}

	// 3. PII Masking
	if !hasPerm("customers:manage") && !hasPerm("staff:manage") {
		if email, ok := data["email"].(string); ok && email != "" {
			data["email"] = maskEmail(email)
		}
		if phone, ok := data["phone"].(string); ok && phone != "" {
			data["phone"] = maskPhone(phone)
		}
	}

	return data
}

func maskEmail(email string) string {
	if len(email) > 4 {
		return email[:2] + "***" + email[len(email)-4:]
	}
	return "***"
}

func maskPhone(phone string) string {
	if len(phone) > 4 {
		return phone[:2] + "****" + phone[len(phone)-2:]
	}
	return "****"
}
