package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/services"
)

type DeviceTokenHandler struct {
	pushSvc *services.PushService
}

func NewDeviceTokenHandler(pushSvc *services.PushService) *DeviceTokenHandler {
	return &DeviceTokenHandler{pushSvc: pushSvc}
}

type RegisterTokenRequest struct {
	Token    string `json:"token"    binding:"required"`
	Platform string `json:"platform"` // web, android, ios
}

// RegisterDeviceToken registers a device push token for the authenticated user.
// POST /api/devices/register
func (h *DeviceTokenHandler) RegisterDeviceToken(c *gin.Context) {
	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")

	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	tid, err := uuid.Parse(tenantID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant"})
		return
	}

	var req RegisterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	platform := req.Platform
	if platform == "" {
		platform = "web"
	}

	if err := h.pushSvc.RegisterToken(uid, tid, req.Token, platform); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "device token registered"})
}

// UnregisterDeviceToken deactivates a device push token.
// DELETE /api/devices/unregister
func (h *DeviceTokenHandler) UnregisterDeviceToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.pushSvc.DeactivateToken(req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unregister token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "device token unregistered"})
}
