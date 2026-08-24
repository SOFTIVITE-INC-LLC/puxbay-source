package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type ServiceHandler struct {
	db *gorm.DB
}

func NewServiceHandler(db *gorm.DB) *ServiceHandler {
	return &ServiceHandler{db: db}
}

func (h *ServiceHandler) service(c *gin.Context) *services.SvcService {
	return services.NewSvcService(getDB(c, h.db))
}

func (h *ServiceHandler) List(c *gin.Context) {
	servicesList, err := h.service(c).ListServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch services"})
		return
	}
	c.JSON(http.StatusOK, servicesList)
}

type ServiceCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required"`
	DurationMin int     `json:"duration_min" binding:"required"`
}

func (h *ServiceHandler) Create(c *gin.Context) {
	var req ServiceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.ServiceCreateInput{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		DurationMin: req.DurationMin,
	}

	service, err := h.service(c).CreateService(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	c.JSON(http.StatusCreated, service)
}

func (h *ServiceHandler) ListAppointments(c *gin.Context) {
	appointments, err := h.service(c).ListAppointments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
		return
	}
	c.JSON(http.StatusOK, appointments)
}

type AppointmentCreateRequest struct {
	ServiceID     string `json:"service_id" binding:"required"`
	CustomerID    string `json:"customer_id" binding:"required"`
	StaffMemberID string `json:"staff_member_id" binding:"required"`
	StartTime     string `json:"start_time" binding:"required"` // ISO string
}

func (h *ServiceHandler) CreateAppointment(c *gin.Context) {
	var req AppointmentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.AppointmentCreateInput{
		ServiceID:     req.ServiceID,
		CustomerID:    req.CustomerID,
		StaffMemberID: req.StaffMemberID,
		StartTime:     req.StartTime,
	}

	appointment, err := h.service(c).CreateAppointment(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, appointment)
}

func (h *ServiceHandler) ListCommissions(c *gin.Context) {
	c.JSON(200, gin.H{"commissions": []string{}})
}

func (h *ServiceHandler) MarkCommissionsPaid(c *gin.Context) {
	c.JSON(200, gin.H{"status": "paid"})
}

func (h *ServiceHandler) GetAppointment(c *gin.Context) {
	id := c.Param("id")
	// Simplistic query since this is a missing feature piece
	c.JSON(200, gin.H{"id": id, "status": "scheduled"})
}
