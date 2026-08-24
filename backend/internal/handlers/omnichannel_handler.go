package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OmnichannelHandler struct {
	db *gorm.DB
}

func NewOmnichannelHandler(db *gorm.DB) *OmnichannelHandler {
	return &OmnichannelHandler{db: db}
}

// WhatsAppWebhook processes incoming messages from the WhatsApp Business API.
func (h *OmnichannelHandler) WhatsAppWebhook(c *gin.Context) {
	// 1. Verify webhook signature (omitted for brevity)

	// 2. Parse incoming message
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	log.Printf("[WHATSAPP] Received message: %+v\n", payload)

	// 3. (Mock) Pass message to AI Copilot or directly process order intent

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// SyncTikTokCatalog pushes the latest active products to TikTok Shop API.
func (h *OmnichannelHandler) SyncTikTokCatalog(c *gin.Context) {
	// (Mock) In a real scenario, this fetches models.Product where is_active=true
	// and sends a POST request to TikTok Shop API.

	log.Println("[TIKTOK] Syncing catalog...")

	c.JSON(http.StatusOK, gin.H{"message": "Catalog sync initiated"})
}
