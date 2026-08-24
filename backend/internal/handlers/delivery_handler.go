package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type DeliveryHandler struct {
	db *gorm.DB
}

func NewDeliveryHandler(db *gorm.DB) *DeliveryHandler {
	return &DeliveryHandler{db: db}
}

// ListDrivers returns all delivery drivers for the tenant.
func (h *DeliveryHandler) ListDrivers(c *gin.Context) {
	var drivers []models.DeliveryDriver
	if err := getDB(c, h.db).Find(&drivers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch drivers"})
		return
	}
	c.JSON(http.StatusOK, drivers)
}

// AddDriver adds a new delivery driver.
func (h *DeliveryHandler) AddDriver(c *gin.Context) {

	var req struct {
		Name        string `json:"name" binding:"required"`
		Phone       string `json:"phone" binding:"required"`
		VehicleInfo string `json:"vehicle_info"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	driver := models.DeliveryDriver{
		TenantScoped: models.TenantScoped{Base: models.Base{ID: uuid.New()}},
		Name:         req.Name,
		Phone:        req.Phone,
		VehicleInfo:  req.VehicleInfo,
	}

	if err := getDB(c, h.db).Create(&driver).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add driver"})
		return
	}

	c.JSON(http.StatusCreated, driver)
}

// DispatchOrder assigns a driver to an order and generates a tracking link.
func (h *DeliveryHandler) DispatchOrder(c *gin.Context) {

	var req struct {
		OrderID  uuid.UUID `json:"order_id" binding:"required"`
		DriverID uuid.UUID `json:"driver_id" binding:"required"`
		Fee      float64   `json:"delivery_fee"`
		Notes    string    `json:"delivery_notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	db := getDB(c, h.db)

	// Generate a secure tracking token
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	trackingToken := hex.EncodeToString(b)
	trackingLink := "https://puxbay.com/track/" + trackingToken

	delivery := models.DeliveryOrder{
		TenantScoped:  models.TenantScoped{Base: models.Base{ID: uuid.New()}},
		OrderID:       req.OrderID,
		DriverID:      &req.DriverID,
		Status:        "assigned",
		TrackingLink:  trackingLink,
		DeliveryFee:   req.Fee,
		DeliveryNotes: req.Notes,
	}

	if err := db.Create(&delivery).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dispatch order"})
		return
	}

	// Also update driver status
	db.Model(&models.DeliveryDriver{}).Where("id = ?", req.DriverID).Update("current_status", "busy")

	c.JSON(http.StatusCreated, delivery)
}

// ListDispatchedOrders returns all active delivery orders.
func (h *DeliveryHandler) ListDispatchedOrders(c *gin.Context) {
	var deliveries []models.DeliveryOrder
	if err := getDB(c, h.db).Preload("Driver").Preload("Order.Customer").Find(&deliveries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dispatched orders"})
		return
	}
	c.JSON(http.StatusOK, deliveries)
}

// CompleteDelivery marks a delivery as completed.
func (h *DeliveryHandler) CompleteDelivery(c *gin.Context) {
	deliveryID := c.Param("id")

	var delivery models.DeliveryOrder
	db := getDB(c, h.db)
	if err := db.Where("id = ?", deliveryID).First(&delivery).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Delivery not found"})
		return
	}

	delivery.Status = "completed"
	if err := db.Save(&delivery).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update delivery"})
		return
	}

	// Update driver status back to available
	if delivery.DriverID != nil {
		db.Model(&models.DeliveryDriver{}).Where("id = ?", *delivery.DriverID).Update("current_status", "available")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery completed"})
}


// TrackOrder is a public endpoint that resolves a tracking link to delivery status.
func (h *DeliveryHandler) TrackOrder(c *gin.Context) {
	token := c.Param("token")
	trackingLink := "https://puxbay.com/track/" + token

	var delivery models.DeliveryOrder
	if err := h.db.Preload("Driver").Preload("Order").Where("tracking_link = ?", trackingLink).First(&delivery).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tracking link not found or expired"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   delivery.Status,
		"driver":   delivery.Driver.Name,
		"lat":      delivery.Driver.Lat,
		"lng":      delivery.Driver.Lng,
		"order_id": delivery.OrderID,
	})
}
