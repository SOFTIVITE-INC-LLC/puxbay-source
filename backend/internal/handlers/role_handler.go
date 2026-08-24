package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type RoleHandler struct {
	db *gorm.DB
}

func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

func (h *RoleHandler) service(c *gin.Context) *services.RoleService {
	return services.NewRoleService(getDB(c, h.db))
}

// ListPermissions returns all available system permissions
// GET /api/v1/permissions
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	perms, err := h.service(c).ListPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list permissions"})
		return
	}
	c.JSON(http.StatusOK, perms)
}

// ListRoles returns all roles for the current tenant
// GET /api/v1/roles
func (h *RoleHandler) ListRoles(c *gin.Context) {
	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, _ := rawTenantID.(uuid.UUID)

	roles, err := h.service(c).ListRoles(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

// CreateRole creates a new custom role
// POST /api/v1/roles
func (h *RoleHandler) CreateRole(c *gin.Context) {
	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, _ := rawTenantID.(uuid.UUID)

	var req struct {
		Name          string      `json:"name" binding:"required"`
		Description   string      `json:"description"`
		PermissionIDs []uuid.UUID `json:"permission_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	role, err := h.service(c).CreateRole(tenantID, req.Name, req.Description, req.PermissionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, role)
}

// UpdateRole updates an existing custom role
// PUT /api/v1/roles/:id
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, _ := rawTenantID.(uuid.UUID)

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}

	var req struct {
		Name          string      `json:"name" binding:"required"`
		Description   string      `json:"description"`
		PermissionIDs []uuid.UUID `json:"permission_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	role, err := h.service(c).UpdateRole(tenantID, roleID, req.Name, req.Description, req.PermissionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

// DeleteRole deletes a custom role
// DELETE /api/v1/roles/:id
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, _ := rawTenantID.(uuid.UUID)

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}

	if err := h.service(c).DeleteRole(tenantID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}
