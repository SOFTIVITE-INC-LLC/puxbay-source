package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type NotificationService struct {
	db          *gorm.DB
	pushService *PushService
}

func NewNotificationService(db *gorm.DB, push *PushService) *NotificationService {
	return &NotificationService{db: db, pushService: push}
}

type NotificationListResult struct {
	Notifications []models.Notification
	Total         int64
	UnreadCount   int64
	Page          int
	PageSize      int
}

func (s *NotificationService) List(userID uuid.UUID, page, pageSize int) (*NotificationListResult, error) {
	offset := (page - 1) * pageSize

	var notifications []models.Notification
	var total int64

	if err := s.db.Where("user_id = ?", userID).Model(&models.Notification{}).Count(&total).Error; err != nil {
		return nil, err
	}
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").
		Limit(pageSize).Offset(offset).Find(&notifications).Error; err != nil {
		return nil, err
	}

	unreadCount := int64(0)
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).Count(&unreadCount).Error; err != nil {
		return nil, err
	}

	return &NotificationListResult{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

type LatestNotificationsResult struct {
	Count         int64
	Notifications []models.Notification
}

func (s *NotificationService) GetLatest(userID uuid.UUID) (*LatestNotificationsResult, error) {
	var notifications []models.Notification
	if err := s.db.Where("user_id = ? AND is_read = ?", userID, false).
		Order("created_at DESC").Limit(5).Find(&notifications).Error; err != nil {
		return nil, err
	}

	var unreadCount int64
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).Count(&unreadCount).Error; err != nil {
		return nil, err
	}

	return &LatestNotificationsResult{
		Count:         unreadCount,
		Notifications: notifications,
	}, nil
}

func (s *NotificationService) MarkAsRead(userID uuid.UUID, id string) error {
	notifID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid notification ID")
	}

	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Update("is_read", true)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("notification not found")
	}

	return nil
}

func (s *NotificationService) MarkAllAsRead(userID uuid.UUID) error {
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		return err
	}
	return nil
}

func (s *NotificationService) GetSettings(userID uuid.UUID) (*models.NotificationSetting, error) {
	var settings models.NotificationSetting
	if err := s.db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		settings = models.NotificationSetting{
			UserID:             userID,
			EmailNotifications: true,
			LowStockAlerts:     true,
			SalesReports:       true,
			SecurityAlerts:     true,
			SystemAlerts:       true,
		}
		if err := s.db.Create(&settings).Error; err != nil {
			return nil, err
		}
	}
	return &settings, nil
}

func (s *NotificationService) UpdateSettings(userID uuid.UUID, req *models.NotificationSetting) (*models.NotificationSetting, error) {
	var settings models.NotificationSetting
	if err := s.db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		settings = models.NotificationSetting{UserID: userID}
		if err := s.db.Create(&settings).Error; err != nil {
			return nil, err
		}
	}

	settings.EmailNotifications = req.EmailNotifications
	settings.LowStockAlerts = req.LowStockAlerts
	settings.SalesReports = req.SalesReports
	settings.SecurityAlerts = req.SecurityAlerts
	settings.SystemAlerts = req.SystemAlerts

	if err := s.db.Save(&settings).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *NotificationService) DeleteNotification(userID uuid.UUID, id string) error {
	notifID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid notification ID")
	}

	result := s.db.Where("id = ? AND user_id = ?", notifID, userID).
		Delete(&models.Notification{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("notification not found")
	}
	return nil
}

// CreateAndPush creates a persistent DB notification and simultaneously broadcasts
// a real-time push via the WebSocket hub to all connected clients for this user/tenant.
func (s *NotificationService) CreateAndPush(tenantID, userID uuid.UUID, title, message, category, link, notifType string) error {
	notif := models.Notification{
		UserID:           userID,
		Title:            title,
		Message:          message,
		Link:             link,
		IsRead:           false,
		NotificationType: notifType,
		Category:         category,
	}

	if err := s.db.Create(&notif).Error; err != nil {
		return err
	}

	// Fire WebSocket push (non-blocking — runs in goroutine)
	if s.pushService != nil {
		go s.pushService.SendToUser(tenantID, userID, title, message, category, link)
	}

	return nil
}

// CreateAndPushToAdmins creates a notification for all admin/manager users in a tenant
// and broadcasts a real-time push to all connected WebSocket clients.
func (s *NotificationService) CreateAndPushToAdmins(tenantID uuid.UUID, title, message, category, link, notifType string) {
	// Find all admin and manager users for this tenant
	var profiles []struct {
		UserID uuid.UUID
	}
	s.db.Table("public.user_profiles").
		Joins("JOIN public.roles ON public.roles.id = public.user_profiles.role_id").
		Where("public.user_profiles.tenant_id = ? AND public.roles.name IN ('admin', 'manager', 'branch_manager')", tenantID).
		Select("public.user_profiles.user_id").
		Scan(&profiles)

	for _, p := range profiles {
		_ = s.CreateAndPush(tenantID, p.UserID, title, message, category, link, notifType)
	}

	// Also broadcast over WebSocket to all connected clients in this tenant
	if s.pushService != nil {
		go s.pushService.SendToTenantAdmins(tenantID, title, message, category, link)
	}
}
