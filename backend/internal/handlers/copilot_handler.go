package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/services"
)

type CopilotHandler struct {
	db             *gorm.DB
	copilotService *services.CopilotService
}

func NewCopilotHandler(cfg *config.Config, db *gorm.DB) *CopilotHandler {
	return &CopilotHandler{
		db:             db,
		copilotService: services.NewCopilotService(cfg, db),
	}
}

type ChatRequest struct {
	History []services.ChatMessage `json:"history" binding:"required"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

func (h *CopilotHandler) Chat(c *gin.Context) {
	tenantIDRaw, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id not found in context"})
		return
	}
	tenantID, ok := tenantIDRaw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant_id has invalid type"})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	reply, err := h.copilotService.Chat(c.Request.Context(), tenantID, req.History)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI processing failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{Reply: reply})
}
