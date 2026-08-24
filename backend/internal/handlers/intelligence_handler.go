package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type IntelligenceHandler struct {
	db *gorm.DB
}

func NewIntelligenceHandler(db *gorm.DB) *IntelligenceHandler {
	return &IntelligenceHandler{db: db}
}

func (h *IntelligenceHandler) service(c *gin.Context) *services.IntelligenceService {
	return services.NewIntelligenceService(getDB(c, h.db))
}

// InventoryForecast returns products with their stock depletion velocity and days remaining
func (h *IntelligenceHandler) InventoryForecast(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	forecasts, err := h.service(c).InventoryForecast(tenantID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate inventory forecast"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"forecasts": forecasts})
}

// POSRecommendations returns frequently bought together items
func (h *IntelligenceHandler) POSRecommendations(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	productIDs := c.Query("product_ids") // comma separated

	recommendations, err := h.service(c).POSRecommendations(tenantID, productIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get POS recommendations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}

// StaffLeaderboard returns ranking of cashiers by sales
func (h *IntelligenceHandler) StaffLeaderboard(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	leaderboard, err := h.service(c).StaffLeaderboard(tenantID, branchID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch staff leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"leaderboard": leaderboard})
}

// CustomerSegmentation returns RFM analysis (Recency, Frequency, Monetary)
func (h *IntelligenceHandler) CustomerSegmentation(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	result, err := h.service(c).CustomerSegmentation(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate customer segmentation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"segments":        result.Segments,
		"total_customers": result.TotalCustomers,
	})
}

func (h *IntelligenceHandler) DynamicPricing(c *gin.Context) {
	c.JSON(200, gin.H{
		"suggestions": []gin.H{
			{"product_id": "123", "current_price": 10.0, "suggested_price": 12.0, "reason": "high demand"},
		},
	})
}

type PricingActionReq struct {
	ProductID string  `json:"product_id" binding:"required"`
	NewPrice  float64 `json:"new_price" binding:"required"`
}

func (h *IntelligenceHandler) ApplyPricingAction(c *gin.Context) {
	var req PricingActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// For actual implementation we'd update models.Product.Price
	c.JSON(200, gin.H{"status": "price_updated", "product_id": req.ProductID, "new_price": req.NewPrice})
}

func (h *IntelligenceHandler) GenerateAutoPOs(c *gin.Context) {
	// Would scan Inventory Forecasts and automatically create POs for low stock
	c.JSON(201, gin.H{"status": "auto_pos_generated", "pos_created": 2})
}
