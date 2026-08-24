package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type HRService struct {
	db *gorm.DB
}

func NewHRService(db *gorm.DB) *HRService {
	return &HRService{db: db}
}

func (s *HRService) ListAttendance(branchID, dateFrom, dateTo, staffID string) ([]models.Attendance, error) {
	var attendances []models.Attendance
	query := s.db.Model(&models.Attendance{})
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	if dateFrom != "" {
		query = query.Where("clock_in >= ?", dateFrom+" 00:00:00")
	}
	if dateTo != "" {
		query = query.Where("clock_in <= ?", dateTo+" 23:59:59")
	}
	if staffID != "" {
		query = query.Where("staff_id = ?", staffID)
	}
	if err := query.Order("clock_in desc").Find(&attendances).Error; err != nil {
		return nil, err
	}
	return attendances, nil
}

func (s *HRService) ClockIn(staffID uuid.UUID) (*models.Attendance, error) {
	var existing models.Attendance
	err := s.db.Where("staff_id = ? AND clock_out IS NULL", staffID).First(&existing).Error
	if err == nil {
		return nil, errors.New("already clocked in")
	}

	attendance := models.Attendance{
		StaffID: staffID,
		ClockIn: time.Now(),
	}

	if err := s.db.Create(&attendance).Error; err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (s *HRService) ClockOut(staffID uuid.UUID) (*models.Attendance, error) {
	var attendance models.Attendance
	err := s.db.Where("staff_id = ? AND clock_out IS NULL", staffID).Order("clock_in desc").First(&attendance).Error
	if err != nil {
		return nil, errors.New("no active clock-in found")
	}

	now := time.Now()
	attendance.ClockOut = &now
	if err := s.db.Save(&attendance).Error; err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (s *HRService) ListLeaveRequests() ([]models.LeaveRequest, error) {
	var leaves []models.LeaveRequest
	if err := s.db.Find(&leaves).Error; err != nil {
		return nil, err
	}
	return leaves, nil
}

type LeaveCreateInput struct {
	StaffID   uuid.UUID
	LeaveType string
	StartDate string
	EndDate   string
	Reason    string
}

func (s *HRService) CreateLeaveRequest(input LeaveCreateInput) (*models.LeaveRequest, error) {
	start, err := time.Parse(time.RFC3339, input.StartDate)
	if err != nil {
		return nil, errors.New("invalid start date format")
	}

	end, err := time.Parse(time.RFC3339, input.EndDate)
	if err != nil {
		return nil, errors.New("invalid end date format")
	}

	leave := models.LeaveRequest{
		StaffID:   input.StaffID,
		LeaveType: input.LeaveType,
		StartDate: start,
		EndDate:   end,
		Reason:    &input.Reason,
		Status:    "pending",
	}

	if err := s.db.Create(&leave).Error; err != nil {
		return nil, err
	}

	return &leave, nil
}

func (s *HRService) ApproveLeaveRequest(id string, reviewerID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&models.LeaveRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         "approved",
		"reviewed_by_id": reviewerID,
		"reviewed_at":    now,
	}).Error
}

func (s *HRService) RejectLeaveRequest(id string, reviewerID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&models.LeaveRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         "rejected",
		"reviewed_by_id": reviewerID,
		"reviewed_at":    now,
	}).Error
}
