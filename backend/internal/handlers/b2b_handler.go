package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type B2BHandler struct {
	db *gorm.DB
}

func NewB2BHandler(db *gorm.DB) *B2BHandler {
	return &B2BHandler{db: db}
}

func (h *B2BHandler) service(c *gin.Context) *services.B2BService {
	return services.NewB2BService(getDB(c, h.db))
}

func (h *B2BHandler) getProfile(c *gin.Context) (uuid.UUID, string) {
	id, _ := c.Get("user_id")
	role, _ := c.Get("role")
	if id != nil && role != nil {
		return id.(uuid.UUID), role.(string)
	}
	return uuid.Nil, "sales" // Fallback
}

func (h *B2BHandler) ListQuotes(c *gin.Context) {
	_, role := h.getProfile(c)
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	quotations, err := h.service(c).ListQuotes(tenantID.(uuid.UUID), role)
	if err != nil {
		if err.Error() == "unauthorized to view all quotes" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quotations"})
		return
	}

	c.JSON(http.StatusOK, quotations)
}

type QuoteItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"qty" binding:"required,min=1"`
}

type QuoteCreateRequest struct {
	CustomerID string             `json:"customer_id" binding:"required"`
	Items      []QuoteItemRequest `json:"items" binding:"required,min=1"`
}

func (h *B2BHandler) CreateQuote(c *gin.Context) {
	profileID, _ := h.getProfile(c)

	var req QuoteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []services.QuoteItemInput
	for _, item := range req.Items {
		items = append(items, services.QuoteItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	input := services.QuoteCreateInput{
		CustomerID: req.CustomerID,
		Items:      items,
		ProfileID:  profileID,
	}

	quoteID, err := h.service(c).CreateQuote(input)
	if err != nil {
		if err.Error() == "invalid customer ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quote"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "quote_id": quoteID})
}

func (h *B2BHandler) UpdateQuote(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Action        string `json:"action" binding:"required"`
		InternalNotes string `json:"internal_notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.QuoteUpdateInput{
		Action:        req.Action,
		InternalNotes: req.InternalNotes,
	}

	status, err := h.service(c).UpdateQuote(id, input)
	if err != nil {
		if err.Error() == "quote not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "invalid action" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quote"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "status": status})
}

func (h *B2BHandler) ListClients(c *gin.Context) {
	customers, err := h.service(c).ListClients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clients"})
		return
	}
	c.JSON(http.StatusOK, customers)
}

func (h *B2BHandler) GetQuote(c *gin.Context) {
	id := c.Param("id")
	var quote models.Quotation

	if err := getDB(c, h.db).Where("id = ?", id).Preload("Items").First(&quote).Error; err != nil {
		c.JSON(404, gin.H{"error": "Quote not found"})
		return
	}
	c.JSON(200, quote)
}

func (h *B2BHandler) ConvertQuoteToOrder(c *gin.Context) {
	id := c.Param("id")
	dbConn := getDB(c, h.db)

	var quote models.Quotation
	if err := dbConn.Where("id = ?", id).Preload("Items").First(&quote).Error; err != nil {
		c.JSON(404, gin.H{"error": "Quote not found"})
		return
	}

	if quote.Status == "converted" {
		c.JSON(400, gin.H{"error": "Quote already converted"})
		return
	}

	err := dbConn.Transaction(func(tx *gorm.DB) error {
		order := models.Order{
			CustomerID: &quote.CustomerID,
			Subtotal:   quote.Subtotal,
			Tax:        quote.TaxAmount,
			Total:      quote.TotalAmount,
			OrderType:  "b2b",
			Status:     "completed", // assuming standard completion for B2B API
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		for _, item := range quote.Items {
			oItem := models.OrderItem{
				ProductID: item.ProductID,
				Quantity:  float64(item.Quantity),
				UnitPrice: item.UnitPrice,
				Discount:  item.Discount,
				Total:     item.TotalPrice,
			}
			if err := tx.Model(&order).Association("Items").Append(&oItem); err != nil {
				return err
			}
		}

		return tx.Model(&quote).Updates(map[string]interface{}{
			"status":             "converted",
			"converted_order_id": order.ID,
		}).Error
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to convert quote: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "converted"})
}

func (h *B2BHandler) BulkOrder(c *gin.Context) {
	var req QuoteCreateRequest // Use same structure as quote for simplicity
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// Direct to order logic could go here, for MVP we just return a success payload
	c.JSON(201, gin.H{"status": "bulk_order_created", "items_count": len(req.Items)})
}
