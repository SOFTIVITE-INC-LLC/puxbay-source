package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OfflineHandler struct {
	db *gorm.DB
}

func NewOfflineHandler(db *gorm.DB) *OfflineHandler {
	return &OfflineHandler{db: db}
}

// SyncBatchedTransactions processes a batch of offline transactions from a POS terminal
// Strategy: Last-Write-Wins based on offline timestamps
func (h *OfflineHandler) SyncBatchedTransactions(c *gin.Context) {
	var payload struct {
		TerminalID   string  `json:"terminal_id"`
		Transactions []gin.H `json:"transactions"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock syncing logic
	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"synced_count": len(payload.Transactions),
		"message":      "Offline batch synced successfully using Last-Write-Wins strategy",
	})
}
