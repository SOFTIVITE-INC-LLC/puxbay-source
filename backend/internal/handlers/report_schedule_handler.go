package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type ReportScheduleHandler struct {
	db        *gorm.DB
	reportSvc *services.ReportService
	emailSvc  *services.EmailService
}

func NewReportScheduleHandler(db *gorm.DB, reportSvc *services.ReportService, emailSvc *services.EmailService) *ReportScheduleHandler {
	return &ReportScheduleHandler{db: db, reportSvc: reportSvc, emailSvc: emailSvc}
}

func (h *ReportScheduleHandler) getReportService(c *gin.Context) *services.ReportService {
	return services.NewReportService(getDB(c, h.db), h.emailSvc.GetSMTPConfig())
}

func getTenantIDHelper(c *gin.Context) (uuid.UUID, error) {
	if tID, exists := c.Get(middleware.ContextKeyTenantID); exists {
		if id, ok := tID.(uuid.UUID); ok {
			return id, nil
		}
		if s, ok := tID.(string); ok {
			return uuid.Parse(s)
		}
	}
	return uuid.Nil, errors.New("tenant not identified")
}

// GetSchedules handles GET /api/v1/reports/schedules
func (h *ReportScheduleHandler) GetSchedules(c *gin.Context) {
	tenantID, err := getTenantIDHelper(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not identified"})
		return
	}

	db := getDB(c, h.db)
	var schedules []models.ReportSchedule
	db.Where("tenant_id = ?", tenantID).Find(&schedules)

	// If no schedules exist, seed default template schedules
	if len(schedules) == 0 {
		defaultSchedules := []models.ReportSchedule{
			{
				TenantID:   tenantID,
				ReportType: "daily_z",
				Recipients: "owner@puxbay.com",
				IsEnabled:  true,
				SendTime:   "23:59",
				IncludePDF: true,
			},
			{
				TenantID:   tenantID,
				ReportType: "weekly_pl",
				Recipients: "owner@puxbay.com",
				IsEnabled:  true,
				SendTime:   "23:59",
				IncludePDF: true,
			},
			{
				TenantID:   tenantID,
				ReportType: "monthly_pl",
				Recipients: "owner@puxbay.com",
				IsEnabled:  true,
				SendTime:   "23:59",
				IncludePDF: true,
			},
		}
		for _, s := range defaultSchedules {
			db.Create(&s)
		}
		db.Where("tenant_id = ?", tenantID).Find(&schedules)
	}

	c.JSON(http.StatusOK, gin.H{"schedules": schedules})
}

type UpdateScheduleRequest struct {
	ReportType string `json:"report_type" binding:"required"`
	Recipients string `json:"recipients" binding:"required"`
	IsEnabled  bool   `json:"is_enabled"`
	SendTime   string `json:"send_time"`
	IncludePDF bool   `json:"include_pdf"`
}

// SaveSchedule handles POST /api/v1/reports/schedules
func (h *ReportScheduleHandler) SaveSchedule(c *gin.Context) {
	tenantID, err := getTenantIDHelper(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not identified"})
		return
	}

	var req UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := getDB(c, h.db)
	var schedule models.ReportSchedule
	err = db.Where("tenant_id = ? AND report_type = ?", tenantID, req.ReportType).First(&schedule).Error
	if err != nil {
		schedule = models.ReportSchedule{
			TenantID:   tenantID,
			ReportType: req.ReportType,
		}
	}

	schedule.Recipients = req.Recipients
	schedule.IsEnabled = req.IsEnabled
	if req.SendTime != "" {
		schedule.SendTime = req.SendTime
	}
	schedule.IncludePDF = req.IncludePDF

	if schedule.ID == uuid.Nil {
		db.Create(&schedule)
	} else {
		db.Save(&schedule)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report schedule saved", "schedule": schedule})
}

type SendTestReportRequest struct {
	ReportType string `json:"report_type" binding:"required"`
	Recipient  string `json:"recipient" binding:"required"`
}

// SendTestReport handles POST /api/v1/reports/send-test
func (h *ReportScheduleHandler) SendTestReport(c *gin.Context) {
	tenantID, err := getTenantIDHelper(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not identified"})
		return
	}

	var req SendTestReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	reportSvc := h.getReportService(c)

	switch req.ReportType {
	case "daily_z":
		data, err := reportSvc.GenerateDailyZReport(tenantID, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		html, err := reportSvc.RenderDailyZReportHTML(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		subject := fmt.Sprintf("📊 [TEST] Daily Z-Report & Sales Audit - %s", data.Date)
		_ = h.emailSvc.SendRawHTML([]string{req.Recipient}, subject, html)

	case "weekly_pl":
		start := now.AddDate(0, 0, -7)
		data, err := reportSvc.GeneratePLReport(tenantID, "Weekly", start, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		html, err := reportSvc.RenderPLReportHTML(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		subject := fmt.Sprintf("📈 [TEST] Weekly Profit & Loss Summary (%s - %s)", data.StartDate, data.EndDate)
		_ = h.emailSvc.SendRawHTML([]string{req.Recipient}, subject, html)

	case "monthly_pl":
		start := now.AddDate(0, -1, 0)
		data, err := reportSvc.GeneratePLReport(tenantID, "Monthly", start, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		html, err := reportSvc.RenderPLReportHTML(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		subject := fmt.Sprintf("📑 [TEST] Monthly Financial P&L Statement (%s - %s)", data.StartDate, data.EndDate)
		_ = h.emailSvc.SendRawHTML([]string{req.Recipient}, subject, html)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Test %s report dispatched to %s", req.ReportType, req.Recipient)})
}

// GetDailyZReportData handles GET /api/v1/reports/daily-z
func (h *ReportScheduleHandler) GetDailyZReportData(c *gin.Context) {
	tenantID, err := getTenantIDHelper(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not identified"})
		return
	}

	dateStr := c.Query("date")
	targetDate := time.Now()
	if dateStr != "" {
		if parsed, parseErr := time.Parse("2006-01-02", dateStr); parseErr == nil {
			targetDate = parsed
		}
	}

	data, err := h.getReportService(c).GenerateDailyZReport(tenantID, targetDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
