package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"github.com/softivite/puxbay/internal/websocket"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	db  *gorm.DB
	hub *websocket.Hub
}

func NewNotificationHandler(db *gorm.DB, hub *websocket.Hub) *NotificationHandler {
	return &NotificationHandler{db: db, hub: hub}
}

func (h *NotificationHandler) service(c *gin.Context) *services.NotificationService {
	tenantDB := getDB(c, h.db)
	pushSvc := services.NewPushService(tenantDB, h.hub)
	return services.NewNotificationService(tenantDB, pushSvc)
}

func (h *NotificationHandler) getUserID(c *gin.Context) uuid.UUID {
	rawUserID, _ := c.Get(middleware.ContextKeyUserID)
	userID, _ := rawUserID.(uuid.UUID)
	return userID
}

// List returns all notifications for the authenticated user (paginated).
func (h *NotificationHandler) List(c *gin.Context) {
	userID := h.getUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.service(c).List(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": result.Notifications,
		"total":         result.Total,
		"unread_count":  result.UnreadCount,
		"page":          result.Page,
		"page_size":     result.PageSize,
	})
}

// GetLatest returns the latest 5 unread notifications (for topbar badge).
func (h *NotificationHandler) GetLatest(c *gin.Context) {
	userID := h.getUserID(c)

	result, err := h.service(c).GetLatest(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch latest notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":         result.Count,
		"notifications": result.Notifications,
	})
}

// MarkAsRead marks a single notification as read.
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := h.getUserID(c)
	id := c.Param("id")

	if err := h.service(c).MarkAsRead(userID, id); err != nil {
		if err.Error() == "invalid notification ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "notification not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// MarkAllAsRead marks all notifications for the user as read.
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := h.getUserID(c)

	if err := h.service(c).MarkAllAsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetSettings returns the notification preferences for the user.
func (h *NotificationHandler) GetSettings(c *gin.Context) {
	userID := h.getUserID(c)

	settings, err := h.service(c).GetSettings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSettings updates notification preferences.
func (h *NotificationHandler) UpdateSettings(c *gin.Context) {
	userID := h.getUserID(c)

	var req models.NotificationSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service(c).UpdateSettings(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// Delete deletes a notification.
func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := h.getUserID(c)
	id := c.Param("id")

	if err := h.service(c).DeleteNotification(userID, id); err != nil {
		if err.Error() == "invalid notification ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "notification not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
