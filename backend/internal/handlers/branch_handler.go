package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/services"
	"github.com/softivite/puxbay/internal/utils"
	"gorm.io/gorm"
)

// BranchHandler handles branch management.
type BranchHandler struct {
	db *gorm.DB
}

// NewBranchHandler creates a new branch handler.
func NewBranchHandler(db *gorm.DB) *BranchHandler {
	return &BranchHandler{db: db}
}

func (h *BranchHandler) service(c *gin.Context) *services.BranchService {
	return services.NewBranchService(getDB(c, h.db))
}

// List returns all branches for the tenant.
// GET /api/v1/branches
func (h *BranchHandler) List(c *gin.Context) {
	p := utils.GetPagination(c)

	branches, total, err := h.service(c).ListBranches(p.Limit, p.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  branches,
		"total": total,
		"page":  p.Page,
		"limit": p.Limit,
	})
}

// CreateRequest defines the payload for creating a branch.
type BranchCreateRequest struct {
	Name           string `json:"name" binding:"required"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
	PrimaryColor   string `json:"primary_color"`
	CurrencySymbol string `json:"currency_symbol"`
	CurrencyCode   string `json:"currency_code"`
	BranchType     string `json:"branch_type"`
}

// Create creates a new branch.
// POST /api/v1/branches
func (h *BranchHandler) Create(c *gin.Context) {
	var req BranchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	tenantIDVal, _ := c.Get("tenant_id")
	tenantID, _ := tenantIDVal.(uuid.UUID)

	input := services.BranchCreateInput{
		TenantID:       tenantID,
		Name:           req.Name,
		Address:        req.Address,
		Phone:          req.Phone,
		PrimaryColor:   req.PrimaryColor,
		CurrencySymbol: req.CurrencySymbol,
		CurrencyCode:   req.CurrencyCode,
		BranchType:     req.BranchType,
	}

	branch, err := h.service(c).CreateBranch(input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "active subscription required to create a branch" || strings.Contains(err.Error(), "branch limit reached") || err.Error() == "tenant ID is missing from request context" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, branch)
}

// Get returns a specific branch.
// GET /api/v1/branches/:id
func (h *BranchHandler) Get(c *gin.Context) {
	id := c.Param("id")

	branch, err := h.service(c).GetBranch(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, branch)
}

// UpdateRequest defines the payload for updating a branch.
type BranchUpdateRequest struct {
	Name              string  `json:"name"`
	Address           string  `json:"address"`
	Phone             string  `json:"phone"`
	PrimaryColor      string  `json:"primary_color"`
	CurrencySymbol    string  `json:"currency_symbol"`
	CurrencyCode      string  `json:"currency_code"`
	ReceiptHeader     *string `json:"receipt_header"`
	ReceiptFooter     *string `json:"receipt_footer"`
	LowStockThreshold *uint   `json:"low_stock_threshold"`
}

// Update modifies an existing branch.
// PUT /api/v1/branches/:id
func (h *BranchHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req BranchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	input := services.BranchUpdateInput{
		Name:              req.Name,
		Address:           req.Address,
		Phone:             req.Phone,
		PrimaryColor:      req.PrimaryColor,
		CurrencySymbol:    req.CurrencySymbol,
		CurrencyCode:      req.CurrencyCode,
		ReceiptHeader:     req.ReceiptHeader,
		ReceiptFooter:     req.ReceiptFooter,
		LowStockThreshold: req.LowStockThreshold,
	}

	branch, err := h.service(c).UpdateBranch(id, input)
	if err != nil {
		if err.Error() == "branch not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branch"})
		return
	}

	c.JSON(http.StatusOK, branch)
}

// Delete removes a branch.
// DELETE /api/v1/branches/:id
func (h *BranchHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service(c).DeleteBranch(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete branch"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// NetworkMetrics returns global KPIs across all branches.
// GET /api/v1/branches/network/metrics
func (h *BranchHandler) NetworkMetrics(c *gin.Context) {
	tenantIDVal, _ := c.Get("tenant_id")
	tenantID, _ := tenantIDVal.(uuid.UUID)

	metrics, err := h.service(c).GetNetworkMetrics(tenantID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch network metrics"})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// BranchMetrics returns KPIs for a specific branch.
// GET /api/v1/branches/:id/metrics
func (h *BranchHandler) BranchMetrics(c *gin.Context) {
	id := c.Param("id")

	metrics, err := h.service(c).GetBranchMetrics(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branch metrics"})
		return
	}
	c.JSON(http.StatusOK, metrics)
}
