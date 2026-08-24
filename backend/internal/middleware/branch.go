package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	BranchContextKey = "branchID"
)

// BranchMiddleware extracts the X-Branch-ID header and injects it into the request context.
func BranchMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		branchIDStr := c.GetHeader("X-Branch-ID")
		if branchIDStr != "" {
			if branchID, err := uuid.Parse(branchIDStr); err == nil {
				// Inject the valid branch UUID into context
				c.Set(BranchContextKey, branchID)
			}
		}
		c.Next()
	}
}

// GetBranchID returns the branch UUID from the Gin context (set by BranchMiddleware).
// Returns (nil, false) when no branch context is present.
func GetBranchID(c *gin.Context) (*uuid.UUID, bool) {
	val, exists := c.Get(BranchContextKey)
	if !exists {
		return nil, false
	}
	bID, ok := val.(uuid.UUID)
	if !ok {
		return nil, false
	}
	return &bID, true
}

// BranchRequiredMiddleware ensures that a branch ID is present in the context.
func BranchRequiredMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := GetBranchID(c); !ok {
			c.AbortWithStatusJSON(400, gin.H{"error": "Branch context required (missing X-Branch-ID header)"})
			return
		}
		c.Next()
	}
}

// ResolveBranchID safely resolves the requested branch against the user's bound branch context.
// If the user is bound to a specific branch, that branch is strictly enforced.
// If the user is a tenant admin (no bound branch), it allows querying the requested branch.
func ResolveBranchID(c *gin.Context, requested string) string {
	// ContextKeyBranchID is set in auth.go for non-global users
	if boundBranch, exists := c.Get("branch_id"); exists {
		if bID, ok := boundBranch.(uuid.UUID); ok {
			return bID.String()
		}
		if bIDStr, ok := boundBranch.(string); ok {
			return bIDStr
		}
	}

	if requested != "" {
		return requested
	}

	// Fallback to X-Branch-ID header parsed by BranchMiddleware
	if headerBranch, exists := c.Get(BranchContextKey); exists {
		if bID, ok := headerBranch.(uuid.UUID); ok {
			return bID.String()
		}
	}

	return requested
}
