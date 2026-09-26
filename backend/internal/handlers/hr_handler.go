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
		return userID
	}
	return profile.ID
}

// --- Attendance Handlers ---

func (h *HRHandler) GetMyAttendance(c *gin.Context) {
	profileID := h.getProfileID(c)
	attendance, todayHours, err := h.service(c).GetMyAttendance(profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attendance status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_clocked_in": attendance != nil,
		"attendance":    attendance,
		"today_hours":   todayHours,
	})
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

func (h *HRHandler) CorrectAttendance(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		ClockOut string `json:"clock_out" binding:"required"`
		Notes    string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clock_out is required (ISO 8601 format)"})
		return
	}

	clockOut, err := time.Parse(time.RFC3339, body.ClockOut)
	if err != nil {
		clockOut, err = time.Parse("2006-01-02T15:04", body.ClockOut)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clock_out format. Use ISO 8601 or YYYY-MM-DDTHH:mm."})
			return
		}
	}

	attendance, err := h.service(c).CorrectAttendance(id, clockOut, body.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

func (h *HRHandler) DeleteAttendance(c *gin.Context) {
	id := c.Param("id")

	if err := h.service(c).DeleteAttendance(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete attendance record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attendance record deleted"})
}

// --- Leave Handlers ---

func (h *HRHandler) ListLeaveRequests(c *gin.Context) {
	leaves, err := h.service(c).ListLeaveRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leave requests"})
		return
	}
	c.JSON(http.StatusOK, leaves)
}

type LeaveCreateRequest struct {
	StaffID   string `json:"staff_id"`
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

	staffID := profileID
	if req.StaffID != "" {
		if parsed, err := uuid.Parse(req.StaffID); err == nil {
			staffID = parsed
		}
	}

	input := services.LeaveCreateInput{
		StaffID:   staffID,
		LeaveType: req.LeaveType,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Reason:    req.Reason,
	}

	leave, err := h.service(c).CreateLeaveRequest(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, leave)
}

func (h *HRHandler) ApproveLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	profileID := h.getProfileID(c)

	var body struct {
		ManagerNotes string `json:"manager_notes"`
	}
	_ = c.ShouldBindJSON(&body)

	if err := h.service(c).ApproveLeaveRequest(id, profileID, body.ManagerNotes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve leave request"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

func (h *HRHandler) RejectLeaveRequest(c *gin.Context) {
	id := c.Param("id")
	profileID := h.getProfileID(c)

	var body struct {
		ManagerNotes string `json:"manager_notes"`
	}
	_ = c.ShouldBindJSON(&body)

	if err := h.service(c).RejectLeaveRequest(id, profileID, body.ManagerNotes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject leave request"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}

func (h *HRHandler) GetLeaveSummary(c *gin.Context) {
	staffID := c.Query("staff_id")
	summary := h.service(c).GetLeaveSummary(staffID)
	c.JSON(http.StatusOK, summary)
}

// --- Payroll Handlers ---

func (h *HRHandler) ListPayrollPeriods(c *gin.Context) {
	periods, err := h.service(c).ListPayrollPeriods()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payroll periods"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"periods": periods})
}

type CreatePayrollPeriodRequest struct {
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

func (h *HRHandler) CreatePayrollPeriod(c *gin.Context) {
	var req CreatePayrollPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	period, err := h.service(c).CreatePayrollPeriod(req.Name, req.StartDate, req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, period)
}

func (h *HRHandler) GetPayrollPeriod(c *gin.Context) {
	id := c.Param("id")
	period, err := h.service(c).GetPayrollPeriod(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll period not found"})
		return
	}
	c.JSON(http.StatusOK, period)
}

func (h *HRHandler) ProcessPayroll(c *gin.Context) {
	id := c.Param("id")
	period, err := h.service(c).ProcessPayroll(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payroll: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed", "period": period})
}

func (h *HRHandler) GetPayslip(c *gin.Context) {
	id := c.Param("id")
	record, err := h.service(c).GetPayslip(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payslip not found"})
		return
	}
	c.JSON(http.StatusOK, record)
}

type PayPayslipRequest struct {
	PaymentMethod    string `json:"payment_method"`
	PaymentReference string `json:"payment_reference"`
}

func (h *HRHandler) MarkPayslipPaid(c *gin.Context) {
	id := c.Param("id")
	var req PayPayslipRequest
	_ = c.ShouldBindJSON(&req)

	record, err := h.service(c).MarkPayslipPaid(id, req.PaymentMethod, req.PaymentReference)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payslip status: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

// --- Commission Rules Handlers ---

func (h *HRHandler) ListCommissionRules(c *gin.Context) {
	rules, err := h.service(c).ListCommissionRules()
	if err != nil {
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

	rule, err := h.service(c).CreateCommissionRule(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create commission rule"})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *HRHandler) DeleteCommissionRule(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).DeleteCommissionRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete commission rule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Commission rule deleted"})
}

// --- Staff Achievements Handlers ---

func (h *HRHandler) ListStaffAchievements(c *gin.Context) {
	achievements, err := h.service(c).ListStaffAchievements()
	if err != nil {
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

	achievement, err := h.service(c).CreateStaffAchievement(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff achievement"})
		return
	}
	c.JSON(http.StatusCreated, achievement)
}

func (h *HRHandler) DeleteStaffAchievement(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).DeleteStaffAchievement(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete staff achievement"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Staff achievement deleted"})
}

// --- Shift Swap Handlers ---

func (h *HRHandler) ListShiftSwapRequests(c *gin.Context) {
	requests, err := h.service(c).ListShiftSwaps()
	if err != nil {
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

	swap, err := h.service(c).CreateShiftSwap(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shift swap request"})
		return
	}
	c.JSON(http.StatusCreated, swap)
}

func (h *HRHandler) ApproveShiftSwap(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).ApproveShiftSwap(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve shift swap"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

func (h *HRHandler) RejectShiftSwap(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).RejectShiftSwap(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject shift swap"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}

