package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type HRHandler struct {
	db *gorm.DB
}

func NewHRHandler(db *gorm.DB) *HRHandler {
	return &HRHandler{db: db}
}

func (h *HRHandler) service(c *gin.Context) *services.HRService {
	return services.NewHRService(getDB(c, h.db))
}

func (h *HRHandler) getProfileID(c *gin.Context) uuid.UUID {
	// Use user_id from JWT claims to look up the UserProfile for this tenant
	userIDRaw, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		return uuid.Nil
	}
	var userID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		userID = v
	case string:
		var err error
		if userID, err = uuid.Parse(v); err != nil {
			return uuid.Nil
		}
	default:
		return uuid.Nil
	}
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	var profile models.UserProfile
	if err := h.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&profile).Error; err != nil {
		// Fallback: use the user ID directly as the staff identifier
		return userID
	}
	return profile.ID
}

func (h *HRHandler) ListAttendance(c *gin.Context) {
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	staffID := c.Query("staff_id")

	attendances, err := h.service(c).ListAttendance(branchID, dateFrom, dateTo, staffID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attendance"})
		return
	}
	c.JSON(http.StatusOK, attendances)
}

// ClockIn matches Django's clock_in action
func (h *HRHandler) ClockIn(c *gin.Context) {
	profileID := h.getProfileID(c)

	attendance, err := h.service(c).ClockIn(profileID)
	if err != nil {
		if err.Error() == "already clocked in" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clock in"})
		return
	}

	c.JSON(http.StatusCreated, attendance)
}

// ClockOut matches Django's clock_out action
func (h *HRHandler) ClockOut(c *gin.Context) {
	profileID := h.getProfileID(c)

	attendance, err := h.service(c).ClockOut(profileID)
	if err != nil {
		if err.Error() == "no active clock-in found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clock out"})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

func (h *HRHandler) ListLeaveRequests(c *gin.Context) {
	leaves, err := h.service(c).ListLeaveRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leave requests"})
		return
	}
	c.JSON(http.StatusOK, leaves)
}

type LeaveCreateRequest struct {
	LeaveType string `json:"leave_type" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Reason    string `json:"reason"`
}

func (h *HRHandler) CreateLeaveRequest(c *gin.Context) {
	profileID := h.getProfileID(c)

	var req LeaveCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.LeaveCreateInput{
		StaffID:   profileID,
		LeaveType: req.LeaveType,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Reason:    req.Reason,
	}

	leave, err := h.service(c).CreateLeaveRequest(input)
	if err != nil {
		if err.Error() == "invalid start date format" || err.Error() == "invalid end date format" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create leave request"})
		return
	}

	c.JSON(http.StatusCreated, leave)
}
func (h *HRHandler) ApproveLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	profileID := h.getProfileID(c)

	if err := h.service(c).ApproveLeaveRequest(id, profileID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve leave request"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

func (h *HRHandler) RejectLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	profileID := h.getProfileID(c)

	if err := h.service(c).RejectLeaveRequest(id, profileID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject leave request"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}
func (h *HRHandler) ListPayrollPeriods(c *gin.Context) {
	var periods []models.PayrollPeriod
	if err := getDB(c, h.db).Order("start_date desc").Find(&periods).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch payroll periods"})
		return
	}
	c.JSON(200, gin.H{"periods": periods})
}

func (h *HRHandler) GetPayrollPeriod(c *gin.Context) {
	id := c.Param("id")
	var period models.PayrollPeriod
	if err := getDB(c, h.db).Where("id = ?", id).First(&period).Error; err != nil {
		c.JSON(404, gin.H{"error": "Payroll period not found"})
		return
	}
	c.JSON(200, period)
}

func (h *HRHandler) ProcessPayroll(c *gin.Context) {
	id := c.Param("id")
	var period models.PayrollPeriod

	err := getDB(c, h.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).First(&period).Error; err != nil {
			return err
		}
		period.IsProcessed = true
		now := time.Now()
		period.ProcessedAt = &now
		return tx.Save(&period).Error
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to process payroll"})
		return
	}

	c.JSON(200, gin.H{"status": "processed", "period": period})
}

func (h *HRHandler) GetPayslip(c *gin.Context) {
	id := c.Param("id") // PayrollRecord ID
	var record models.PayrollRecord
	if err := getDB(c, h.db).Where("id = ?", id).First(&record).Error; err != nil {
		c.JSON(404, gin.H{"error": "Payslip not found"})
		return
	}
	c.JSON(200, record)
}

func (h *HRHandler) ListCommissionRules(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	var rules []models.CommissionRule
	if err := getDB(c, h.db).Where("tenant_id = ?", tenantID).Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commission rules"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *HRHandler) CreateCommissionRule(c *gin.Context) {
	var req models.CommissionRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := getDB(c, h.db).Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create commission rule"})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *HRHandler) ListStaffAchievements(c *gin.Context) {
	var achievements []models.StaffAchievement
	if err := getDB(c, h.db).Find(&achievements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch staff achievements"})
		return
	}
	c.JSON(http.StatusOK, achievements)
}

func (h *HRHandler) CreateStaffAchievement(c *gin.Context) {
	var req models.StaffAchievement
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := getDB(c, h.db).Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff achievement"})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *HRHandler) ListShiftSwapRequests(c *gin.Context) {
	var requests []models.ShiftSwapRequest
	if err := getDB(c, h.db).Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shift swap requests"})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func (h *HRHandler) CreateShiftSwapRequest(c *gin.Context) {
	var req models.ShiftSwapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := getDB(c, h.db).Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shift swap request"})
		return
	}
	c.JSON(http.StatusCreated, req)
}

// CorrectAttendance allows a manager to manually set a clock_out time on an attendance record
func (h *HRHandler) CorrectAttendance(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		ClockOut string `json:"clock_out" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clock_out is required (ISO 8601 format)"})
		return
	}

	clockOut, err := time.Parse(time.RFC3339, body.ClockOut)
	if err != nil {
		// Try a simpler format
		clockOut, err = time.Parse("2006-01-02T15:04", body.ClockOut)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clock_out format. Use ISO 8601."})
			return
		}
	}

	var attendance models.Attendance
	if err := getDB(c, h.db).Where("id = ?", id).First(&attendance).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	attendance.ClockOut = &clockOut
	if err := getDB(c, h.db).Save(&attendance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update attendance"})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

// DeleteAttendance allows a manager to remove a ghost/erroneous attendance record
func (h *HRHandler) DeleteAttendance(c *gin.Context) {
	id := c.Param("id")

	if err := getDB(c, h.db).Where("id = ?", id).Delete(&models.Attendance{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete attendance record"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
