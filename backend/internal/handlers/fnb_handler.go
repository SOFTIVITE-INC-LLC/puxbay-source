package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"github.com/softivite/puxbay/internal/websocket"
	"gorm.io/gorm"
)

type FnBHandler struct {
	db  *gorm.DB
	hub *websocket.Hub
}

func NewFnBHandler(db *gorm.DB, hub *websocket.Hub) *FnBHandler {
	return &FnBHandler{db: db, hub: hub}
}

func (h *FnBHandler) service(c *gin.Context) *services.FNBService {
	return services.NewFNBService(getDB(c, h.db))
}

func (h *FnBHandler) ListTables(c *gin.Context) {
	tables, err := h.service(c).ListTables()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tables"})
		return
	}
	c.JSON(http.StatusOK, tables)
}

func (h *FnBHandler) CreateTable(c *gin.Context) {
	var req models.DiningTable
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service(c).CreateTable(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create table"})
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (h *FnBHandler) UpdateTableStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	table, err := h.service(c).UpdateTableStatus(id, req.Status)
	if err != nil {
		if err.Error() == "table not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update table status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": table.Status})
}

func (h *FnBHandler) ListKDS(c *gin.Context) {
	tickets, err := h.service(c).ListKDS()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch KDS tickets"})
		return
	}
	c.JSON(http.StatusOK, tickets)
}

func (h *FnBHandler) AdvanceTicketStatus(c *gin.Context) {
	id := c.Param("id")

	ticket, err := h.service(c).AdvanceTicketStatus(id)
	if err != nil {
		if err.Error() == "ticket not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "cannot advance from current status" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to advance ticket status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"new_status": ticket.Status})
}

func (h *FnBHandler) CreateSplitBill(c *gin.Context) {
	// Simulated split bill logic
	c.JSON(201, gin.H{"status": "split_bill_created"})
}

func (h *FnBHandler) GetSplitBill(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{"id": id, "status": "active"})
}
