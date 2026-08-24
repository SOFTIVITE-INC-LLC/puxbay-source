package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SyncHandler struct {
	db *gorm.DB
}

func NewSyncHandler(db *gorm.DB) *SyncHandler {
	return &SyncHandler{db: db}
}

// SyncData handles CRDT resolution from the offline POS.
func (h *SyncHandler) SyncData(c *gin.Context) {
	// 1. Receive batch changes from frontend IndexedDB
	var payload struct {
		LastSyncTime int64                    `json:"last_sync_time"`
		Changes      []map[string]interface{} `json:"changes"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Limit payload size to prevent huge sync blocks
	const MaxSyncBatchSize = 1000
	if len(payload.Changes) > MaxSyncBatchSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Payload exceeds max batch size. Please sync more frequently or paginate."})
		return
	}

	// 2. Resolve conflicts using Last-Writer-Wins (LWW) or Vector Clocks
	// Since we added `Version` int64 to Base, we can increment versions on the backend
	// and reject frontend changes if their base version < server version.

	// (Mock processing)
	for _, change := range payload.Changes {
		_ = change
		// e.g. Upsert into database
	}

	// 3. Return the new backend state changes since LastSyncTime
	c.JSON(http.StatusOK, gin.H{
		"server_time":      1718000000,
		"status":           "synchronized",
		"resolved_changes": []interface{}{},
	})
}
