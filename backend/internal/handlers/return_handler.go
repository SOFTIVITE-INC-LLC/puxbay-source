package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type ReturnHandler struct {
	db *gorm.DB
}

func NewReturnHandler(db *gorm.DB) *ReturnHandler {
	return &ReturnHandler{db: db}
}

func (h *ReturnHandler) service(c *gin.Context) *services.ReturnService {
	tenantID, _ := uuid.Parse(c.GetString(string(middleware.ContextKeyTenantID)))
	return services.NewReturnService(getDB(c, h.db), tenantID)
}

// List returns all return requests.
func (h *ReturnHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	params := services.ReturnListParams{
		BranchID: middleware.ResolveBranchID(c, c.Query("branch_id")),
		Status:   c.Query("status"),
		Limit:    limit,
		Offset:   offset,
	}

	returns, total, err := h.service(c).ListReturns(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch returns"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"returns": returns, "total": total})
}

// Create creates a new return request.
func (h *ReturnHandler) Create(c *gin.Context) {
	var req struct {
		OrderID      uuid.UUID           `json:"order_id" binding:"required"`
		CustomerID   *uuid.UUID          `json:"customer_id"`
		Reason       string              `json:"reason" binding:"required"`
		ReasonDetail string              `json:"reason_detail"`
		RefundMethod string              `json:"refund_method"`
		RefundAmount float64             `json:"refund_amount"`
		Items        []models.ReturnItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var branchID *uuid.UUID
	if ctxBranchID, ok := middleware.GetBranchID(c); ok {
		branchID = ctxBranchID
	}

	input := services.ReturnCreateInput{
		OrderID:      req.OrderID,
		BranchID:     branchID,
		CustomerID:   req.CustomerID,
		Reason:       req.Reason,
		ReasonDetail: req.ReasonDetail,
		RefundMethod: req.RefundMethod,
		RefundAmount: req.RefundAmount,
		Items:        req.Items,
	}

	ret, err := h.service(c).CreateReturn(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create return request"})
		return
	}

	c.JSON(http.StatusCreated, ret)
}

// Get returns a single return request.
func (h *ReturnHandler) Get(c *gin.Context) {
	id := c.Param("id")

	ret, err := h.service(c).GetReturn(id)
	if err != nil {
		if err.Error() == "invalid return ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "return not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch return"})
		return
	}

	c.JSON(http.StatusOK, ret)
}

// Approve approves a return request and restocks eligible items.
func (h *ReturnHandler) Approve(c *gin.Context) {
	id := c.Param("id")

	ret, err := h.service(c).ApproveReturn(id)
	if err != nil {
		if err.Error() == "invalid return ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "return not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve return"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Return approved", "return": ret})
}

// Reject rejects a return request.
func (h *ReturnHandler) Reject(c *gin.Context) {
	id := c.Param("id")

	ret, err := h.service(c).RejectReturn(id)
	if err != nil {
		if err.Error() == "invalid return ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "return not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject return"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Return rejected", "return": ret})
}

// ProcessRefund completes the refund for an approved return.
func (h *ReturnHandler) ProcessRefund(c *gin.Context) {
	id := c.Param("id")

	ret, netRefund, err := h.service(c).ProcessRefund(id)
	if err != nil {
		if err.Error() == "invalid return ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "approved return not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process refund"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Refund processed",
		"net_refund": netRefund,
		"return":     ret,
	})
}
