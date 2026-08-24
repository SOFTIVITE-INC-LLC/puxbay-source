package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/middleware"
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
