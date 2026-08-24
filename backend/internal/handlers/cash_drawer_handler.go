package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type CashDrawerHandler struct {
	db *gorm.DB
}

func NewCashDrawerHandler(db *gorm.DB) *CashDrawerHandler {
	return &CashDrawerHandler{db: db}
}

func (h *CashDrawerHandler) getDB(c *gin.Context) *gorm.DB {
	return getDB(c, h.db)
}

func (h *CashDrawerHandler) getUserID(c *gin.Context) uuid.UUID {
	rawUserID, _ := c.Get(middleware.ContextKeyUserID)
	userID, _ := rawUserID.(uuid.UUID)
	return userID
}

// Open opens a new cash drawer session.
func (h *CashDrawerHandler) Open(c *gin.Context) {
	var req struct {
		BranchID     string  `json:"branch_id" binding:"required"`
		OpeningFloat float64 `json:"opening_float"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branchID, err := uuid.Parse(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	userID := h.getUserID(c)
	session := models.CashDrawerSession{
		BranchID:       branchID,
		UserID:         userID,
		OpeningBalance: req.OpeningFloat,
	}

	if err := h.getDB(c).Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open cash drawer"})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// Close closes an active cash drawer session.
func (h *CashDrawerHandler) Close(c *gin.Context) {
	var req struct {
		SessionID    string  `json:"session_id" binding:"required"`
		ClosingFloat float64 `json:"closing_float"`
		Notes        string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var session models.CashDrawerSession
	if err := h.getDB(c).First(&session, "id = ?", sessionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	now := time.Now()
	notes := req.Notes
	session.ClosedAt = &now
	session.ClosingBalance = req.ClosingFloat
	if notes != "" {
		session.Notes = &notes
	}

	if err := h.getDB(c).Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close cash drawer"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetReport returns a cash drawer session report (EOD).
func (h *CashDrawerHandler) GetReport(c *gin.Context) {
	id := c.Param("id")
	sessionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var session models.CashDrawerSession
	if err := h.getDB(c).First(&session, "id = ?", sessionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session":    session,
		"opening":    session.OpeningBalance,
		"closing":    session.ClosingBalance,
		"difference": session.ClosingBalance - session.OpeningBalance,
	})
}

// List returns all cash drawer sessions.
func (h *CashDrawerHandler) List(c *gin.Context) {
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	var sessions []models.CashDrawerSession
	query := h.getDB(c)
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	query = query.Order("created_at DESC").Limit(50)

	if err := query.Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sessions"})
		return
	}

	c.JSON(http.StatusOK, sessions)
}
