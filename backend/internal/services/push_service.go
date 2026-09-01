package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	ws "github.com/softivite/puxbay/internal/websocket"
	"gorm.io/gorm"
)

// PushService delivers real-time push notifications using the existing WebSocket hub
// for web clients, and stores tokens for future native mobile push delivery.
type PushService struct {
	db  *gorm.DB
	hub *ws.Hub
}

func NewPushService(db *gorm.DB, hub *ws.Hub) *PushService {
	return &PushService{db: db, hub: hub}
}

// PushPayload is the JSON structure sent over WebSocket.
type PushPayload struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Category  string `json:"category"`
	NotifType string `json:"notif_type,omitempty"`
	SoundType string `json:"sound_type,omitempty"`
	Link      string `json:"link,omitempty"`
	Icon      string `json:"icon,omitempty"`
}

// RegisterToken stores a device token for a user.
func (s *PushService) RegisterToken(userID, tenantID uuid.UUID, token, platform string) error {
	dt := models.DeviceToken{
		UserID:   userID,
		TenantID: tenantID,
		Token:    token,
		Platform: platform,
		IsActive: true,
	}
	return s.db.Where(models.DeviceToken{Token: token}).
		Assign(models.DeviceToken{UserID: userID, TenantID: tenantID, Platform: platform, IsActive: true}).
		FirstOrCreate(&dt).Error
}

// DeactivateToken marks a device token as inactive (e.g. on logout).
func (s *PushService) DeactivateToken(token string) error {
	return s.db.Model(&models.DeviceToken{}).
		Where("token = ?", token).
		Update("is_active", false).Error
}

func soundForCategory(category, notifType string) string {
	switch {
	case notifType == "online_order" || category == "online_order" || category == "storefront":
		return "online_order"
	case notifType == "kiosk_order" || category == "kiosk" || category == "kiosk_order":
		return "kiosk_order"
	case notifType == "pos_completed" || category == "pos_order_completed" || category == "sale":
		return "pos_completed"
	case notifType == "low_stock" || category == "low_stock" || category == "inventory":
		return "low_stock"
	case notifType == "anomaly" || category == "anomaly" || category == "security":
		return "anomaly"
	default:
		return "general"
	}
}

// SendToUser broadcasts a push notification to all active WebSocket connections
// for the given user (via their tenant channel), and stores the token list
// for future native delivery.
func (s *PushService) SendToUser(tenantID uuid.UUID, userID uuid.UUID, title, message, category, link string) {
	s.SendToUserWithSound(tenantID, userID, title, message, category, link, "", soundForCategory(category, ""))
}

func (s *PushService) SendToUserWithSound(tenantID uuid.UUID, userID uuid.UUID, title, message, category, link, notifType, soundType string) {
	if soundType == "" {
		soundType = soundForCategory(category, notifType)
	}

	payload := PushPayload{
		Type:      "notification",
		Title:     title,
		Message:   message,
		Category:  category,
		NotifType: notifType,
		SoundType: soundType,
		Link:      link,
		Icon:      iconForCategory(category),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[PushService] Failed to marshal push payload: %v", err)
		return
	}

	if s.hub != nil {
		s.hub.BroadcastMessage(tenantID.String(), data)
	}
}

// SendToTenantAdmins broadcasts a notification to all admin WebSocket clients.
func (s *PushService) SendToTenantAdmins(tenantID uuid.UUID, title, message, category, link string) {
	s.SendToTenantAdminsWithSound(tenantID, title, message, category, link, "", soundForCategory(category, ""))
}

func (s *PushService) SendToTenantAdminsWithSound(tenantID uuid.UUID, title, message, category, link, notifType, soundType string) {
	if soundType == "" {
		soundType = soundForCategory(category, notifType)
	}

	payload := PushPayload{
		Type:      "notification",
		Title:     title,
		Message:   message,
		Category:  category,
		NotifType: notifType,
		SoundType: soundType,
		Link:      link,
		Icon:      iconForCategory(category),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[PushService] Failed to marshal admin push payload: %v", err)
		return
	}

	if s.hub != nil {
		s.hub.BroadcastMessage(tenantID.String(), data)
	}
}

// GetActiveTokens returns stored device tokens for a user (for future native push).
func (s *PushService) GetActiveTokens(userID uuid.UUID) ([]models.DeviceToken, error) {
	var tokens []models.DeviceToken
	if err := s.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func iconForCategory(category string) string {
	icons := map[string]string{
		"inventory": "inventory_2",
		"sales":     "point_of_sale",
		"security":  "security",
		"system":    "settings",
		"order":     "receipt_long",
		"loyalty":   "stars",
		"anomaly":   "warning",
	}
	if icon, ok := icons[category]; ok {
		return icon
	}
	return fmt.Sprintf("%s", "notifications")
}
