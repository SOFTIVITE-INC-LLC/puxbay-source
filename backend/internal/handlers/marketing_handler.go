package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type MarketingHandler struct {
	db *gorm.DB
}

func NewMarketingHandler(db *gorm.DB) *MarketingHandler {
	return &MarketingHandler{db: db}
}

func (h *MarketingHandler) service(c *gin.Context) *services.MarketingService {
	return services.NewMarketingService(getDB(c, h.db))
}

// ─── Campaigns ─────────────────────────────────────────────────────────────

func (h *MarketingHandler) ListCampaigns(c *gin.Context) {
	campaigns, err := h.service(c).ListCampaigns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch campaigns"})
		return
	}
	c.JSON(http.StatusOK, campaigns)
}

type CampaignCreateRequest struct {
	Name         string  `json:"name" binding:"required"`
	Type         string  `json:"type" binding:"required"` // email, sms, push
	Status       string  `json:"status" binding:"required"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	Budget       float64 `json:"budget"`
	Subject      string  `json:"subject"`
	Message      string  `json:"message"`
	IsAutomated  bool    `json:"is_automated"`
	TriggerEvent string  `json:"trigger_event"`
	SegmentID    string  `json:"segment_id"`
}

func (h *MarketingHandler) CreateCampaign(c *gin.Context) {
	var req CampaignCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.CampaignInput{
		Name:         req.Name,
		Type:         req.Type,
		Status:       req.Status,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Budget:       req.Budget,
		Subject:      req.Subject,
		Message:      req.Message,
		IsAutomated:  req.IsAutomated,
		TriggerEvent: req.TriggerEvent,
		SegmentID:    req.SegmentID,
	}

	campaign, err := h.service(c).CreateCampaign(input)
	if err != nil {
		if err.Error() == "invalid start date format" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create campaign"})
		return
	}
	c.JSON(http.StatusCreated, campaign)
}

func (h *MarketingHandler) GetCampaign(c *gin.Context) {
	id := c.Param("id")
	campaign, err := h.service(c).GetCampaign(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, campaign)
}

func (h *MarketingHandler) UpdateCampaign(c *gin.Context) {
	id := c.Param("id")
	var req CampaignCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.CampaignInput{
		Name:         req.Name,
		Type:         req.Type,
		Status:       req.Status,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Budget:       req.Budget,
		Subject:      req.Subject,
		Message:      req.Message,
		IsAutomated:  req.IsAutomated,
		TriggerEvent: req.TriggerEvent,
		SegmentID:    req.SegmentID,
	}

	campaign, err := h.service(c).UpdateCampaign(id, input)
	if err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "invalid start date format" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update campaign"})
		return
	}
	c.JSON(http.StatusOK, campaign)
}

func (h *MarketingHandler) DeleteCampaign(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).DeleteCampaign(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete campaign"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *MarketingHandler) SendCampaign(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).SendCampaign(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}

func (h *MarketingHandler) RecordCampaignOpen(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).RecordCampaignOpen(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record open"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

type ConversionRequest struct {
	Revenue float64 `json:"revenue"`
}

func (h *MarketingHandler) RecordCampaignConversion(c *gin.Context) {
	id := c.Param("id")
	var req ConversionRequest
	c.ShouldBindJSON(&req)
	if err := h.service(c).RecordCampaignConversion(id, req.Revenue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record conversion"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

type TriggerRequest struct {
	EventType string `json:"event_type" binding:"required"`
}

func (h *MarketingHandler) TriggerEventCampaigns(c *gin.Context) {
	var req TriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	campaigns, err := h.service(c).TriggerEventCampaigns(req.EventType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger campaigns"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"triggered": len(campaigns), "campaigns": campaigns})
}

// ─── Segments ──────────────────────────────────────────────────────────────

func (h *MarketingHandler) ListSegments(c *gin.Context) {
	segments, err := h.service(c).ListSegments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch segments"})
		return
	}
	c.JSON(http.StatusOK, segments)
}

type SegmentRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	CriteriaJSON string `json:"criteria_json"`
}

func (h *MarketingHandler) CreateSegment(c *gin.Context) {
	var req SegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input := services.SegmentInput{
		Name:         req.Name,
		Description:  req.Description,
		CriteriaJSON: req.CriteriaJSON,
	}
	segment, err := h.service(c).CreateSegment(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create segment"})
		return
	}
	c.JSON(http.StatusCreated, segment)
}

func (h *MarketingHandler) UpdateSegment(c *gin.Context) {
	id := c.Param("id")
	var req SegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	segment, err := h.service(c).UpdateSegment(id, services.SegmentInput{
		Name:         req.Name,
		Description:  req.Description,
		CriteriaJSON: req.CriteriaJSON,
	})
	if err != nil {
		if err.Error() == "segment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update segment"})
		return
	}
	c.JSON(http.StatusOK, segment)
}

func (h *MarketingHandler) DeleteSegment(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).DeleteSegment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete segment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ─── Promotions ───────────────────────────────────────────────────────────

func (h *MarketingHandler) ListPromotions(c *gin.Context) {
	promos, err := h.service(c).ListPromotions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch promotions"})
		return
	}
	c.JSON(http.StatusOK, promos)
}

func (h *MarketingHandler) CreatePromotion(c *gin.Context) {
	var req models.Promotion
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	promo, err := h.service(c).CreatePromotion(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create promotion"})
		return
	}
	c.JSON(http.StatusCreated, promo)
}

func (h *MarketingHandler) UpdatePromotion(c *gin.Context) {
	id := c.Param("id")
	var req models.Promotion
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	promo, err := h.service(c).UpdatePromotion(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update promotion"})
		return
	}
	c.JSON(http.StatusOK, promo)
}

func (h *MarketingHandler) DeletePromotion(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).DeletePromotion(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete promotion"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ─── Discount Codes ──────────────────────────────────────────────────────

func (h *MarketingHandler) ListDiscounts(c *gin.Context) {
	discounts, err := h.service(c).ListDiscounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch discounts"})
		return
	}
	c.JSON(http.StatusOK, discounts)
}

func (h *MarketingHandler) CreateDiscount(c *gin.Context) {
	var req models.DiscountCode
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	discount, err := h.service(c).CreateDiscount(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create discount"})
		return
	}
	c.JSON(http.StatusCreated, discount)
}

func (h *MarketingHandler) DeleteDiscount(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).DeleteDiscount(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete discount"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ─── Loyalty Redemption ──────────────────────────────────────────────────

type RedemptionRequest struct {
	CustomerID    string  `json:"customer_id" binding:"required"`
	Points        int     `json:"points" binding:"required"`
	DiscountValue float64 `json:"discount_value" binding:"required"`
}

func (h *MarketingHandler) RedeemPointsForDiscount(c *gin.Context) {
	var req RedemptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	code, err := h.service(c).RedeemPointsForDiscount(req.CustomerID, req.Points, req.DiscountValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to redeem points"})
		return
	}
	c.JSON(http.StatusCreated, code)
}
