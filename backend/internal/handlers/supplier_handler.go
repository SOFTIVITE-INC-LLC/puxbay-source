package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"github.com/softivite/puxbay/internal/utils"
	"gorm.io/gorm"
)

type SupplierHandler struct {
	db *gorm.DB
}

func NewSupplierHandler(db *gorm.DB) *SupplierHandler {
	return &SupplierHandler{db: db}
}

func (h *SupplierHandler) service(c *gin.Context) *services.SupplierService {
	return services.NewSupplierService(getDB(c, h.db))
}

func (h *SupplierHandler) List(c *gin.Context) {
	suppliers, err := h.service(c).ListSuppliers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suppliers"})
		return
	}
	role, _ := c.Get(middleware.ContextKeyRole)
	permissions := c.GetStringSlice(middleware.ContextKeyPermissions)
	maskedSuppliers := utils.MaskCollection(suppliers, role.(string), permissions)

	c.JSON(http.StatusOK, maskedSuppliers)
}

type SupplierCreateRequest struct {
	Name          string  `json:"name" binding:"required"`
	ContactPerson *string `json:"contact_person"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
	Address       *string `json:"address"`
	TaxNumber     *string `json:"tax_number"`
	PaymentTerms  *string `json:"payment_terms"`
	Notes         *string `json:"notes"`
}

func (h *SupplierHandler) Create(c *gin.Context) {
	var req SupplierCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.SupplierInput{
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		TaxNumber:     req.TaxNumber,
		PaymentTerms:  req.PaymentTerms,
		Notes:         req.Notes,
	}

	supplier, err := h.service(c).CreateSupplier(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create supplier"})
		return
	}

	c.JSON(http.StatusCreated, supplier)
}

func (h *SupplierHandler) Get(c *gin.Context) {
	id := c.Param("id")

	supplier, err := h.service(c).GetSupplier(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	role, _ := c.Get(middleware.ContextKeyRole)
	permissions := c.GetStringSlice(middleware.ContextKeyPermissions)
	maskedSupplier := utils.MaskCollection(supplier, role.(string), permissions)

	c.JSON(http.StatusOK, maskedSupplier)
}

func (h *SupplierHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req SupplierCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.SupplierInput{
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		TaxNumber:     req.TaxNumber,
		PaymentTerms:  req.PaymentTerms,
		Notes:         req.Notes,
	}

	supplier, err := h.service(c).UpdateSupplier(id, input)
	if err != nil {
		if err.Error() == "supplier not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update supplier"})
		return
	}

	c.JSON(http.StatusOK, supplier)
}

func (h *SupplierHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service(c).DeleteSupplier(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete supplier"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *SupplierHandler) ListProducts(c *gin.Context) {
	id := c.Param("id")
	products, err := h.service(c).ListSupplierProducts(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch supplier products"})
		return
	}
	c.JSON(http.StatusOK, products)
}

type SupplierProductCreateRequest struct {
	ProductID   string  `json:"product_id" binding:"required"`
	SupplierSKU string  `json:"supplier_sku"`
	UnitCost    float64 `json:"unit_cost" binding:"required"`
}

func (h *SupplierHandler) AddProduct(c *gin.Context) {
	id := c.Param("id")
	var req SupplierProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.SupplierProductInput{
		ProductID:   req.ProductID,
		SupplierSKU: req.SupplierSKU,
		UnitCost:    req.UnitCost,
	}

	product, err := h.service(c).AddSupplierProduct(id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add supplier product"})
		return
	}
	c.JSON(http.StatusCreated, product)
}

func (h *SupplierHandler) ListLedger(c *gin.Context) {
	id := c.Param("id")
	entries, err := h.service(c).ListLedgerEntries(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ledger entries"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

type LedgerEntryCreateRequest struct {
	EntryType   string  `json:"entry_type" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	ReferenceID *string `json:"reference_id"`
	Notes       *string `json:"notes"`
}

func (h *SupplierHandler) AddLedger(c *gin.Context) {
	id := c.Param("id")
	var req LedgerEntryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.LedgerEntryInput{
		EntryType:   req.EntryType,
		Amount:      req.Amount,
		ReferenceID: req.ReferenceID,
		Notes:       req.Notes,
	}

	entry, err := h.service(c).AddLedgerEntry(id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add ledger entry", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, entry)
}

// POST /api/v1/suppliers/invoices/:id/disburse (Admin 1-Click Payout)
func (h *SupplierHandler) DisburseInvoicePayout(c *gin.Context) {
	invoiceID := c.Param("id")
	res, err := h.service(c).InitiateEarlyPayout("", invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /api/v1/suppliers/price-requests (Admin queue)
func (h *SupplierHandler) ListPriceProposals(c *gin.Context) {
	status := c.Query("status")
	requests, err := h.service(c).ListAllPriceRequests(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, requests)
}

// GET /api/v1/suppliers/invoices (Admin Accounts Payable queue)
func (h *SupplierHandler) ListAllInvoices(c *gin.Context) {
	status := c.Query("status")
	invoices, err := h.service(c).ListAllInvoices(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, invoices)
}

// GET /api/v1/suppliers/:id/details (Full 360-degree supplier dossier)
func (h *SupplierHandler) GetDetails(c *gin.Context) {
	id := c.Param("id")
	details, err := h.service(c).GetSupplierDetails(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, details)
}

// POST /api/v1/suppliers/price-requests/:id/approve
func (h *SupplierHandler) ApprovePriceProposal(c *gin.Context) {
	reqID := c.Param("id")
	var body struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&body)

	if err := h.service(c).ApprovePriceRequest(reqID, body.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Price adjustment approved"})
}

// POST /api/v1/suppliers/price-requests/:id/reject
func (h *SupplierHandler) RejectPriceProposal(c *gin.Context) {
	reqID := c.Param("id")
	var body struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&body)

	if err := h.service(c).RejectPriceRequest(reqID, body.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Price adjustment rejected"})
}

// GET /api/v1/suppliers/rmas (Admin/Manager RMA view)
func (h *SupplierHandler) ListRMAs(c *gin.Context) {
	branchID := c.Query("branch_id")
	var bidPtr *string
	if branchID != "" {
		bidPtr = &branchID
	}
	rmas, err := h.service(c).ListAllRMAs(bidPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rmas)
}

// POST /api/v1/suppliers/rmas (Manager logs defect claim)
func (h *SupplierHandler) CreateRMA(c *gin.Context) {
	var req models.SupplierRMA
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rma, err := h.service(c).CreateSupplierRMA(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rma)
}

// POST /api/v1/suppliers/rmas/:id/resolve (Admin issues credit note or marks refunded)
func (h *SupplierHandler) ResolveRMA(c *gin.Context) {
	rmaID := c.Param("id")
	var body struct {
		Status          string  `json:"status"`
		ResolutionNotes string  `json:"resolution_notes"`
		CreditAmount    float64 `json:"credit_amount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rma, err := h.service(c).ResolveSupplierRMA(rmaID, body.Status, body.ResolutionNotes, body.CreditAmount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rma)
}

// POST /api/v1/suppliers/announcements
func (h *SupplierHandler) CreateAnnouncement(c *gin.Context) {
	var req models.SupplierAnnouncement
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ann, err := h.service(c).CreateSupplierAnnouncement(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ann)
}

// GET /api/v1/suppliers/dock-slots
func (h *SupplierHandler) ListBranchDockSlots(c *gin.Context) {
	branchID := c.Query("branch_id")
	date := c.Query("date")
	slots, err := h.service(c).ListBranchDeliverySlots(branchID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, slots)
}
