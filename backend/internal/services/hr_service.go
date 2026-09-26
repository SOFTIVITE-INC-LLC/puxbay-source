package services

import (
	"errors"
	"math"
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

// --- Attendance Management ---

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

func (s *HRService) GetMyAttendance(staffID uuid.UUID) (*models.Attendance, float64, error) {
	var active models.Attendance
	err := s.db.Where("staff_id = ? AND clock_out IS NULL", staffID).Order("clock_in desc").First(&active).Error
	var activePtr *models.Attendance
	if err == nil {
		activePtr = &active
	}

	todayStart := time.Now().Format("2006-01-02") + " 00:00:00"
	var todayRecords []models.Attendance
	_ = s.db.Where("staff_id = ? AND clock_in >= ?", staffID, todayStart).Find(&todayRecords).Error

	var todayHours float64
	for _, r := range todayRecords {
		if r.ClockOut != nil {
			todayHours += math.Max(0, r.ClockOut.Sub(r.ClockIn).Hours())
		} else {
			todayHours += math.Max(0, time.Since(r.ClockIn).Hours())
		}
	}
	todayHours = math.Round(todayHours*100) / 100

	return activePtr, todayHours, nil
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
		Status:  "present",
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
	attendance.Status = "completed"
	durationHours := math.Max(0, now.Sub(attendance.ClockIn).Hours())
	attendance.HoursWorked = math.Round(durationHours*100) / 100

	if err := s.db.Save(&attendance).Error; err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (s *HRService) CorrectAttendance(id string, clockOut time.Time, notes string) (*models.Attendance, error) {
	var attendance models.Attendance
	if err := s.db.Where("id = ?", id).First(&attendance).Error; err != nil {
		return nil, errors.New("attendance record not found")
	}

	attendance.ClockOut = &clockOut
	attendance.Status = "completed"
	durationHours := math.Max(0, clockOut.Sub(attendance.ClockIn).Hours())
	attendance.HoursWorked = math.Round(durationHours*100) / 100
	if notes != "" {
		attendance.Notes = &notes
	}

	if err := s.db.Save(&attendance).Error; err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (s *HRService) DeleteAttendance(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.Attendance{}).Error
}

// --- Leave Management ---

func (s *HRService) ListLeaveRequests() ([]models.LeaveRequest, error) {
	var leaves []models.LeaveRequest
	if err := s.db.Preload("ReviewedBy").Order("created_at desc").Find(&leaves).Error; err != nil {
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

func parseDateFlexible(val string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, val)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("2006-01-02T15:04", val)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("2006-01-02", val)
	if err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("invalid date format")
}

func (s *HRService) CreateLeaveRequest(input LeaveCreateInput) (*models.LeaveRequest, error) {
	start, err := parseDateFlexible(input.StartDate)
	if err != nil {
		return nil, errors.New("invalid start date format")
	}

	end, err := parseDateFlexible(input.EndDate)
	if err != nil {
		return nil, errors.New("invalid end date format")
	}

	days := int(math.Ceil(end.Sub(start).Hours()/24)) + 1
	if days < 1 {
		days = 1
	}

	leave := models.LeaveRequest{
		StaffID:   input.StaffID,
		LeaveType: input.LeaveType,
		StartDate: start,
		EndDate:   end,
		DaysCount: days,
		Reason:    &input.Reason,
		Status:    "pending",
	}

	if err := s.db.Create(&leave).Error; err != nil {
		return nil, err
	}

	return &leave, nil
}

func (s *HRService) ApproveLeaveRequest(id string, reviewerID uuid.UUID, managerNotes string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":         "approved",
		"reviewed_by_id": reviewerID,
		"reviewed_at":    now,
	}
	if managerNotes != "" {
		updates["manager_notes"] = managerNotes
	}
	return s.db.Model(&models.LeaveRequest{}).Where("id = ?", id).Updates(updates).Error
}

func (s *HRService) RejectLeaveRequest(id string, reviewerID uuid.UUID, managerNotes string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":         "rejected",
		"reviewed_by_id": reviewerID,
		"reviewed_at":    now,
	}
	if managerNotes != "" {
		updates["manager_notes"] = managerNotes
	}
	return s.db.Model(&models.LeaveRequest{}).Where("id = ?", id).Updates(updates).Error
}

type LeaveSummary struct {
	TotalQuota    int `json:"total_quota"`
	UsedDays      int `json:"used_days"`
	PendingDays   int `json:"pending_days"`
	RemainingDays int `json:"remaining_days"`
}

func (s *HRService) GetLeaveSummary(staffID string) LeaveSummary {
	summary := LeaveSummary{
		TotalQuota:    21, // Standard standard 21 days annual leave entitlement
		UsedDays:      0,
		PendingDays:   0,
		RemainingDays: 21,
	}

	var approvedLeaves []models.LeaveRequest
	q := s.db.Where("status = ?", "approved")
	if staffID != "" {
		q = q.Where("staff_id = ?", staffID)
	}
	if err := q.Find(&approvedLeaves).Error; err == nil {
		for _, l := range approvedLeaves {
			summary.UsedDays += l.DaysCount
		}
	}

	var pendingLeaves []models.LeaveRequest
	pq := s.db.Where("status = ?", "pending")
	if staffID != "" {
		pq = pq.Where("staff_id = ?", staffID)
	}
	if err := pq.Find(&pendingLeaves).Error; err == nil {
		for _, l := range pendingLeaves {
			summary.PendingDays += l.DaysCount
		}
	}

	summary.RemainingDays = int(math.Max(0, float64(summary.TotalQuota-summary.UsedDays)))
	return summary
}

// --- Payroll Management Engine ---

func (s *HRService) CreatePayrollPeriod(name string, startDateStr, endDateStr string) (*models.PayrollPeriod, error) {
	start, err := parseDateFlexible(startDateStr)
	if err != nil {
		return nil, errors.New("invalid start date format")
	}
	end, err := parseDateFlexible(endDateStr)
	if err != nil {
		return nil, errors.New("invalid end date format")
	}

	period := models.PayrollPeriod{
		Name:        name,
		StartDate:   start,
		EndDate:     end,
		IsProcessed: false,
	}

	if err := s.db.Create(&period).Error; err != nil {
		return nil, err
	}

	return &period, nil
}

func (s *HRService) ListPayrollPeriods() ([]models.PayrollPeriod, error) {
	var periods []models.PayrollPeriod
	if err := s.db.Order("start_date desc").Find(&periods).Error; err != nil {
		return nil, err
	}
	return periods, nil
}

func (s *HRService) GetPayrollPeriod(id string) (*models.PayrollPeriod, error) {
	var period models.PayrollPeriod
	if err := s.db.Where("id = ?", id).Preload("Records").First(&period).Error; err != nil {
		return nil, errors.New("payroll period not found")
	}
	return &period, nil
}

func (s *HRService) ProcessPayroll(periodID string) (*models.PayrollPeriod, error) {
	var period models.PayrollPeriod
	if err := s.db.Where("id = ?", periodID).First(&period).Error; err != nil {
		return nil, errors.New("payroll period not found")
	}

	// 1. Fetch all staff profiles
	var profiles []models.UserProfile
	if err := s.db.Preload("User").Preload("Role").Find(&profiles).Error; err != nil {
		return nil, err
	}

	// 2. Fetch active commission rules
	var commissionRules []models.CommissionRule
	_ = s.db.Where("is_active = ?", true).Find(&commissionRules).Error

	var totalGross, totalNet float64
	staffCount := 0

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing draft records for this period to recalculate fresh
		if err := tx.Where("period_id = ?", period.ID).Delete(&models.PayrollRecord{}).Error; err != nil {
			return err
		}

		for _, profile := range profiles {
			// Compute hours worked
			var attendances []models.Attendance
			_ = tx.Where("staff_id = ? AND clock_in >= ? AND clock_in <= ?",
				profile.ID, period.StartDate, period.EndDate).Find(&attendances).Error

			var totalHours float64
			for _, a := range attendances {
				if a.ClockOut != nil {
					totalHours += math.Max(0, a.ClockOut.Sub(a.ClockIn).Hours())
				}
			}
			totalHours = math.Round(totalHours*100) / 100

			// Compute total completed sales generated by this staff
			var salesTotal float64
			var orders []models.Order
			_ = tx.Where("(cashier_id = ? OR cashier_id = ?) AND created_at >= ? AND created_at <= ? AND status IN ('completed', 'paid')",
				profile.UserID, profile.ID, period.StartDate, period.EndDate).Find(&orders).Error

			for _, o := range orders {
				salesTotal += o.Total
			}

			// Calculate Commission
			var commissionAmount float64
			for _, rule := range commissionRules {
				if rule.BranchID == nil || (profile.BranchID != nil && *rule.BranchID == *profile.BranchID) {
					if salesTotal >= rule.MinSalesAmount {
						if rule.CommissionPercentage > 0 {
							commissionAmount += (salesTotal * (rule.CommissionPercentage / 100.0))
						}
						commissionAmount += rule.FlatBonus
					}
				}
			}
			commissionAmount = math.Round(commissionAmount*100) / 100

			// Base salary calculation
			baseSalary := profile.BaseSalary
			if baseSalary <= 0 && profile.HourlyRate > 0 {
				baseSalary = math.Round((profile.HourlyRate*totalHours)*100) / 100
			}

			bonus := 0.0
			deductions := 0.0
			netPay := math.Round((baseSalary+commissionAmount+bonus-deductions)*100) / 100
			if netPay < 0 {
				netPay = 0
			}

			staffName := profile.User.FirstName + " " + profile.User.LastName
			if staffName == " " {
				staffName = profile.User.Username
			}
			roleName := "Staff"
			if profile.Role != nil && profile.Role.Name != "" {
				roleName = profile.Role.Name
			}

			record := models.PayrollRecord{
				PeriodID:           period.ID,
				StaffID:            profile.ID,
				StaffName:          staffName,
				StaffRole:          roleName,
				HoursWorked:        totalHours,
				SalesAmount:        salesTotal,
				BaseSalarySnapshot: baseSalary,
				TotalCommission:    commissionAmount,
				Bonus:              bonus,
				Deductions:         deductions,
				NetPay:             netPay,
				IsPaid:             false,
			}

			if err := tx.Create(&record).Error; err != nil {
				return err
			}

			totalGross += (baseSalary + commissionAmount + bonus)
			totalNet += netPay
			staffCount++
		}

		now := time.Now()
		period.IsProcessed = true
		period.ProcessedAt = &now
		period.TotalGross = math.Round(totalGross*100) / 100
		period.TotalNet = math.Round(totalNet*100) / 100
		period.StaffCount = staffCount

		return tx.Save(&period).Error
	})

	if err != nil {
		return nil, err
	}

	return s.GetPayrollPeriod(periodID)
}

func (s *HRService) GetPayslip(recordID string) (*models.PayrollRecord, error) {
	var record models.PayrollRecord
	if err := s.db.Where("id = ?", recordID).Preload("Staff").First(&record).Error; err != nil {
		return nil, errors.New("payslip record not found")
	}
	return &record, nil
}

func (s *HRService) MarkPayslipPaid(recordID string, paymentMethod, paymentRef string) (*models.PayrollRecord, error) {
	var record models.PayrollRecord
	if err := s.db.Where("id = ?", recordID).First(&record).Error; err != nil {
		return nil, errors.New("payslip not found")
	}

	now := time.Now()
	record.IsPaid = true
	record.PaidAt = &now
	if paymentMethod != "" {
		record.PaymentMethod = &paymentMethod
	}
	if paymentRef != "" {
		record.PaymentReference = &paymentRef
	}

	if err := s.db.Save(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

// --- Commission Rules ---

func (s *HRService) ListCommissionRules() ([]models.CommissionRule, error) {
	var rules []models.CommissionRule
	if err := s.db.Order("created_at desc").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *HRService) CreateCommissionRule(rule *models.CommissionRule) (*models.CommissionRule, error) {
	if err := s.db.Create(rule).Error; err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *HRService) DeleteCommissionRule(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.CommissionRule{}).Error
}

// --- Staff Achievements ---

func (s *HRService) ListStaffAchievements() ([]models.StaffAchievement, error) {
	var achievements []models.StaffAchievement
	if err := s.db.Order("awarded_at desc").Find(&achievements).Error; err != nil {
		return nil, err
	}
	return achievements, nil
}

func (s *HRService) CreateStaffAchievement(req *models.StaffAchievement) (*models.StaffAchievement, error) {
	if req.AwardedAt.IsZero() {
		req.AwardedAt = time.Now()
	}
	if err := s.db.Create(req).Error; err != nil {
		return nil, err
	}
	return req, nil
}

func (s *HRService) DeleteStaffAchievement(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.StaffAchievement{}).Error
}

// --- Shift Swapping ---

func (s *HRService) ListShiftSwaps() ([]models.ShiftSwapRequest, error) {
	var swaps []models.ShiftSwapRequest
	if err := s.db.Order("created_at desc").Find(&swaps).Error; err != nil {
		return nil, err
	}
	return swaps, nil
}

func (s *HRService) CreateShiftSwap(swap *models.ShiftSwapRequest) (*models.ShiftSwapRequest, error) {
	swap.Status = "pending"
	if err := s.db.Create(swap).Error; err != nil {
		return nil, err
	}
	return swap, nil
}

func (s *HRService) ApproveShiftSwap(id string) error {
	return s.db.Model(&models.ShiftSwapRequest{}).Where("id = ?", id).Update("status", "approved").Error
}

func (s *HRService) RejectShiftSwap(id string) error {
	return s.db.Model(&models.ShiftSwapRequest{}).Where("id = ?", id).Update("status", "rejected").Error
}

