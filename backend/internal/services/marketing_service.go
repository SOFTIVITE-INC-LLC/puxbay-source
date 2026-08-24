package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type MarketingService struct {
	db *gorm.DB
}

func NewMarketingService(db *gorm.DB) *MarketingService {
	return &MarketingService{db: db}
}

// ─── Campaigns ───────────────────────────────────────────────────────────────

func (s *MarketingService) ListCampaigns() ([]models.MarketingCampaign, error) {
	var campaigns []models.MarketingCampaign
	if err := s.db.Order("created_at desc").Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

type CampaignInput struct {
	Name         string
	Type         string
	Status       string
	StartDate    string
	EndDate      string
	Budget       float64
	Subject      string
	Message      string
	IsAutomated  bool
	TriggerEvent string
	SegmentID    string
}

func parseDate(val string) (*time.Time, error) {
	if val == "" {
		return nil, nil
	}
	// Try RFC3339 first, then date-only
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", val); err == nil {
		return &t, nil
	}
	return nil, errors.New("invalid date format")
}

func (s *MarketingService) CreateCampaign(input CampaignInput) (*models.MarketingCampaign, error) {
	scheduledAt, err := parseDate(input.StartDate)
	if err != nil {
		return nil, errors.New("invalid start date format")
	}

	campaign := models.MarketingCampaign{
		Name:         input.Name,
		CampaignType: input.Type,
		Status:       input.Status,
		ScheduledAt:  scheduledAt,
		IsAutomated:  input.IsAutomated,
		TriggerEvent: input.TriggerEvent,
	}

	if input.Subject != "" {
		campaign.Subject = &input.Subject
	}
	campaign.Message = input.Message

	if input.SegmentID != "" {
		sid, err := uuid.Parse(input.SegmentID)
		if err == nil {
			campaign.SegmentID = &sid
		}
	}

	if err := s.db.Create(&campaign).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (s *MarketingService) GetCampaign(id string) (*models.MarketingCampaign, error) {
	var campaign models.MarketingCampaign
	if err := s.db.First(&campaign, "id = ?", id).Error; err != nil {
		return nil, errors.New("campaign not found")
	}
	return &campaign, nil
}

func (s *MarketingService) UpdateCampaign(id string, input CampaignInput) (*models.MarketingCampaign, error) {
	var campaign models.MarketingCampaign
	if err := s.db.First(&campaign, "id = ?", id).Error; err != nil {
		return nil, errors.New("campaign not found")
	}

	scheduledAt, err := parseDate(input.StartDate)
	if err != nil {
		return nil, errors.New("invalid start date format")
	}

	campaign.Name = input.Name
	campaign.CampaignType = input.Type
	campaign.Status = input.Status
	campaign.ScheduledAt = scheduledAt
	campaign.IsAutomated = input.IsAutomated
	campaign.TriggerEvent = input.TriggerEvent
	if input.Subject != "" {
		campaign.Subject = &input.Subject
	}
	if input.Message != "" {
		campaign.Message = input.Message
	}
	if input.SegmentID != "" {
		sid, err := uuid.Parse(input.SegmentID)
		if err == nil {
			campaign.SegmentID = &sid
		}
	} else {
		campaign.SegmentID = nil
	}

	if err := s.db.Save(&campaign).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (s *MarketingService) DeleteCampaign(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.MarketingCampaign{}).Error
}

func (s *MarketingService) SendCampaign(id string) error {
	var campaign models.MarketingCampaign
	if err := s.db.First(&campaign, "id = ?", id).Error; err != nil {
		return errors.New("campaign not found")
	}

	if campaign.Status == "sent" {
		return errors.New("campaign already sent")
	}

	campaign.Status = "sent"
	now := time.Now()
	campaign.SentAt = &now

	// Increment open count for demo — in production, a real email provider webhook triggers this.
	campaign.OpenCount = 0

	return s.db.Save(&campaign).Error
}

// RecordCampaignOpen increments the open analytics counter.
func (s *MarketingService) RecordCampaignOpen(id string) error {
	return s.db.Model(&models.MarketingCampaign{}).Where("id = ?", id).
		UpdateColumn("open_count", gorm.Expr("open_count + ?", 1)).Error
}

// RecordCampaignConversion increments conversions and adds revenue attribution.
func (s *MarketingService) RecordCampaignConversion(id string, revenue float64) error {
	return s.db.Model(&models.MarketingCampaign{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"conversion_count":  gorm.Expr("conversion_count + ?", 1),
			"revenue_generated": gorm.Expr("revenue_generated + ?", revenue),
		}).Error
}

// TriggerEventCampaigns finds all automated campaigns with a matching trigger event
// and marks their last_run_at, simulating a dispatch.
func (s *MarketingService) TriggerEventCampaigns(eventType string) ([]models.MarketingCampaign, error) {
	var campaigns []models.MarketingCampaign
	if err := s.db.Where("is_automated = true AND trigger_event = ? AND status != ?", eventType, "sent").
		Find(&campaigns).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	for i := range campaigns {
		campaigns[i].LastRunAt = &now
		s.db.Save(&campaigns[i])
	}
	return campaigns, nil
}

// ─── Customer Segments ────────────────────────────────────────────────────────

func (s *MarketingService) ListSegments() ([]models.CustomerSegment, error) {
	var segments []models.CustomerSegment
	if err := s.db.Order("name asc").Find(&segments).Error; err != nil {
		return nil, err
	}
	return segments, nil
}

type SegmentInput struct {
	Name         string
	Description  string
	CriteriaJSON string // JSON string like: {"min_spend": 500, "days_inactive": 30}
}

func (s *MarketingService) CreateSegment(input SegmentInput) (*models.CustomerSegment, error) {
	segment := models.CustomerSegment{
		Name:         input.Name,
		Description:  input.Description,
		CriteriaJSON: input.CriteriaJSON,
	}
	if err := s.db.Create(&segment).Error; err != nil {
		return nil, err
	}
	return &segment, nil
}

func (s *MarketingService) UpdateSegment(id string, input SegmentInput) (*models.CustomerSegment, error) {
	var segment models.CustomerSegment
	if err := s.db.First(&segment, "id = ?", id).Error; err != nil {
		return nil, errors.New("segment not found")
	}
	segment.Name = input.Name
	segment.Description = input.Description
	segment.CriteriaJSON = input.CriteriaJSON
	if err := s.db.Save(&segment).Error; err != nil {
		return nil, err
	}
	return &segment, nil
}

func (s *MarketingService) DeleteSegment(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.CustomerSegment{}).Error
}

// ─── Promotions ───────────────────────────────────────────────────────────────

func (s *MarketingService) ListPromotions() ([]models.Promotion, error) {
	var promos []models.Promotion
	if err := s.db.Find(&promos).Error; err != nil {
		return nil, err
	}
	return promos, nil
}

func (s *MarketingService) CreatePromotion(input models.Promotion) (*models.Promotion, error) {
	if err := s.db.Create(&input).Error; err != nil {
		return nil, err
	}
	return &input, nil
}

func (s *MarketingService) UpdatePromotion(id string, input models.Promotion) (*models.Promotion, error) {
	var promo models.Promotion
	if err := s.db.First(&promo, "id = ?", id).Error; err != nil {
		return nil, errors.New("promotion not found")
	}
	input.ID = promo.ID
	if err := s.db.Save(&input).Error; err != nil {
		return nil, err
	}
	return &input, nil
}

func (s *MarketingService) DeletePromotion(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.Promotion{}).Error
}

// ─── Discount Codes ───────────────────────────────────────────────────────────

func (s *MarketingService) ListDiscounts() ([]models.DiscountCode, error) {
	var discounts []models.DiscountCode
	if err := s.db.Find(&discounts).Error; err != nil {
		return nil, err
	}
	return discounts, nil
}

func (s *MarketingService) CreateDiscount(input models.DiscountCode) (*models.DiscountCode, error) {
	if err := s.db.Create(&input).Error; err != nil {
		return nil, err
	}
	return &input, nil
}

func (s *MarketingService) DeleteDiscount(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.DiscountCode{}).Error
}

// RedeemPointsForDiscount creates a unique discount code in exchange for loyalty points.
func (s *MarketingService) RedeemPointsForDiscount(customerID string, points int, discountValue float64) (*models.DiscountCode, error) {
	code := models.DiscountCode{
		Code:           fmt.Sprintf("REDEEM-%s-%d", customerID[:6], time.Now().Unix()),
		Type:           "fixed",
		Value:          discountValue,
		Status:         "active",
		MaxUses:        intPtr(1),
		CurrentUses:    0,
		ValidFrom:      time.Now(),
		ValidUntil:     timePtr(time.Now().Add(30 * 24 * time.Hour)),
		PointsRequired: &points,
	}
	if err := s.db.Create(&code).Error; err != nil {
		return nil, err
	}
	return &code, nil
}

func intPtr(v int) *int              { return &v }
func timePtr(t time.Time) *time.Time { return &t }
