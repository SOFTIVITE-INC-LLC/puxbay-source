package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type RoleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// ListPermissions returns all available system permissions grouped or flat.
func (s *RoleService) ListPermissions() ([]models.Permission, error) {
	var perms []models.Permission
	err := s.db.Order("category asc, code asc").Find(&perms).Error
	return perms, err
}

// ListRoles returns all roles accessible to a tenant (including system roles).
func (s *RoleService) ListRoles(tenantID uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	err := s.db.Preload("Permissions").
		Where("tenant_id = ? OR is_system = ?", tenantID, true).
		Order("name asc").
		Find(&roles).Error
	return roles, err
}

// CreateRole creates a custom role for a tenant.
func (s *RoleService) CreateRole(tenantID uuid.UUID, name, description string, permissionIDs []uuid.UUID) (*models.Role, error) {
	var perms []models.Permission
	if len(permissionIDs) > 0 {
		if err := s.db.Where("id IN ?", permissionIDs).Find(&perms).Error; err != nil {
			return nil, err
		}
	}

	role := &models.Role{
		TenantID:    &tenantID,
		Name:        name,
		Description: description,
		IsSystem:    false,
		Permissions: perms,
	}

	if err := s.db.Create(role).Error; err != nil {
		return nil, err
	}

	return role, nil
}

// UpdateRole updates a custom role.
func (s *RoleService) UpdateRole(tenantID, roleID uuid.UUID, name, description string, permissionIDs []uuid.UUID) (*models.Role, error) {
	var role models.Role
	if err := s.db.Where("id = ? AND tenant_id = ?", roleID, tenantID).First(&role).Error; err != nil {
		return nil, errors.New("role not found or unauthorized")
	}

	if role.IsSystem {
		return nil, errors.New("cannot modify system roles")
	}

	role.Name = name
	role.Description = description

	// Replace permissions
	var perms []models.Permission
	if len(permissionIDs) > 0 {
		if err := s.db.Where("id IN ?", permissionIDs).Find(&perms).Error; err != nil {
			return nil, err
		}
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&role).Error; err != nil {
			return err
		}
		if err := tx.Model(&role).Association("Permissions").Replace(perms); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteRole deletes a custom role.
func (s *RoleService) DeleteRole(tenantID, roleID uuid.UUID) error {
	var role models.Role
	if err := s.db.Where("id = ? AND tenant_id = ?", roleID, tenantID).First(&role).Error; err != nil {
		return errors.New("role not found or unauthorized")
	}

	if role.IsSystem {
		return errors.New("cannot delete system roles")
	}

	// Make sure no users are using this role
	var count int64
	if err := s.db.Model(&models.UserProfile{}).Where("role_id = ?", roleID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}

	if count > 0 {
		return errors.New("cannot delete role because it is currently assigned to users")
	}

	return s.db.Delete(&role).Error
}
