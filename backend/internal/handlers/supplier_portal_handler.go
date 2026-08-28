package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

// SupplierPortalHandler handles supplier-facing portal endpoints.
type SupplierPortalHandler struct {
	db          *gorm.DB
	authService *services.AuthService
}

func NewSupplierPortalHandler(db *gorm.DB, authService *services.AuthService) *SupplierPortalHandler {
	return &SupplierPortalHandler{db: db, authService: authService}
}

func (h *SupplierPortalHandler) supplierService(c *gin.Context) *services.SupplierService {
	return services.NewSupplierService(getDB(c, h.db))
}

// POST /api/v1/supplier-portal/login
func (h *SupplierPortalHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	supplier, err := h.supplierService(c).AuthenticateSupplierPortal(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Get the real tenant UUID from context (set by TenantMiddleware from subdomain lookup)
	var tenantID uuid.UUID
	if tid, exists := c.Get(middleware.ContextKeyTenantID); exists {
		if id, ok := tid.(uuid.UUID); ok {
			tenantID = id
		}
	}
	if tenantID == uuid.Nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not resolve tenant context"})
		return
	}

	token, err := h.authService.GenerateSupplierToken(supplier.ID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"supplier": gin.H{
			"id":   supplier.ID,
			"name": supplier.Name,
		},
	})
}

// GET /api/v1/supplier-portal/me
func (h *SupplierPortalHandler) Me(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	supplier, err := h.supplierService(c).GetSupplier(supplierID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
		return
	}
	c.JSON(http.StatusOK, supplier)
}

// GET /api/v1/supplier-portal/purchase-orders
func (h *SupplierPortalHandler) ListPurchaseOrders(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	orders, err := h.supplierService(c).ListSupplierPurchaseOrders(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// resolveSupplierID extracts the supplier_id from the JWT claims.
func (h *SupplierPortalHandler) resolveSupplierID(c *gin.Context) string {
	claims, ok := c.Get(middleware.ContextKeyClaims)
	if !ok {
		return ""
	}
	cl, ok := claims.(*services.Claims)
	if !ok || cl.SupplierID == nil {
		return ""
	}
	return cl.SupplierID.String()
}

// SupplierAuthMiddleware validates that the JWT has role=supplier and supplier_id set.
func SupplierAuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 8 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}
		tokenStr := authHeader[7:] // strip "Bearer "

		claims, err := authService.ValidateToken(tokenStr)
		if err != nil || claims.Role != "supplier" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired supplier token"})
			return
		}

		c.Set(middleware.ContextKeyClaims, claims)
		c.Set(middleware.ContextKeyTenantID, claims.TenantID)
		c.Next()
	}
}

// --- Admin-side: invite supplier to portal ---

// POST /api/v1/suppliers/:id/invite
func (h *SupplierPortalHandler) InviteUser(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	supplier, err := h.supplierService(c).InvitePortalAccess(id, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":      "Portal access granted",
		"portal_email": supplier.PortalEmail,
	})
}

// --- Updated ledger endpoint with date support ---

// POST /api/v1/suppliers/:id/ledger/full
func (h *SupplierPortalHandler) AddLedgerFull(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		EntryType       string  `json:"entry_type" binding:"required"`
		Amount          float64 `json:"amount" binding:"required,gt=0"`
		ReferenceID     *string `json:"reference_id"`
		Notes           *string `json:"notes"`
		TransactionDate *string `json:"transaction_date"` // ISO 8601 date string
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.UpdatedLedgerEntryInput{
		EntryType:   req.EntryType,
		Amount:      req.Amount,
		ReferenceID: req.ReferenceID,
		Notes:       req.Notes,
	}

	if req.TransactionDate != nil && *req.TransactionDate != "" {
		t, err := time.Parse("2006-01-02", *req.TransactionDate)
		if err == nil {
			input.TransactionDate = &t
		}
	}

	entry, err := h.supplierService(c).AddLedgerEntryFull(id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, entry)
}

// DELETE /api/v1/suppliers/:id/products/:productId
func (h *SupplierPortalHandler) RemoveProduct(c *gin.Context) {
	supplierID := c.Param("id")
	productID := c.Param("productId")

	if err := services.NewSupplierService(getDB(c, h.db)).RemoveSupplierProduct(supplierID, productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product removed from catalog"})
}
