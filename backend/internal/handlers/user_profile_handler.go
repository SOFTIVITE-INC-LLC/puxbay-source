package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

// UserProfileHandler handles staff/user profile management for a tenant.
type UserProfileHandler struct {
	db *gorm.DB
}

// NewUserProfileHandler creates a new user profile handler.
func NewUserProfileHandler(db *gorm.DB) *UserProfileHandler {
	return &UserProfileHandler{db: db}
}

func (h *UserProfileHandler) service(c *gin.Context) *services.UserProfileService {
	return services.NewUserProfileService(getDB(c, h.db))
}

// List returns all user profiles (staff members) for the tenant.
// GET /api/v1/profiles
func (h *UserProfileHandler) List(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	profiles, err := h.service(c).ListProfiles(tenantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user profiles"})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

// Get returns a specific user profile.
// GET /api/v1/profiles/:id
func (h *UserProfileHandler) Get(c *gin.Context) {
	id := c.Param("id")
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	profile, err := h.service(c).GetProfile(id, tenantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateRequest defines the payload for updating a user profile.
type ProfileUpdateRequest struct {
	Role                  string  `json:"role"`
	CanPerformCreditSales *bool   `json:"can_perform_credit_sales"`
	BaseSalary            float64 `json:"base_salary"`
	HourlyRate            float64 `json:"hourly_rate"`
}

// Update modifies an existing user profile's settings.
// PUT /api/v1/profiles/:id
func (h *UserProfileHandler) Update(c *gin.Context) {
	id := c.Param("id")
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	var req ProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	input := services.ProfileUpdateInput{
		Role:                  req.Role,
		CanPerformCreditSales: req.CanPerformCreditSales,
		BaseSalary:            req.BaseSalary,
		HourlyRate:            req.HourlyRate,
	}

	profile, err := h.service(c).UpdateProfile(id, tenantID.(uuid.UUID), input)
	if err != nil {
		if err.Error() == "user profile not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// SetPOSPin allows a user to set their POS override PIN.
// PUT /api/v1/profile/pos-pin
func (h *UserProfileHandler) SetPOSPin(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	var req struct {
		PIN string `json:"pin" binding:"required,len=4,numeric"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PIN format. Must be a 4-digit number."})
		return
	}

	if err := h.service(c).SetPOSPin(userID.(uuid.UUID), tenantID.(uuid.UUID), req.PIN); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set POS PIN"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "POS PIN updated successfully"})
}
