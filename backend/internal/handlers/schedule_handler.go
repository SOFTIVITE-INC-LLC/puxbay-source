package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type ScheduleHandler struct {
	db *gorm.DB
}

func NewScheduleHandler(db *gorm.DB) *ScheduleHandler {
	return &ScheduleHandler{db: db}
}

type CreateShiftInput struct {
	StaffID   uuid.UUID `json:"staff_id" binding:"required"`
	StartTime string    `json:"start_time" binding:"required"`
	EndTime   string    `json:"end_time" binding:"required"`
	Role      string    `json:"role"`
	Notes     string    `json:"notes"`
}

func (h *ScheduleHandler) ListShifts(c *gin.Context) {
	branchID, _ := c.Get(middleware.ContextKeyBranchID)
	db := getDB(c, h.db)

	var shifts []models.StaffShift
	if err := db.Where("branch_id = ?", branchID).Order("start_time asc").Find(&shifts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shifts"})
		return
	}

	c.JSON(http.StatusOK, shifts)
}

func (h *ScheduleHandler) CreateShift(c *gin.Context) {
	branchID, _ := c.Get(middleware.ContextKeyBranchID)
	db := getDB(c, h.db)

	var input CreateShiftInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start time format"})
		return
	}
	endTime, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end time format"})
		return
	}

	bID := branchID.(uuid.UUID)
	shift := models.StaffShift{
		BranchScoped: models.BranchScoped{
			TenantScoped: models.TenantScoped{
				Base: models.Base{},
			},
			BranchID: &bID,
		},
		StaffID:   input.StaffID,
		StartTime: startTime,
		EndTime:   endTime,
		Role:      input.Role,
		Notes:     input.Notes,
	}

	if err := db.Create(&shift).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shift", "details": err.Error()})
		return
	}

	// Wait, schema isolation handles TenantID, but BranchScoped contains TenantScoped which embeds Base.
	// We don't need to explicitly assign TenantID anymore.

	c.JSON(http.StatusCreated, shift)
}
