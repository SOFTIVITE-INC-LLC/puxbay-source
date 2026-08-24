package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"github.com/softivite/puxbay/internal/utils"
	"gorm.io/gorm"
)

type InventoryHandler struct {
	db *gorm.DB
}

func NewInventoryHandler(db *gorm.DB) *InventoryHandler {
	return &InventoryHandler{db: db}
}

func (h *InventoryHandler) service(c *gin.Context) *services.InventoryService {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	return services.NewInventoryService(getDB(c, h.db), tenantID.(uuid.UUID))
}



// ---------------------------------------------------------
// Transfers
// ---------------------------------------------------------

func (h *InventoryHandler) ListTransfers(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	p := utils.GetPagination(c)

	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	transfers, total, err := h.service(c).ListTransfers(tenantID.(uuid.UUID), p.Limit, p.Offset, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transfers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  transfers,
		"total": total,
		"page":  p.Page,
		"limit": p.Limit,
	})
}

type TransferCreateRequest struct {
	ReferenceNo  string `json:"reference_no" binding:"required"`
	FromBranchID string `json:"from_branch_id" binding:"required"`
	ToBranchID   string `json:"to_branch_id" binding:"required"`
	Notes        string `json:"notes"`
	Items        []struct {
		ProductID string  `json:"product_id" binding:"required"`
		Quantity  float64 `json:"quantity" binding:"required"`
	} `json:"items" binding:"required,min=1"`
}

func (h *InventoryHandler) CreateTransfer(c *gin.Context) {
	rawUserID, _ := c.Get(middleware.ContextKeyUserID)
	userID := rawUserID.(uuid.UUID)

	var req TransferCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fromBranch, _ := uuid.Parse(req.FromBranchID)
	toBranch, _ := uuid.Parse(req.ToBranchID)

	input := services.TransferCreateInput{
		ReferenceNo:  req.ReferenceNo,
		FromBranchID: fromBranch,
		ToBranchID:   toBranch,
		Notes:        req.Notes,
		CreatedByID:  userID,
	}

	for _, item := range req.Items {
		productID, _ := uuid.Parse(item.ProductID)
		input.Items = append(input.Items, services.TransferItemInput{
			ProductID: productID,
			Quantity:  item.Quantity,
		})
	}

	transfer, err := h.service(c).CreateTransfer(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transfer"})
		return
	}

	c.JSON(http.StatusCreated, transfer)
}

// ---------------------------------------------------------
// Purchase Orders
// ---------------------------------------------------------

func (h *InventoryHandler) ListPOs(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	p := utils.GetPagination(c)

	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	pos, total, err := h.service(c).ListPOs(tenantID.(uuid.UUID), p.Limit, p.Offset, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch POs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  pos,
		"total": total,
		"page":  p.Page,
		"limit": p.Limit,
	})
}

type POCreateRequest struct {
	PONumber   string `json:"po_number" binding:"required"`
	SupplierID string `json:"supplier_id" binding:"required"`
	Notes      string `json:"notes"`
	Items      []struct {
		ProductID       string  `json:"product_id" binding:"required"`
		QuantityOrdered float64 `json:"quantity_ordered" binding:"required"`
		UnitCost        float64 `json:"unit_cost" binding:"required"`
	} `json:"items" binding:"required,min=1"`
}

func (h *InventoryHandler) CreatePO(c *gin.Context) {
	var req POCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	supplierUUID, _ := uuid.Parse(req.SupplierID)
	var branchUUID uuid.UUID
	if ctxBranchID, ok := middleware.GetBranchID(c); ok {
		branchUUID = *ctxBranchID
	}

	input := services.POCreateInput{
		PONumber:   req.PONumber,
		SupplierID: supplierUUID,
		BranchID:   branchUUID,
		Notes:      req.Notes,
	}

	for _, item := range req.Items {
		productID, _ := uuid.Parse(item.ProductID)
		input.Items = append(input.Items, services.POItemInput{
			ProductID:       productID,
			QuantityOrdered: item.QuantityOrdered,
			UnitCost:        item.UnitCost,
		})
	}

	po, err := h.service(c).CreatePO(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create PO"})
		return
	}

	// Audit log — fire-and-forget
	tenantRaw, _ := c.Get(middleware.ContextKeyTenantID)
	userRaw, _ := c.Get(middleware.ContextKeyUserID)
	tenantID, _ := tenantRaw.(uuid.UUID)
	userID, _ := userRaw.(uuid.UUID)
	auditAsync(h.db, tenantID, userID, "CREATE_PURCHASE_ORDER", "purchase_orders", c.ClientIP(), map[string]interface{}{"po_id": po.ID, "po_number": po.PONumber, "branch_id": branchUUID})

	c.JSON(http.StatusCreated, po)
}

func (h *InventoryHandler) ListStocktakes(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	stocktakes, total, err := h.service(c).ListStocktakes(tenantID.(uuid.UUID), limit, offset, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stocktakes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  stocktakes,
		"total": total,
	})
}

func (h *InventoryHandler) GetTransfer(c *gin.Context) {
	id := c.Param("id")
	transfer, err := h.service(c).GetTransfer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found"})
		return
	}
	c.JSON(http.StatusOK, transfer)
}

func (h *InventoryHandler) ApproveTransfer(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).UpdateTransferStatus(id, "approved"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve transfer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

func (h *InventoryHandler) ShipTransfer(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).UpdateTransferStatus(id, "shipped"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ship transfer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "shipped"})
}

func (h *InventoryHandler) ReceiveTransfer(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).UpdateTransferStatus(id, "received"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to receive transfer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func (h *InventoryHandler) GetPO(c *gin.Context) {
	id := c.Param("id")
	po, err := h.service(c).GetPO(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PO not found"})
		return
	}
	c.JSON(http.StatusOK, po)
}

type POReceiveRequest struct {
	Items []POReceiveItemReq `json:"items" binding:"required"`
}

type POReceiveItemReq struct {
	ItemID           string  `json:"item_id" binding:"required"`
	QuantityReceived float64 `json:"quantity_received" binding:"required"`
}

func (h *InventoryHandler) ReceivePO(c *gin.Context) {
	id := c.Param("id")
	var req POReceiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.POReceiveInput{}
	for _, item := range req.Items {
		itemID, _ := uuid.Parse(item.ItemID)
		input.Items = append(input.Items, services.POReceiveItemInput{
			ItemID:           itemID,
			QuantityReceived: item.QuantityReceived,
		})
	}

	if err := h.service(c).ReceivePO(id, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to receive PO", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

type CreateStocktakeReq struct {
	Name string `json:"name" binding:"required"`
}

func (h *InventoryHandler) CreateStocktake(c *gin.Context) {
	var req CreateStocktakeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var branchID uuid.UUID
	if ctxBranchID, ok := middleware.GetBranchID(c); ok {
		branchID = *ctxBranchID
	}
	st, err := h.service(c).CreateStocktake(services.StocktakeInput{
		Name:     req.Name,
		BranchID: branchID,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create stocktake"})
		return
	}

	// Audit log — fire-and-forget
	tenantRaw, _ := c.Get(middleware.ContextKeyTenantID)
	userRaw, _ := c.Get(middleware.ContextKeyUserID)
	tenantID, _ := tenantRaw.(uuid.UUID)
	userID, _ := userRaw.(uuid.UUID)
	auditAsync(h.db, tenantID, userID, "CREATE_STOCKTAKE", "stocktakes", c.ClientIP(), map[string]interface{}{"stocktake_id": st.ID, "name": st.Name, "branch_id": branchID})

	c.JSON(http.StatusCreated, st)
}

func (h *InventoryHandler) GetStocktake(c *gin.Context) {
	id := c.Param("id")
	st, err := h.service(c).GetStocktake(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stocktake not found"})
		return
	}
	c.JSON(http.StatusOK, st)
}

func (h *InventoryHandler) FinalizeStocktake(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).FinalizeStocktake(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize stocktake"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "finalized"})
}

func (h *InventoryHandler) ListMovements(c *gin.Context) {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	movements, total, err := h.service(c).ListMovements(tenantID.(uuid.UUID), limit, offset, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list movements"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  movements,
		"total": total,
	})
}

type ReceiveStockReq struct {
	ProductID string  `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required"`
	Reason    string  `json:"reason"`
}

func (h *InventoryHandler) ReceiveStock(c *gin.Context) {
	var req ReceiveStockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	rawUserID, _ := c.Get(middleware.ContextKeyUserID)
	userID := rawUserID.(uuid.UUID)
	pID := &userID

	tenantID, _ := c.Get(middleware.ContextKeyTenantID)

	var bID uuid.UUID
	if ctxBranchID, ok := middleware.GetBranchID(c); ok {
		bID = *ctxBranchID
	}

	if err := h.service(c).ReceiveStock(req.ProductID, req.Quantity, req.Reason, pID, tenantID.(uuid.UUID), bID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to receive stock"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "received_stock"})
}

func (h *InventoryHandler) LowStockAlerts(c *gin.Context) {
	alerts, err := h.service(c).LowStockAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch low stock alerts"})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

// GetProductHistory returns the audit history of inventory changes for a product
func (h *InventoryHandler) GetProductHistory(c *gin.Context) {
	productID := c.Param("id")
	history, err := h.service(c).GetProductHistory(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	var formattedHistory []gin.H
	for _, h := range history {
		// Map backend reason to frontend expected action
		action := "edit"
		if h.Quantity > 0 {
			action = "add"
		} else if h.Quantity < 0 {
			action = "remove"
		}

		absQty := h.Quantity
		if absQty < 0 {
			absQty = -absQty
		}

		formattedHistory = append(formattedHistory, gin.H{
			"date":     h.CreatedAt,
			"action":   action,
			"quantity": absQty,
			"user":     "System", // Can be extended to fetch real user name
			"notes":    h.Reason,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"product_id": productID,
		"history":    formattedHistory,
	})
}

// GetProductComponents returns the composite items (BOM/Recipe) for a product.
// Gap #16: Queries real ProductComponent table instead of returning hardcoded data.
func (h *InventoryHandler) GetProductComponents(c *gin.Context) {
	productID := c.Param("id")

	var components []models.ProductComponent
	if err := getDB(c, h.db).Where("parent_product_id = ?", productID).Preload("ComponentProduct").Find(&components).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch components"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product_id": productID,
		"components": components,
	})
}

func (h *InventoryHandler) PublicStocktakeSession(c *gin.Context) {
	token := c.Param("token")
	var session models.StocktakeSession
	if err := getDB(c, h.db).Where("access_token = ?", token).Preload("Branch").First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *InventoryHandler) PublicStocktakeScan(c *gin.Context) {
	token := c.Param("token")
	query := c.Query("q")

	results, err := h.service(c).ScanStocktakeProduct(token, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

type StocktakeUpdateReq struct {
	ProductID string  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	Mode      string  `json:"mode"` // 'set' or 'add'
}

func (h *InventoryHandler) PublicStocktakeUpdate(c *gin.Context) {
	token := c.Param("token")
	var req StocktakeUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	newCount, err := h.service(c).UpdateStocktakeCount(token, productID, req.Quantity, req.Mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"product_id": req.ProductID,
		"new_count":  newCount,
	})
}

// ---------------------------------------------------------
// Batch & Expiry Tracking
// ---------------------------------------------------------

type BatchCreateRequest struct {
	BatchNumber     string  `json:"batch_number" binding:"required"`
	Quantity        float64 `json:"quantity" binding:"required"`
	ExpiryDate      string  `json:"expiry_date"`
	ManufactureDate string  `json:"manufacture_date"`
}

func (h *InventoryHandler) ListBatches(c *gin.Context) {
	productID := c.Param("id")
	batches, err := h.service(c).ListBatches(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch batches"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": batches, "total": len(batches)})
}

func (h *InventoryHandler) CreateBatch(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req BatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var branchID uuid.UUID
	if ctxBranchID, ok := middleware.GetBranchID(c); ok {
		branchID = *ctxBranchID
	}

	input := services.BatchCreateInput{
		ProductID:   productID,
		BranchID:    branchID,
		BatchNumber: req.BatchNumber,
		Quantity:    req.Quantity,
	}
	if req.ExpiryDate != "" {
		if t, err := time.Parse("2006-01-02", req.ExpiryDate); err == nil {
			input.ExpiryDate = &t
		}
	}
	if req.ManufactureDate != "" {
		if t, err := time.Parse("2006-01-02", req.ManufactureDate); err == nil {
			input.ManufactureDate = &t
		}
	}

	batch, err := h.service(c).CreateBatch(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create batch"})
		return
	}
	c.JSON(http.StatusCreated, batch)
}

func (h *InventoryHandler) UpdateBatch(c *gin.Context) {
	batchID := c.Param("batchId")

	var req BatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.BatchCreateInput{
		BatchNumber: req.BatchNumber,
		Quantity:    req.Quantity,
	}
	if req.ExpiryDate != "" {
		if t, err := time.Parse("2006-01-02", req.ExpiryDate); err == nil {
			input.ExpiryDate = &t
		}
	}
	if req.ManufactureDate != "" {
		if t, err := time.Parse("2006-01-02", req.ManufactureDate); err == nil {
			input.ManufactureDate = &t
		}
	}

	batch, err := h.service(c).UpdateBatch(batchID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update batch"})
		return
	}
	c.JSON(http.StatusOK, batch)
}

func (h *InventoryHandler) DeleteBatch(c *gin.Context) {
	batchID := c.Param("batchId")
	if err := h.service(c).DeleteBatch(batchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete batch"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *InventoryHandler) ListExpiringBatches(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	batches, err := h.service(c).ListExpiringBatches(branchID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch expiring batches"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": batches, "total": len(batches)})
}
