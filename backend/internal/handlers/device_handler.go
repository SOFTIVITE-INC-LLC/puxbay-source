package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/workers"
	"gorm.io/gorm"
)

type DeviceHandler struct {
	db          *gorm.DB
	asynqClient *asynq.Client
}

func NewDeviceHandler(db *gorm.DB, cfg *config.Config) *DeviceHandler {
	var client *asynq.Client
	if cfg != nil {
		redisOpt, err := asynq.ParseRedisURI(cfg.Redis.URL)
		if err == nil {
			client = asynq.NewClient(redisOpt)
		} else {
			log.Printf("Failed to parse Redis URI for device handler: %v", err)
		}
	}

	return &DeviceHandler{db: db, asynqClient: client}
}

func (h *DeviceHandler) List(c *gin.Context) {
	db := getDB(c, h.db)
	var devices []models.Device
	
	query := db.Model(&models.Device{})
	
	if branchID := c.Query("branch_id"); branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	
	if err := query.Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch devices"})
		return
	}

	// Compute live online/offline based on last_seen_at
	now := time.Now()
	for i := range devices {
		if devices[i].LastSeenAt != nil && now.Sub(*devices[i].LastSeenAt) < 5*time.Minute {
			devices[i].Status = "online"
		} else if devices[i].LastSeenAt != nil {
			devices[i].Status = "offline"
		}
	}

	c.JSON(http.StatusOK, devices)
}

type DeviceCreateRequest struct {
	BranchID   *string `json:"branch_id"`
	Name       string  `json:"name" binding:"required"`
	DeviceType string  `json:"device_type" binding:"required"`
	IPAddress  string  `json:"ip_address"`
	MACAddress string  `json:"mac_address"`
	Config     string  `json:"config"` // JSON string
}

func (h *DeviceHandler) Create(c *gin.Context) {
	var req DeviceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	db := getDB(c, h.db)

	device := models.Device{
		Name:       req.Name,
		DeviceType: req.DeviceType,
		IPAddress:  req.IPAddress,
		MACAddress: req.MACAddress,
		Status:     "online",
	}

	if req.BranchID != nil && *req.BranchID != "" {
		parsed, err := uuid.Parse(*req.BranchID)
		if err == nil {
			device.BranchID = &parsed
		}
	}
	
	if req.Config != "" {
		device.Config = []byte(req.Config)
	} else {
		device.Config = []byte("{}")
	}

	if err := db.Create(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create device"})
		return
	}

	c.JSON(http.StatusCreated, device)
}

func (h *DeviceHandler) Get(c *gin.Context) {
	db := getDB(c, h.db)
	id := c.Param("id")

	var device models.Device
	if err := db.First(&device, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (h *DeviceHandler) Update(c *gin.Context) {
	db := getDB(c, h.db)
	id := c.Param("id")

	var device models.Device
	if err := db.First(&device, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	var req DeviceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	device.Name = req.Name
	device.DeviceType = req.DeviceType
	device.IPAddress = req.IPAddress
	device.MACAddress = req.MACAddress

	if req.BranchID != nil && *req.BranchID != "" {
		parsed, err := uuid.Parse(*req.BranchID)
		if err == nil {
			device.BranchID = &parsed
		}
	}
	
	if req.Config != "" {
		device.Config = []byte(req.Config)
	}

	if err := db.Save(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device"})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (h *DeviceHandler) Delete(c *gin.Context) {
	db := getDB(c, h.db)
	id := c.Param("id")

	if err := db.Delete(&models.Device{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete device"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Heartbeat records the device's last-seen timestamp and marks it online.
func (h *DeviceHandler) Heartbeat(c *gin.Context) {
	db := getDB(c, h.db)
	id := c.Param("id")

	now := time.Now()
	result := db.Model(&models.Device{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_seen_at": now,
		"status":       "online",
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update heartbeat"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "online", "last_seen_at": now})
}

// PrintDocument enqueues a print job for the device.
func (h *DeviceHandler) PrintDocument(c *gin.Context) {
	db := getDB(c, h.db)
	deviceIDStr := c.Param("id")

	var device models.Device
	if err := db.First(&device, "id = ?", deviceIDStr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	if device.DeviceType != "printer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device is not a printer"})
		return
	}

	var req struct {
		DocumentType string `json:"document_type" binding:"required"`
		Content      string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	printJob := models.PrintJob{
		DocumentType: req.DocumentType,
		Content:      req.Content,
		Status:       "pending",
	}

	if device.BranchID != nil {
		printJob.BranchID = *device.BranchID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Printer must be associated with a branch"})
		return
	}

	if err := db.Create(&printJob).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create print job"})
		return
	}

	if h.asynqClient != nil {
		err := workers.EnqueuePrintJob(h.asynqClient, printJob.ID.String())
		if err != nil {
			log.Printf("Failed to enqueue print job: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue print job"})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Print queue is unavailable"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "queued", "job_id": printJob.ID})
}
