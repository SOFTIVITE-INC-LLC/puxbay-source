package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type SvcService struct {
	db *gorm.DB
}

func NewSvcService(db *gorm.DB) *SvcService {
	return &SvcService{db: db}
}

func (s *SvcService) ListServices() ([]models.Service, error) {
	var services []models.Service

	if err := s.db.Where("is_active = ?", true).Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

type ServiceCreateInput struct {
	Name        string
	Description string
	Price       float64
	DurationMin int
}

func (s *SvcService) CreateService(input ServiceCreateInput) (*models.Service, error) {
	service := models.Service{
		Name:            input.Name,
		Description:     &input.Description,
		Price:           input.Price,
		DurationMinutes: uint(input.DurationMin),
		IsActive:        true,
	}

	if err := s.db.Create(&service).Error; err != nil {
		return nil, err
	}

	return &service, nil
}

func (s *SvcService) ListAppointments() ([]models.Appointment, error) {
	var appointments []models.Appointment

	if err := s.db.Preload("Service").Preload("Customer").Find(&appointments).Error; err != nil {
		return nil, err
	}
	return appointments, nil
}

type AppointmentCreateInput struct {
	ServiceID     string
	CustomerID    string
	StaffMemberID string
	StartTime     string // ISO string
}

func (s *SvcService) CreateAppointment(input AppointmentCreateInput) (*models.Appointment, error) {
	serviceUUID, err := uuid.Parse(input.ServiceID)
	if err != nil {
		return nil, errors.New("invalid service ID")
	}

	customerUUID, err := uuid.Parse(input.CustomerID)
	if err != nil {
		return nil, errors.New("invalid customer ID")
	}

	staffUUID, err := uuid.Parse(input.StaffMemberID)
	if err != nil {
		return nil, errors.New("invalid staff member ID")
	}

	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		return nil, errors.New("invalid start time format")
	}

	appointment := models.Appointment{
		ServiceID:     serviceUUID,
		CustomerID:    &customerUUID,
		StaffMemberID: &staffUUID,
		StartTime:     startTime,
		Status:        "scheduled",
	}

	if err := s.db.Create(&appointment).Error; err != nil {
		return nil, err
	}

	return &appointment, nil
}
