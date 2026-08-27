package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type CreditHandler struct {
	db      *gorm.DB
	service *services.CreditService
}

func NewCreditHandler(db *gorm.DB, service *services.CreditService) *CreditHandler {
	return &CreditHandler{db: db, service: service}
}

func (h *CreditHandler) getService(c *gin.Context) *services.CreditService {
	return services.NewCreditService(getDB(c, h.db), h.service.GetSMSService())
}

func getTenantIDFromContext(c *gin.Context) (uuid.UUID, error) {
	if tID, exists := c.Get(middleware.ContextKeyTenantID); exists {
		if id, ok := tID.(uuid.UUID); ok {
			return id, nil
		}
		if s, ok := tID.(string); ok {
			return uuid.Parse(s)
		}
	}
	return uuid.Nil, errors.New("tenant not identified")
}

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	if uID, exists := c.Get(middleware.ContextKeyUserID); exists {
		if id, ok := uID.(uuid.UUID); ok {
			return id, nil
		}
		if s, ok := uID.(string); ok {
			return uuid.Parse(s)
		}
	}
	return uuid.Nil, errors.New("user not identified")
}

// GetCreditAccount handles GET /api/v1/customers/:id/credit-account
func (h *CreditHandler) GetCreditAccount(c *gin.Context) {
	tenantID, err := getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not identified"})
		return
	}

	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	details, err := h.getService(c).GetCreditAccountDetails(tenantID, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, details)
}

type SetCreditLimitRequest struct {
	CreditLimit float64 `json:"credit_limit" binding:"required,min=0"`
	DaysToRepay int     `json:"days_to_repay"`
	Notes       string  `json:"notes"`
}

// SetCreditLimit handles POST /api/v1/customers/:id/credit-limit
func (h *CreditHandler) SetCreditLimit(c *gin.Context) {
	tenantID, err := getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not identified"})
		return
	}

	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var req SetCreditLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := h.getService(c).SetCreditLimit(tenantID, customerID, req.CreditLimit, req.DaysToRepay, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Credit limit updated successfully", "account": acc})
}

type CreditDrawdownRequest struct {
	Amount           float64    `json:"amount" binding:"required,gt=0"`
	OrderID          *uuid.UUID `json:"order_id"`
	InstalmentsCount int        `json:"instalments_count"`
	Notes            string     `json:"notes"`
}

// DrawdownCredit handles POST /api/v1/customers/:id/credit/drawdown
func (h *CreditHandler) DrawdownCredit(c *gin.Context) {
	tenantID, err := getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not identified"})
		return
	}

	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var req CreditDrawdownRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := getUserIDFromContext(c)
	var createdByID *uuid.UUID
	if userID != uuid.Nil {
		createdByID = &userID
	}

	tx, err := h.getService(c).DrawdownCredit(tenantID, customerID, req.OrderID, req.Amount, req.InstalmentsCount, createdByID, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Credit drawdown processed successfully", "transaction": tx})
}

type RecordRepaymentRequest struct {
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method"` // cash, momo, card, bank_transfer
	Reference     string  `json:"reference"`
	Notes         string  `json:"notes"`
}

// RecordRepayment handles POST /api/v1/customers/:id/credit/repay
func (h *CreditHandler) RecordRepayment(c *gin.Context) {
	tenantID, err := getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not identified"})
		return
	}

	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var req RecordRepaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := getUserIDFromContext(c)
	var createdByID *uuid.UUID
	if userID != uuid.Nil {
		createdByID = &userID
	}

	if req.PaymentMethod == "" {
		req.PaymentMethod = "cash"
	}

	tx, err := h.getService(c).RecordRepayment(tenantID, customerID, req.Amount, req.PaymentMethod, req.Reference, createdByID, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Repayment recorded successfully", "transaction": tx})
}

// GetOverdueAccounts handles GET /api/v1/credit/overdue
func (h *CreditHandler) GetOverdueAccounts(c *gin.Context) {
	tenantID, err := getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not identified"})
		return
	}

	accounts, err := h.getService(c).GetOverdueAccounts(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"overdue_accounts": accounts, "total_overdue": len(accounts)})
}

// SendRepaymentReminder handles POST /api/v1/customers/:id/credit/send-reminder
func (h *CreditHandler) SendRepaymentReminder(c *gin.Context) {
	tenantID, err := getTenantIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not identified"})
		return
	}

	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	acc, err := h.getService(c).GetOrCreateCreditAccount(tenantID, customerID)
	if err != nil || acc.Balance <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No outstanding balance for this customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Repayment reminder sent successfully", "balance": acc.Balance})
}
