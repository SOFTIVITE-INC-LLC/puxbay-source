package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletAPIHandler struct {
	db *gorm.DB
}

func NewWalletAPIHandler(db *gorm.DB) *WalletAPIHandler {
	return &WalletAPIHandler{db: db}
}

func (h *WalletAPIHandler) service(c *gin.Context) *services.WalletService {
	return services.NewWalletService(getDB(c, h.db))
}

func (h *WalletAPIHandler) Dashboard(c *gin.Context) {
	customerPhone := c.Query("phone")
	customerIDStr := c.Query("customer_id")

	data, err := h.service(c).GetDashboard(customerIDStr, customerPhone)
	if err != nil {
		if err.Error() == "invalid customer_id" || err.Error() == "phone or customer_id required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer":        data.Customer,
		"recent_orders":   data.RecentOrders,
		"gift_cards":      data.GiftCards,
		"loyalty_history": data.LoyaltyHistory,
	})
}

func (h *WalletAPIHandler) LookupCustomer(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.service(c).LookupCustomer(req.Phone)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer_id":    customer.ID,
		"name":           customer.Name,
		"loyalty_points": customer.LoyaltyPts,
	})
}

func (h *WalletAPIHandler) GetGiftCards(c *gin.Context) {
	customerIDStr := c.Param("customer_id")

	giftCards, err := h.service(c).GetGiftCards(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"gift_cards": giftCards})
}

func (h *WalletAPIHandler) CreateGiftCard(c *gin.Context) {
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

func (h *WalletAPIHandler) GetLoyaltyTransactions(c *gin.Context) {
	customerIDStr := c.Param("customer_id")

	transactions, err := h.service(c).GetLoyaltyTransactions(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

func (h *WalletAPIHandler) AdjustLoyaltyPoints(c *gin.Context) {
	customerIDStr := c.Param("customer_id")

	var req struct {
		Points      float64 `json:"points" binding:"required"`
		Description string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.LoyaltyAdjustmentInput{
		Points:      req.Points,
		Description: req.Description,
	}

	customer, transaction, err := h.service(c).AdjustLoyaltyPoints(customerIDStr, input)
	if err != nil {
		if err.Error() == "invalid customer_id" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Points adjusted successfully",
		"new_balance": customer.LoyaltyPts,
		"transaction": transaction,
	})
}

func (h *WalletAPIHandler) GetBalance(c *gin.Context) {
	customerIDStr := c.Query("customer_id")

	customer, err := h.service(c).GetBalance(customerIDStr)
	if err != nil {
		if err.Error() == "customer_id required" || err.Error() == "invalid customer_id" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"loyalty_points": customer.LoyaltyPts,
		"store_credit":   customer.StoreCredit,
		"debt_balance":   customer.DebtBalance,
	})
}

func (h *WalletAPIHandler) ListTransactions(c *gin.Context) {
	customerID := c.Query("customer_id")
	var transactions []models.StoreCreditTransaction
	if err := getDB(c, h.db).Where("customer_id = ?", customerID).Order("created_at desc").Find(&transactions).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch wallet transactions"})
		return
	}
	c.JSON(200, gin.H{"transactions": transactions})
}

type TopUpRequest struct {
	CustomerID string  `json:"customer_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Reference  string  `json:"reference"`
}

func (h *WalletAPIHandler) TopUp(c *gin.Context) {
	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err := getDB(c, h.db).Transaction(func(tx *gorm.DB) error {
		// Gap #13: Row-level lock to prevent concurrent topup overwrites
		var customer models.Customer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.CustomerID).First(&customer).Error; err != nil {
			return err
		}

		// Use atomic DB expression instead of in-memory arithmetic
		if err := tx.Model(&customer).Update("store_credit", gorm.Expr("store_credit + ?", req.Amount)).Error; err != nil {
			return err
		}

		creditTx := models.StoreCreditTransaction{
			CustomerID:      customer.ID,
			Amount:          req.Amount,
			TransactionType: "topup",
			Reference:       req.Reference,
			Notes:           "Wallet TopUp",
		}
		return tx.Create(&creditTx).Error
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to process top-up: " + err.Error()})
		return
	}
	c.JSON(201, gin.H{"status": "topup_successful"})
}

type TransferRequest struct {
	FromCustomerID string  `json:"from_customer_id" binding:"required"`
	ToCustomerID   string  `json:"to_customer_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
}

func (h *WalletAPIHandler) Transfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := getDB(c, h.db).Transaction(func(tx *gorm.DB) error {
		// Gap #12: Row-level locks to prevent double-spend on concurrent transfers
		var sender models.Customer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.FromCustomerID).First(&sender).Error; err != nil {
			return err
		}
		if sender.StoreCredit < req.Amount {
			return fmt.Errorf("insufficient funds")
		}

		var receiver models.Customer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.ToCustomerID).First(&receiver).Error; err != nil {
			return err
		}

		// Use atomic DB expressions instead of in-memory arithmetic
		if err := tx.Model(&sender).Update("store_credit", gorm.Expr("store_credit - ?", req.Amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(&receiver).Update("store_credit", gorm.Expr("store_credit + ?", req.Amount)).Error; err != nil {
			return err
		}

		outTx := models.StoreCreditTransaction{
			CustomerID:      sender.ID,
			Amount:          -req.Amount,
			TransactionType: "transfer_out",
			Notes:           "Transfer to " + receiver.Name,
		}
		inTx := models.StoreCreditTransaction{
			CustomerID:      receiver.ID,
			Amount:          req.Amount,
			TransactionType: "transfer_in",
			Notes:           "Transfer from " + sender.Name,
		}

		if err := tx.Create(&outTx).Error; err != nil {
			return err
		}
		return tx.Create(&inTx).Error
	})

	if err != nil {
		c.JSON(400, gin.H{"error": "Transfer failed: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "transfer_successful"})
}

func (h *WalletAPIHandler) AdjustStoreCredit(c *gin.Context) {
	customerIDStr := c.Param("customer_id")

	var req struct {
		Amount float64 `json:"amount" binding:"required"`
		Note   string  `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.StoreCreditAdjustmentInput{
		Amount: req.Amount,
		Note:   req.Note,
	}

	customer, err := h.service(c).AdjustStoreCredit(customerIDStr, input)
	if err != nil {
		if err.Error() == "invalid customer_id" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Store credit adjusted successfully",
		"store_credit":  customer.StoreCredit,
	})
}
