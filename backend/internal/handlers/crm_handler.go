package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type CRMHandler struct {
	db *gorm.DB
}

func NewCRMHandler(db *gorm.DB) *CRMHandler {
	return &CRMHandler{}
}

func (h *CRMHandler) service(c *gin.Context) *services.CRMService {
	return services.NewCRMService(getDB(c, h.db))
}

// --- Loyalty Transactions ---

// ListLoyaltyTransactions returns loyalty transactions for a customer.
func (h *CRMHandler) ListLoyaltyTransactions(c *gin.Context) {
	customerID := c.Query("customer_id")

	transactions, err := h.service(c).ListLoyaltyTransactions(customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch loyalty transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// --- Gift Cards ---

// ListGiftCards returns all gift cards.
func (h *CRMHandler) ListGiftCards(c *gin.Context) {
	status := c.Query("status")

	giftCards, err := h.service(c).ListGiftCards(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch gift cards"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"gift_cards": giftCards})
}

// CreateGiftCard creates a new gift card.
func (h *CRMHandler) CreateGiftCard(c *gin.Context) {
	var req models.GiftCard
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service(c).CreateGiftCard(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gift card"})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// --- CRM Settings ---

// GetCRMSettings returns loyalty program configuration.
func (h *CRMHandler) GetCRMSettings(c *gin.Context) {
	settings, err := h.service(c).GetCRMSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get CRM settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateCRMSettings updates loyalty program configuration.
func (h *CRMHandler) UpdateCRMSettings(c *gin.Context) {
	var req struct {
		PointsPerCurrency float64 `json:"points_per_currency"`
		RedemptionRate    float64 `json:"redemption_rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service(c).UpdateCRMSettings(req.PointsPerCurrency, req.RedemptionRate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update CRM settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// --- Customer Tiers ---

// ListCustomerTiers returns all loyalty tiers.
func (h *CRMHandler) ListTiers(c *gin.Context) {
	tiers, err := h.service(c).ListCustomerTiers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer tiers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tiers": tiers})
}

// CreateCustomerTier creates a new loyalty tier.
func (h *CRMHandler) CreateTier(c *gin.Context) {
	var req models.CustomerTier
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service(c).CreateCustomerTier(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tier"})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// UpdateCustomerTier updates a loyalty tier.
func (h *CRMHandler) UpdateTier(c *gin.Context) {
	id := c.Param("id")

	var req models.CustomerTier
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tier, err := h.service(c).UpdateCustomerTier(id, req)
	if err != nil {
		if err.Error() == "invalid tier ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "tier not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tier"})
		return
	}

	c.JSON(http.StatusOK, tier)
}

// DeleteCustomerTier deletes a loyalty tier.
func (h *CRMHandler) DeleteTier(c *gin.Context) {
	id := c.Param("id")

	if err := h.service(c).DeleteCustomerTier(id); err != nil {
		if err.Error() == "invalid tier ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tier deleted"})
}

// --- Customer Credit ---

// ListCustomerCreditTransactions returns credit transactions for a customer.
func (h *CRMHandler) ListCustomerCreditTransactions(c *gin.Context) {
	customerID := c.Param("customer_id")

	data, err := h.service(c).ListCustomerCreditTransactions(customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch credit transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions":     data.Transactions,
		"outstanding_debt": data.OutstandingDebt,
	})
}

// RecordCustomerPayment records a customer credit payment.
func (h *CRMHandler) RecordCustomerPayment(c *gin.Context) {
	customerIDStr := c.Param("customer_id")

	var req struct {
		Amount    float64 `json:"amount" binding:"required,min=0.01"`
		Reference string  `json:"reference"`
		Notes     string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.PaymentInput{
		Amount:    req.Amount,
		Reference: req.Reference,
		Notes:     req.Notes,
	}

	customer, tx, err := h.service(c).RecordCustomerPayment(customerIDStr, input)
	if err != nil {
		if err.Error() == "invalid customer_id" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "customer not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Payment recorded",
		"new_balance": customer.DebtBalance,
		"transaction": tx,
	})
}

// GetFeedbackList returns all customer feedback.
func (h *CRMHandler) ListFeedback(c *gin.Context) {
	feedback, err := h.service(c).GetFeedbackList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feedback"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feedback": feedback})
}

func (h *CRMHandler) ListStoreCreditTransactions(c *gin.Context) {
	customerID := c.Query("customer_id")
	data, err := h.service(c).ListCustomerCreditTransactions(customerID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch credit transactions"})
		return
	}
	c.JSON(200, gin.H{
		"transactions": data.Transactions,
		"store_credit": data.OutstandingDebt,
	})
}

type CreateFeedbackRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	Rating     uint   `json:"rating" binding:"required,min=1,max=5"`
	Comment    string `json:"comment"`
}

func (h *CRMHandler) CreateFeedback(c *gin.Context) {
	var req CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	db, _ := c.Get("db")
	if db == nil {
		db = c.MustGet("db")
	}

	cID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid customer ID"})
		return
	}

	feedback := models.CustomerFeedback{
		CustomerID: cID,
		Rating:     req.Rating,
	}
	if req.Comment != "" {
		feedback.Comment = &req.Comment
	}

	if err := getDB(c, h.db).Create(&feedback).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to submit feedback"})
		return
	}

	c.JSON(201, feedback)
}

// DeleteFeedback hard-deletes a feedback entry by ID.
func (h *CRMHandler) DeleteFeedback(c *gin.Context) {
	id := c.Param("id")
	db, _ := c.Get("db")
	if db == nil {
		db = c.MustGet("db")
	}
	if err := getDB(c, h.db).Delete(&models.CustomerFeedback{}, "id = ?", id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete feedback"})
		return
	}
	c.JSON(200, gin.H{"message": "Feedback deleted"})
}

// --- Support Tickets ---

type CreateTicketInput struct {
	CustomerID  *uuid.UUID `json:"customer_id,omitempty"`
	Subject     string     `json:"subject" binding:"required"`
	Description string     `json:"description" binding:"required"`
	Priority    string     `json:"priority"`
}

type ReplyTicketInput struct {
	Message string `json:"message" binding:"required"`
}

func (h *CRMHandler) ListTickets(c *gin.Context) {
	db := getDB(c, h.db)
	var tickets []models.SupportTicket

	if err := db.Preload("Customer").Order("created_at desc").Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tickets"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

func (h *CRMHandler) CreateTicket(c *gin.Context) {
	db := getDB(c, h.db)

	var input CreateTicketInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	priority := "medium"
	if input.Priority != "" {
		priority = input.Priority
	}

	ticket := models.SupportTicket{
		TenantScoped: models.TenantScoped{
			Base: models.Base{},
		},
		CustomerID:  input.CustomerID,
		Subject:     input.Subject,
		Description: input.Description,
		Status:      "open",
		Priority:    priority,
	}

	if err := db.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func (h *CRMHandler) GetTicketMessages(c *gin.Context) {
	db := getDB(c, h.db)
	ticketIDStr := c.Param("id")

	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var messages []models.TicketMessage
	if err := db.Where("ticket_id = ?", ticketID).Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *CRMHandler) ReplyTicket(c *gin.Context) {
	db := getDB(c, h.db)
	ticketIDStr := c.Param("id")

	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var input ReplyTicketInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Assuming staff reply for now since this is the admin API
	// Wait, we need to extract User ID from context. If not found, use a fallback zero UUID.
	var senderID uuid.UUID
	userID, exists := c.Get(middleware.ContextKeyUserID)
	if exists {
		senderID = userID.(uuid.UUID)
	}

	msg := models.TicketMessage{
		TicketID: ticketID,
		SenderID: senderID,
		Message:  input.Message,
		IsStaff:  true,
	}

	if err := db.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reply"})
		return
	}

	c.JSON(http.StatusCreated, msg)
}
