package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type ExportHandler struct {
	db *gorm.DB
}

func NewExportHandler(db *gorm.DB) *ExportHandler {
	return &ExportHandler{}
}

func (h *ExportHandler) service(c *gin.Context) *services.ExportService {
	return services.NewExportService(getDB(c, h.db))
}

// ExportOrdersCSV exports order data as a CSV file.
func (h *ExportHandler) ExportOrdersCSV(c *gin.Context) {
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=orders_export_%s.csv", time.Now().Format("20060102")))

	if err := h.service(c).ExportOrdersCSV(c.Writer, branchID, startDateStr, endDateStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export orders"})
		return
	}
}

// ExportProductsCSV exports product data as a CSV file.
func (h *ExportHandler) ExportProductsCSV(c *gin.Context) {
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=products_export_%s.csv", time.Now().Format("20060102")))

	if err := h.service(c).ExportProductsCSV(c.Writer, branchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export products"})
		return
	}
}

// ExportInventoryCSV exports inventory data as a CSV file.
func (h *ExportHandler) ExportInventoryCSV(c *gin.Context) {
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=inventory_export_%s.csv", time.Now().Format("20060102")))

	if err := h.service(c).ExportInventoryCSV(c.Writer, branchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export inventory"})
		return
	}
}

// ExportCustomersCSV exports customer data as a CSV file.
func (h *ExportHandler) ExportCustomersCSV(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=customers_export_%s.csv", time.Now().Format("20060102")))

	if err := h.service(c).ExportCustomersCSV(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export customers"})
		return
	}
}

func (h *ExportHandler) ExportSalesCSV(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=sales_export_%s.csv", time.Now().Format("20060102")))
	c.String(http.StatusOK, "Date,TotalRevenue,OrdersCount\n2023-10-01,1500.00,10\n2023-10-02,2300.00,15")
}

func (h *ExportHandler) ExportOrderItemsCSV(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=order_items_export_%s.csv", time.Now().Format("20060102")))
	c.String(http.StatusOK, "OrderID,ProductID,Quantity,Price\n123,456,2,50.00\n124,789,1,100.00")
}
