package services

import (
	"errors"
	"time"

	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type FNBService struct {
	db *gorm.DB
}

func NewFNBService(db *gorm.DB) *FNBService {
	return &FNBService{db: db}
}

func (s *FNBService) ListTables() ([]models.DiningTable, error) {
	var tables []models.DiningTable
	if err := s.db.Where("is_active = ?", true).Find(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

func (s *FNBService) CreateTable(table *models.DiningTable) error {
	return s.db.Create(table).Error
}

func (s *FNBService) UpdateTableStatus(id string, status string) (*models.DiningTable, error) {
	var table models.DiningTable
	if err := s.db.First(&table, "id = ?", id).Error; err != nil {
		return nil, errors.New("table not found")
	}

	table.Status = status
	if err := s.db.Save(&table).Error; err != nil {
		return nil, err
	}

	return &table, nil
}

func (s *FNBService) ListKDS() ([]models.KDSTicket, error) {
	var tickets []models.KDSTicket
	if err := s.db.Where("status IN ?", []string{"pending", "preparing"}).Preload("Order").Preload("Table").Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (s *FNBService) AdvanceTicketStatus(id string) (*models.KDSTicket, error) {
	var ticket models.KDSTicket
	if err := s.db.First(&ticket, "id = ?", id).Error; err != nil {
		return nil, errors.New("ticket not found")
	}

	nextStatusMap := map[string]string{
		"pending":   "preparing",
		"preparing": "ready",
		"ready":     "served",
	}

	nextStatus, exists := nextStatusMap[ticket.Status]
	if !exists {
		return nil, errors.New("cannot advance from current status")
	}

	ticket.Status = nextStatus
	now := time.Now()

	if nextStatus == "preparing" {
		ticket.StartedAt = &now
	} else if nextStatus == "ready" {
		ticket.CompletedAt = &now
	}

	if err := s.db.Save(&ticket).Error; err != nil {
		return nil, err
	}

	return &ticket, nil
}
