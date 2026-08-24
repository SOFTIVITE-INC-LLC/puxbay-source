package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type WebhookHandler struct {
	db *gorm.DB
}

func NewWebhookHandler(db *gorm.DB) *WebhookHandler {
	return &WebhookHandler{}
}

func (h *WebhookHandler) getDB(c *gin.Context) *gorm.DB {
	return getDB(c, h.db)
}

func (h *WebhookHandler) List(c *gin.Context) {
	db := h.getDB(c)
	var endpoints []models.WebhookEndpoint
	if err := db.Find(&endpoints).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list webhooks"})
		return
	}
	c.JSON(http.StatusOK, endpoints)
}

func (h *WebhookHandler) Create(c *gin.Context) {
	db := h.getDB(c)
	var endpoint models.WebhookEndpoint
	if err := c.ShouldBindJSON(&endpoint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if endpoint.Secret == "" {
		endpoint.Secret = uuid.NewString() // Generate a secret if not provided
	}

	if err := db.Create(&endpoint).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook"})
		return
	}
	c.JSON(http.StatusCreated, endpoint)
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	db := h.getDB(c)
	id := c.Param("id")

	if err := db.Delete(&models.WebhookEndpoint{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted"})
}
