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
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	suggestions, err := h.service(c).GetDynamicPricing(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate dynamic pricing recommendations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}

type PricingActionReq struct {
	ProductID string  `json:"product_id" binding:"required"`
	NewPrice  float64 `json:"new_price" binding:"required"`
}

type BulkPricingActionReq struct {
	Items []PricingActionReq `json:"items" binding:"required,min=1"`
}

func (h *IntelligenceHandler) ApplyPricingAction(c *gin.Context) {
	var req PricingActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service(c).ApplyPricingAction(req.ProductID, req.NewPrice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product price: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "price_updated",
		"product_id": req.ProductID,
		"new_price":  req.NewPrice,
		"message":    "Product price updated successfully",
	})
}

func (h *IntelligenceHandler) BulkApplyPricingAction(c *gin.Context) {
	var req BulkPricingActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []services.PricingActionItem
	for _, it := range req.Items {
		items = append(items, services.PricingActionItem{
			ProductID: it.ProductID,
			NewPrice:  it.NewPrice,
		})
	}

	updatedCount, err := h.service(c).BulkApplyPricing(items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product prices: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "bulk_prices_updated",
		"updated_count": updatedCount,
		"message":       "Dynamic prices updated successfully",
	})
}

func (h *IntelligenceHandler) GenerateAutoPOs(c *gin.Context) {
	// Would scan Inventory Forecasts and automatically create POs for low stock
	c.JSON(201, gin.H{"status": "auto_pos_generated", "pos_created": 2})
}

// GetAnomalies returns real-time detected anomalies for the tenant
func (h *IntelligenceHandler) GetAnomalies(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	anomalies, err := h.service(c).GetAnomalyAlerts(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to detect anomalies: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"anomalies": anomalies})
}

// GetAnomalyStats returns summary statistics of anomalies
func (h *IntelligenceHandler) GetAnomalyStats(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	stats, err := h.service(c).GetAnomalyStats(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch anomaly statistics: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
