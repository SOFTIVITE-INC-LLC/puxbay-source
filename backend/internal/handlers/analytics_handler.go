package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type AnalyticsHandler struct {
	db *gorm.DB
}

func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

func (h *AnalyticsHandler) service(c *gin.Context) *services.AnalyticsService {
	return services.NewAnalyticsService(getDB(c, h.db))
}

// Dashboard returns aggregated metrics and recent transactions
func (h *AnalyticsHandler) Dashboard(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	result, err := h.service(c).DashboardOverview(tenantID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dashboard overview"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SalesTrends returns sales data over a period
func (h *AnalyticsHandler) SalesTrends(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	from := c.Query("from")
	to := c.Query("to")

	result, err := h.service(c).SalesTrends(tenantID, branchID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate sales trends"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current_revenue":  result.CurrentRevenue,
		"current_orders":   result.CurrentOrders,
		"previous_revenue": result.PreviousRevenue,
		"previous_orders":  result.PreviousOrders,
		"revenue_growth":   result.RevenueGrowth,
		"order_growth":     result.OrderGrowth,
		"daily_data":       result.DailyData,
		"period":           result.Period,
	})
}

// RevenueBreakdown returns revenue by category and payment method
func (h *AnalyticsHandler) RevenueBreakdown(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	from := c.Query("from")
	to := c.Query("to")

	result, err := h.service(c).RevenueBreakdown(tenantID, branchID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate revenue breakdown"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"by_category":       result.ByCategory,
		"by_payment_method": result.ByPaymentMethod,
	})
}

// TopProducts returns the best-selling products
func (h *AnalyticsHandler) TopProducts(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	from := c.Query("from")
	to := c.Query("to")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.service(c).TopProducts(tenantID, branchID, from, to, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate top products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"by_quantity": result.ByQuantity,
		"by_revenue":  result.ByRevenue,
	})
}

// CustomerMetrics returns customer analytics
func (h *AnalyticsHandler) CustomerMetrics(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	result, err := h.service(c).CustomerMetrics(tenantID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate customer metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_customers":  result.TotalCustomers,
		"active_customers": result.ActiveCustomers,
		"avg_order_value":  result.AvgOrderValue,
	})
}

// RealTimeMetrics returns metrics for the current day
func (h *AnalyticsHandler) RealTimeMetrics(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	result, err := h.service(c).RealTimeMetrics(tenantID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate real-time metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"today_revenue":      result.TodayRevenue,
		"today_orders":       result.TodayOrders,
		"inventory_value":    result.InventoryValue,
		"low_stock_count":    result.LowStockCount,
		"out_of_stock_count": result.OutOfStockCount,
		"total_products":     result.TotalProducts,
	})
}

func (h *AnalyticsHandler) SalesHeatmap(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	result, err := h.service(c).SalesHeatmap(tenantID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load heatmap"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) ReportBuilder(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	type ReportRequest struct {
		Metrics    []string `json:"metrics"`
		Dimensions []string `json:"dimensions"`
		From       string   `json:"from"`
		To         string   `json:"to"`
	}
	var req ReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	result, err := h.service(c).GenerateCustomReport(tenantID, branchID, req.Metrics, req.Dimensions, req.From, req.To)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) StaffPerformance(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	from := c.Query("from")
	to := c.Query("to")

	result, err := h.service(c).StaffPerformanceReport(tenantID, branchID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate staff performance"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) SalesGoalProgress(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	result, err := h.service(c).SalesGoalProgress(tenantID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sales goal progress"})
		return
	}

	c.JSON(http.StatusOK, result)
}
