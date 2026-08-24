package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type StaffHandler struct {
	db           *gorm.DB
	authService  *services.AuthService
	smsService   *services.SMSService
	emailService *services.EmailService
	rootDomain   string
}

func NewStaffHandler(db *gorm.DB, auth *services.AuthService, sms *services.SMSService, email *services.EmailService, rootDomain string) *StaffHandler {
	return &StaffHandler{db: db, authService: auth, smsService: sms, emailService: email, rootDomain: rootDomain}
}

func (h *StaffHandler) service(c *gin.Context) *services.StaffService {
	return services.NewStaffService(getDB(c, h.db), h.authService, h.smsService, h.emailService, h.rootDomain)
}

func (h *StaffHandler) List(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tid := tenantID.(uuid.UUID)

	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	profiles, err := h.service(c).ListStaff(tid, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch staff members"})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

type StaffCreateRequest struct {
	Username  string  `json:"username" binding:"required"`
	Email     string  `json:"email" binding:"required,email"`
	Password  string  `json:"password"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     string  `json:"phone"`
	Role      string  `json:"role" binding:"required"`
	BranchID  *string `json:"branch_id"`
}

func (h *StaffHandler) Create(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tid := tenantID.(uuid.UUID)

	var req StaffCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ShouldBindJSON error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.StaffCreateInput{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      req.Role,
	}

	if req.BranchID != nil && *req.BranchID != "" {
		parsed, _ := uuid.Parse(*req.BranchID)
		input.BranchID = &parsed
	}

	profile, err := h.service(c).CreateStaff(tid, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, profile)
}

func (h *StaffHandler) Get(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tid := tenantID.(uuid.UUID)
	id := c.Param("id")

	profile, err := h.service(c).GetStaff(id, tid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *StaffHandler) Update(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tid := tenantID.(uuid.UUID)
	id := c.Param("id")

	var req StaffCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.StaffUpdateInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
	}

	if req.BranchID != nil && *req.BranchID != "" {
		parsed, _ := uuid.Parse(*req.BranchID)
		input.BranchID = &parsed
	}

	profile, err := h.service(c).UpdateStaff(id, tid, input)
	if err != nil {
		if err.Error() == "staff not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update staff: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *StaffHandler) Delete(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tid := tenantID.(uuid.UUID)
	id := c.Param("id")

	if err := h.service(c).DeleteStaff(id, tid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete staff"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
