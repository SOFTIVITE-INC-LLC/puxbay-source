package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type GiftCardHandler struct {
	db *gorm.DB
}

func NewGiftCardHandler(db *gorm.DB) *GiftCardHandler {
	return &GiftCardHandler{db: db}
}

func (h *GiftCardHandler) getDB(c *gin.Context) *gorm.DB {
	return getDB(c, h.db)
}

// List returns all gift cards.
func (h *GiftCardHandler) List(c *gin.Context) {
	var cards []models.GiftCard
	query := h.getDB(c)
	if err := query.Order("created_at DESC").Find(&cards).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch gift cards"})
		return
	}
	c.JSON(http.StatusOK, cards)
}

// Get returns a single gift card.
func (h *GiftCardHandler) Get(c *gin.Context) {
	id := c.Param("id")
	cardID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gift card ID"})
		return
	}
	var card models.GiftCard
	if err := h.getDB(c).First(&card, "id = ?", cardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gift card not found"})
		return
	}
	c.JSON(http.StatusOK, card)
}

// Create creates a new gift card.
func (h *GiftCardHandler) Create(c *gin.Context) {
	var input struct {
		InitialBalance float64 `json:"initial_balance"`
		PurchaserID    string  `json:"purchaser_id"`
		ExpiresAt      string  `json:"expires_at"`
		CustomCode     string  `json:"custom_code"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var card models.GiftCard
	card.InitialBalance = input.InitialBalance
	card.CurrentBalance = input.InitialBalance
	card.IsActive = true
	card.Status = "active"

	if input.CustomCode != "" {
		card.Code = input.CustomCode
	} else {
		card.Code = "GC-" + uuid.New().String()[:8]
	}

	if input.PurchaserID != "" {
		pid, err := uuid.Parse(input.PurchaserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchaser ID"})
			return
		}
		card.PurchaserID = &pid
	}

	if input.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", input.ExpiresAt)
		if err != nil {
			t, err = time.Parse(time.RFC3339, input.ExpiresAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expiry date format"})
				return
			}
		}
		card.ExpiresAt = &t
	}

	if err := h.getDB(c).Create(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gift card"})
		return
	}
	c.JSON(http.StatusCreated, card)
}

// Redeem redeems a gift card by code.
func (h *GiftCardHandler) Redeem(c *gin.Context) {
	var req struct {
		Code   string  `json:"code" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var card models.GiftCard
	if err := h.getDB(c).Where("code = ? AND is_active = ?", req.Code, true).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gift card not found or inactive"})
		return
	}

	redeemAmount := req.Amount
	if card.CurrentBalance < req.Amount {
		redeemAmount = card.CurrentBalance
	}

	card.CurrentBalance -= redeemAmount
	if card.CurrentBalance <= 0 {
		card.IsActive = false
	}

	if err := h.getDB(c).Save(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to redeem gift card"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "redeemed", "balance": card.CurrentBalance})
}

// CheckBalance checks the balance of a gift card by code.
func (h *GiftCardHandler) CheckBalance(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is required"})
		return
	}

	var card models.GiftCard
	if err := h.getDB(c).Where("code = ?", code).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gift card not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"gift_card": card})
}

// Disable disables a gift card.
func (h *GiftCardHandler) Disable(c *gin.Context) {
	id := c.Param("id")
	cardID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gift card ID"})
		return
	}

	var card models.GiftCard
	if err := h.getDB(c).First(&card, "id = ?", cardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gift card not found"})
		return
	}

	card.Status = "disabled"
	card.IsActive = false

	if err := h.getDB(c).Save(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable gift card"})
		return
	}

	c.JSON(http.StatusOK, card)
}
