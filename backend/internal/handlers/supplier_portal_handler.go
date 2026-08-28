package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

// setSupplierAuthCookies writes the supplier JWT as an HttpOnly, domain-wide session cookie.
func setSupplierAuthCookies(c *gin.Context, token string, rootDomain string) {
	isProduction := os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "staging"

	cookieDomain := ""
	if isProduction && rootDomain != "" {
		domain := rootDomain
		if idx := strings.LastIndex(domain, ":"); idx != -1 {
			domain = domain[:idx]
		}
		cookieDomain = "." + domain
	}

	maxAge := 86400 // 24 hours

	if isProduction {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}

	c.SetCookie("pux_session", token, maxAge, "/", cookieDomain, isProduction, true)
	c.SetCookie("pux_supplier_session", token, maxAge, "/", cookieDomain, isProduction, true)
}

// clearSupplierAuthCookies removes the session cookies from the browser.
func clearSupplierAuthCookies(c *gin.Context, rootDomain string) {
	isProduction := os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "staging"
	cookieDomain := ""
	if isProduction && rootDomain != "" {
		domain := rootDomain
		if idx := strings.LastIndex(domain, ":"); idx != -1 {
			domain = domain[:idx]
		}
		cookieDomain = "." + domain
	}

	if isProduction {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}

	c.SetCookie("pux_session", "", -1, "/", cookieDomain, isProduction, true)
	c.SetCookie("pux_supplier_session", "", -1, "/", cookieDomain, isProduction, true)
}

// SupplierPortalHandler handles supplier-facing portal endpoints.
type SupplierPortalHandler struct {
	db          *gorm.DB
	authService *services.AuthService
	rootDomain  string
}

func NewSupplierPortalHandler(db *gorm.DB, authService *services.AuthService, rootDomain string) *SupplierPortalHandler {
	return &SupplierPortalHandler{db: db, authService: authService, rootDomain: rootDomain}
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

	// Set HttpOnly session cookie for seamless session-based auth
	setSupplierAuthCookies(c, token, h.rootDomain)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"supplier": gin.H{
			"id":   supplier.ID,
			"name": supplier.Name,
		},
	})
}

// POST /api/v1/supplier-portal/logout
func (h *SupplierPortalHandler) Logout(c *gin.Context) {
	clearSupplierAuthCookies(c, h.rootDomain)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
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

// GET /api/v1/supplier-portal/dashboard
func (h *SupplierPortalHandler) GetDashboard(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	stats, err := h.supplierService(c).GetSupplierDashboardStats(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// POST /api/v1/supplier-portal/purchase-orders/:id/acknowledge
func (h *SupplierPortalHandler) AcknowledgePO(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	poID := c.Param("id")
	var req struct {
		Status       string `json:"status"` // confirmed, rejected, partially_received
		ExpectedDate string `json:"expected_date"`
		Notes        string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	po, err := h.supplierService(c).AcknowledgePO(supplierID, poID, req.Status, req.ExpectedDate, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, po)
}

// GET /api/v1/supplier-portal/shipments
func (h *SupplierPortalHandler) ListASNs(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	asns, err := h.supplierService(c).ListSupplierASNs(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shipments"})
		return
	}
	c.JSON(http.StatusOK, asns)
}

// POST /api/v1/supplier-portal/shipments
func (h *SupplierPortalHandler) CreateASN(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req models.SupplierASN
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	asn, err := h.supplierService(c).CreateSupplierASN(supplierID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, asn)
}

// GET /api/v1/supplier-portal/invoices
func (h *SupplierPortalHandler) ListInvoices(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	invoices, err := h.supplierService(c).ListSupplierInvoices(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invoices"})
		return
	}
	c.JSON(http.StatusOK, invoices)
}

// POST /api/v1/supplier-portal/purchase-orders/:id/invoice
func (h *SupplierPortalHandler) FlipToInvoice(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	poID := c.Param("id")
	var req struct {
		InvoiceNumber string `json:"invoice_number"`
		DueDate       string `json:"due_date"`
	}
	_ = c.ShouldBindJSON(&req)

	var due time.Time
	if req.DueDate != "" {
		due, _ = time.Parse("2006-01-02", req.DueDate)
	}

	inv, err := h.supplierService(c).FlipPOToInvoice(supplierID, poID, req.InvoiceNumber, due)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, inv)
}

// GET /api/v1/supplier-portal/catalog
func (h *SupplierPortalHandler) ListCatalog(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	products, err := h.supplierService(c).ListSupplierProducts(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch catalog"})
		return
	}
	c.JSON(http.StatusOK, products)
}

// GET /api/v1/supplier-portal/price-requests
func (h *SupplierPortalHandler) ListPriceRequests(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	reqs, err := h.supplierService(c).ListSupplierPriceRequests(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price requests"})
		return
	}
	c.JSON(http.StatusOK, reqs)
}

// POST /api/v1/supplier-portal/price-requests
func (h *SupplierPortalHandler) CreatePriceRequest(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req models.SupplierPriceChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.supplierService(c).CreateSupplierPriceRequest(supplierID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// GET /api/v1/supplier-portal/quotes
func (h *SupplierPortalHandler) ListQuotes(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	quotes, err := h.supplierService(c).ListSupplierQuotes(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quotes"})
		return
	}
	c.JSON(http.StatusOK, quotes)
}

// POST /api/v1/supplier-portal/quotes
func (h *SupplierPortalHandler) CreateQuote(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req models.SupplierQuote
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quote, err := h.supplierService(c).CreateSupplierQuote(supplierID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, quote)
}

// GET /api/v1/supplier-portal/payout-account
func (h *SupplierPortalHandler) GetPayoutAccount(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	acc, err := h.supplierService(c).GetSupplierPayoutAccount(supplierID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	c.JSON(http.StatusOK, acc)
}

// POST /api/v1/supplier-portal/payout-account
func (h *SupplierPortalHandler) SavePayoutAccount(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req models.SupplierPayoutAccount
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := h.supplierService(c).SaveSupplierPayoutAccount(supplierID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, acc)
}

// GET /api/v1/supplier-portal/messages
func (h *SupplierPortalHandler) ListMessages(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	refID := c.Query("reference_id")

	msgs, err := h.supplierService(c).ListSupplierMessages(supplierID, refID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

// POST /api/v1/supplier-portal/messages
func (h *SupplierPortalHandler) SendMessage(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req models.SupplierMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.supplierService(c).SendSupplierMessage(supplierID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, msg)
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
// Supports HttpOnly session cookies (pux_supplier_session / pux_session) and Authorization: Bearer fallback.
func SupplierAuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		// 1. Prefer HttpOnly session cookie
		if cookie, err := c.Cookie("pux_supplier_session"); err == nil && cookie != "" {
			tokenStr = cookie
		} else if cookie, err := c.Cookie("pux_session"); err == nil && cookie != "" {
			tokenStr = cookie
		}

		// 2. Fallback to Authorization: Bearer header
		if tokenStr == "" {
			authHeader := c.GetHeader("Authorization")
			if len(authHeader) >= 8 && strings.EqualFold(authHeader[:7], "bearer ") {
				tokenStr = authHeader[7:]
			}
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication session required"})
			return
		}

		claims, err := authService.ValidateToken(tokenStr)
		if err != nil || claims.Role != "supplier" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired supplier session"})
			return
		}

		c.Set(middleware.ContextKeyClaims, claims)
		c.Set(middleware.ContextKeyTenantID, claims.TenantID)
		if claims.SupplierID != nil {
			c.Set("supplier_id", claims.SupplierID.String())
		}
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
