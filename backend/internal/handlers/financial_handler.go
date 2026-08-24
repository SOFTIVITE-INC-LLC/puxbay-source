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

type FinancialHandler struct {
	db *gorm.DB
}

func NewFinancialHandler(db *gorm.DB) *FinancialHandler {
	return &FinancialHandler{db: db}
}

func (h *FinancialHandler) service(c *gin.Context) *services.FinancialService {
	return services.NewFinancialService(getDB(c, h.db))
}

func (h *FinancialHandler) getProfileID(c *gin.Context) uuid.UUID {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	var profile models.UserProfile
	if err := h.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&profile).Error; err == nil {
		return profile.ID
	}
	return uuid.Nil
}

func (h *FinancialHandler) ListExpenses(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	expenses, err := h.service(c).ListExpenses(tenantID.(uuid.UUID), branchID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch expenses"})
		return
	}
	c.JSON(http.StatusOK, expenses)
}

type ExpenseCreateRequest struct {
	CategoryID         string  `json:"category_id" binding:"required"`
	Amount             float64 `json:"amount" binding:"required"`
	Date               string  `json:"date" binding:"required"` // ISO string
	Description        string  `json:"description"`
	IsRecurring        bool    `json:"is_recurring"`
	RecurrenceInterval string  `json:"recurrence_interval"`
	ReceiptURL         string  `json:"receipt_url"`
}

func (h *FinancialHandler) CreateExpense(c *gin.Context) {
	profileID := h.getProfileID(c)

	var req ExpenseCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var branchID *uuid.UUID
	if ctxBranchID, ok := middleware.GetBranchID(c); ok {
		branchID = ctxBranchID
	}

	input := services.ExpenseCreateInput{
		CategoryID:         req.CategoryID,
		Amount:             req.Amount,
		Date:               req.Date,
		Description:        req.Description,
		IsRecurring:        req.IsRecurring,
		RecurrenceInterval: req.RecurrenceInterval,
		ReceiptURL:         req.ReceiptURL,
		CreatedByID:        profileID,
		BranchID:           branchID,
	}

	expense, err := h.service(c).CreateExpense(input)
	if err != nil {
		if err.Error() == "invalid date format" || err.Error() == "invalid category ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense"})
		return
	}

	c.JSON(http.StatusCreated, expense)
}

func (h *FinancialHandler) GetProfitAndLoss(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	data, err := h.service(c).GetProfitAndLoss(tenantID.(uuid.UUID), branchID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate P&L"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"gross_revenue":       data.GrossRevenue,
		"cogs":                data.COGS,
		"gross_profit":        data.GrossProfit,
		"total_expenses":      data.TotalExpenses,
		"net_profit":          data.NetProfit,
		"tax_collected":       data.TaxCollected,
		"operating_cash_flow": data.OperatingCashFlow,
		"monthly_data":        data.MonthlyData,
	})
}

func (h *FinancialHandler) GetTaxConfig(c *gin.Context) {
	taxConfig, err := h.service(c).GetTaxConfig()
	if err != nil {
		// Return sensible defaults if not configured
		c.JSON(http.StatusOK, gin.H{"tax_rate": 0.0, "tax_type": "none", "is_active": false})
		return
	}
	c.JSON(http.StatusOK, taxConfig)
}

func (h *FinancialHandler) GetTaxReport(c *gin.Context) {
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	report, err := h.service(c).GetTaxReport(branchID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tax report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tax_summary": report})
}

func (h *FinancialHandler) ListReturns(c *gin.Context) {
	// Dummy implementation for returns
	c.JSON(http.StatusOK, []gin.H{})
}

func (h *FinancialHandler) ProcessReturnRefund(c *gin.Context) {
	// Dummy implementation
	c.JSON(http.StatusOK, gin.H{"message": "Refund processed"})
}

func (h *FinancialHandler) ListExpenseCategories(c *gin.Context) {
	var categories []models.ExpenseCategory
	if err := getDB(c, h.db).Order("name asc").Find(&categories).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch expense categories"})
		return
	}
	c.JSON(200, gin.H{"categories": categories})
}

type ExpenseCategoryRequest struct {
	Name          string  `json:"name" binding:"required"`
	Type          string  `json:"type" binding:"required"`
	Description   string  `json:"description"`
	MonthlyBudget float64 `json:"monthly_budget"`
}

func (h *FinancialHandler) CreateExpenseCategory(c *gin.Context) {
	var req ExpenseCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	cat := models.ExpenseCategory{
		Name:          req.Name,
		Type:          req.Type,
		MonthlyBudget: req.MonthlyBudget,
	}
	if req.Description != "" {
		cat.Description = &req.Description
	}

	if err := getDB(c, h.db).Create(&cat).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(201, cat)
}

type ExpenseUpdateRequest struct {
	CategoryID         string  `json:"category_id"`
	Amount             float64 `json:"amount"`
	Date               string  `json:"date"`
	Description        string  `json:"description"`
	IsRecurring        *bool   `json:"is_recurring"`
	RecurrenceInterval *string `json:"recurrence_interval"`
	ReceiptURL         *string `json:"receipt_url"`
}

func (h *FinancialHandler) UpdateExpense(c *gin.Context) {
	expenseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID"})
		return
	}

	var req ExpenseUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense, err := h.service(c).UpdateExpense(expenseID, services.ExpenseUpdateInput{
		CategoryID:         req.CategoryID,
		Amount:             req.Amount,
		Date:               req.Date,
		Description:        req.Description,
		IsRecurring:        req.IsRecurring,
		RecurrenceInterval: req.RecurrenceInterval,
		ReceiptURL:         req.ReceiptURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expense)
}

func (h *FinancialHandler) DeleteExpense(c *gin.Context) {
	expenseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID"})
		return
	}

	if err := h.service(c).DeleteExpense(expenseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *FinancialHandler) UpdateTaxConfig(c *gin.Context) {
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var config models.TaxConfiguration
	gdb := getDB(c, h.db)
	if err := gdb.First(&config).Error; err != nil {
		// No config exists, create one
		if err := gdb.Create(&config).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tax config"})
			return
		}
	}

	if err := gdb.Model(&config).Updates(body).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tax config"})
		return
	}

	c.JSON(http.StatusOK, config)
}
