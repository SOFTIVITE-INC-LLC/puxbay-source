package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type ShiftService struct {
	db *gorm.DB
}

func NewShiftService(db *gorm.DB) *ShiftService {
	return &ShiftService{db: db}
}

func (s *ShiftService) OpenShift(branchID, userID uuid.UUID, openingFloat float64, notes string) (*models.CashRegisterShift, error) {
	// Check if there is already an open shift for this branch
	var count int64
	s.db.Model(&models.CashRegisterShift{}).Where("branch_id = ? AND status = ?", branchID, "open").Count(&count)
	if count > 0 {
		return nil, errors.New("there is already an open shift for this branch")
	}

	shift := models.CashRegisterShift{
		BranchID:     branchID,
		OpenedByID:   userID,
		OpenedAt:     time.Now(),
		OpeningFloat: openingFloat,
		Status:       "open",
		Notes:        notes,
	}

	if err := s.db.Create(&shift).Error; err != nil {
		return nil, err
	}

	return &shift, nil
}

func (s *ShiftService) CloseShift(shiftID, userID uuid.UUID, actualCash float64, notes string) (*models.CashRegisterShift, error) {
	var shift models.CashRegisterShift
	if err := s.db.Where("id = ?", shiftID).First(&shift).Error; err != nil {
		return nil, errors.New("shift not found")
	}

	if shift.Status == "closed" {
		return nil, errors.New("shift is already closed")
	}

	// Calculate expected cash based on OpeningFloat + Cash payments during this shift
	// (Querying orders/payments is simplified here)
	var cashPayments float64
	s.db.Model(&models.Order{}).
		Where("branch_id = ? AND created_at >= ? AND payment_method = ?", shift.BranchID, shift.OpenedAt, "cash").
		Select("COALESCE(SUM(amount_paid), 0)").
		Scan(&cashPayments)

	expectedCash := shift.OpeningFloat + cashPayments
	variance := actualCash - expectedCash
	now := time.Now()

	updates := map[string]interface{}{
		"status":        "closed",
		"closed_by_id":  userID,
		"closed_at":     &now,
		"expected_cash": expectedCash,
		"actual_cash":   actualCash,
		"variance":      variance,
	}

	if notes != "" {
		updates["notes"] = shift.Notes + "\nClose Notes: " + notes
	}

	if err := s.db.Model(&shift).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Fetch updated
	s.db.Where("id = ?", shiftID).First(&shift)
	return &shift, nil
}
