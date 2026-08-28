package handlers

import (
	"encoding/json"
	"fmt"
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
	db           *gorm.DB
	authService  *services.AuthService
	emailService *services.EmailService
	smsService   *services.SMSService
	rootDomain   string
}

func NewSupplierPortalHandler(db *gorm.DB, authService *services.AuthService, rootDomain string, emailSvc *services.EmailService, smsSvc *services.SMSService) *SupplierPortalHandler {
	return &SupplierPortalHandler{
		db:           db,
		authService:  authService,
		emailService: emailSvc,
		smsService:   smsSvc,
		rootDomain:   rootDomain,
	}
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

	currency := h.resolveTenantCurrency(c)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"supplier": gin.H{
			"id":       supplier.ID,
			"name":     supplier.Name,
			"currency": currency,
		},
		"currency": currency,
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
	currency := h.resolveTenantCurrency(c)
	c.JSON(http.StatusOK, gin.H{
		"id":              supplier.ID,
		"name":            supplier.Name,
		"email":           supplier.Email,
		"phone":           supplier.Phone,
		"address":         supplier.Address,
		"tax_number":      supplier.TaxNumber,
		"payment_terms":   supplier.PaymentTerms,
		"credit_balance":  supplier.CreditBalance,
		"credit_limit":    supplier.CreditLimit,
		"portal_email":    supplier.PortalEmail,
		"is_active":       supplier.IsActive,
		"currency":        currency,
		"tenant_currency": currency,
	})
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
	stats["currency"] = h.resolveTenantCurrency(c)
	stats["tenant_currency"] = stats["currency"]
	c.JSON(http.StatusOK, stats)
}

func (h *SupplierPortalHandler) resolveTenantCurrency(c *gin.Context) string {
	var tenantID uuid.UUID
	if tid, exists := c.Get(middleware.ContextKeyTenantID); exists {
		if id, ok := tid.(uuid.UUID); ok {
			tenantID = id
		}
	} else if tid, exists := c.Get("tenant_id"); exists {
		if id, ok := tid.(uuid.UUID); ok {
			tenantID = id
		}
	}

	if tenantID != uuid.Nil {
		var tenant models.Tenant
		if err := h.db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err == nil {
			var metadata struct {
				Currency string `json:"currency"`
			}
			if len(tenant.Metadata) > 0 {
				_ = json.Unmarshal(tenant.Metadata, &metadata)
			}
			if metadata.Currency != "" {
				return metadata.Currency
			}
		}
	}
	return "GHS"
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

// GET /api/v1/supplier-portal/forecasts
func (h *SupplierPortalHandler) GetForecasts(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	forecasts, err := h.supplierService(c).GetDemandForecasts(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch forecasts"})
		return
	}
	c.JSON(http.StatusOK, forecasts)
}

// GET /api/v1/supplier-portal/team
func (h *SupplierPortalHandler) GetTeam(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	members, err := h.supplierService(c).GetTeamMembers(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team members"})
		return
	}
	c.JSON(http.StatusOK, members)
}

// POST /api/v1/supplier-portal/team/invite
func (h *SupplierPortalHandler) InviteTeamMember(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req models.SupplierTeamMember
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.supplierService(c).InviteTeamMember(supplierID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, member)
}

// POST /api/v1/supplier-portal/invoices/:id/payout
func (h *SupplierPortalHandler) InitiateEarlyPayout(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	invoiceID := c.Param("id")

	res, err := h.supplierService(c).InitiateEarlyPayout(supplierID, invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// POST /api/v1/supplier-portal/receive-scan
func (h *SupplierPortalHandler) ReceiveScan(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req struct {
		QRPayload string `json:"qr_payload" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.supplierService(c).ProcessQRScan(supplierID, req.QRPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /api/v1/supplier-portal/rmas
func (h *SupplierPortalHandler) ListRMAs(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	rmas, err := h.supplierService(c).ListSupplierRMAs(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch RMAs"})
		return
	}
	c.JSON(http.StatusOK, rmas)
}

// POST /api/v1/supplier-portal/rmas/:id/replace
func (h *SupplierPortalHandler) ResolveRMAReplacement(c *gin.Context) {
	rmaID := c.Param("id")
	var req struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&req)

	res, err := h.supplierService(c).ResolveSupplierRMA(rmaID, "replacement_dispatched", req.Notes, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /api/v1/supplier-portal/dock-slots
func (h *SupplierPortalHandler) ListDockSlots(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	slots, err := h.supplierService(c).ListDeliverySlots(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch delivery slots"})
		return
	}
	c.JSON(http.StatusOK, slots)
}

// POST /api/v1/supplier-portal/dock-slots
func (h *SupplierPortalHandler) BookDockSlot(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req models.SupplierDeliverySlot
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	supUUID, _ := uuid.Parse(supplierID)
	req.SupplierID = supUUID

	slot, err := h.supplierService(c).BookDeliverySlot(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, slot)
}

// GET /api/v1/supplier-portal/documents
func (h *SupplierPortalHandler) ListDocuments(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	docs, err := h.supplierService(c).ListSupplierDocuments(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch compliance documents"})
		return
	}
	c.JSON(http.StatusOK, docs)
}

// POST /api/v1/supplier-portal/documents
func (h *SupplierPortalHandler) UploadDocument(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req models.SupplierDocument
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doc, err := h.supplierService(c).UploadSupplierDocument(supplierID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, doc)
}

// GET /api/v1/supplier-portal/announcements
func (h *SupplierPortalHandler) ListAnnouncements(c *gin.Context) {
	anns, err := h.supplierService(c).ListSupplierAnnouncements()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch announcements"})
		return
	}
	c.JSON(http.StatusOK, anns)
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

	// Resolve tenant info for portal URL and notification branding
	var tenantName, tenantSubdomain string
	if tid, exists := c.Get(middleware.ContextKeyTenantID); exists {
		if tenantID, ok := tid.(uuid.UUID); ok {
			var tenant models.Tenant
			if dbErr := h.db.Where("id = ?", tenantID).First(&tenant).Error; dbErr == nil {
				tenantName = tenant.Name
				tenantSubdomain = tenant.Subdomain
			}
		}
	}
	if tenantName == "" {
		tenantName = "Your Merchant"
	}

	// Build the portal login URL
	protocol := "https"
	if h.rootDomain == "localhost" || strings.HasPrefix(h.rootDomain, "localhost:") {
		protocol = "http"
	}
	var portalURL string
	if tenantSubdomain != "" {
		portalURL = fmt.Sprintf("%s://%s.%s/supplier-portal/login", protocol, tenantSubdomain, h.rootDomain)
	} else {
		portalURL = fmt.Sprintf("%s://%s/supplier-portal/login", protocol, h.rootDomain)
	}

	supplierName := supplier.Name

	// Send email notification asynchronously
	if h.emailService != nil {
		go h.emailService.SendSupplierPortalWelcomeEmail(
			req.Email,
			supplierName,
			tenantName,
			portalURL,
			req.Email,
			req.Password,
		)
	}

	// Send SMS notification asynchronously if supplier has a phone number
	if h.smsService != nil && supplier.Phone != nil && *supplier.Phone != "" {
		smsMsg := fmt.Sprintf(
			"Hi %s, %s has invited you to their Supplier Portal. Login: %s | Email: %s | Pass: %s",
			supplierName, tenantName, portalURL, req.Email, req.Password,
		)
		go h.smsService.SendSMS([]string{*supplier.Phone}, smsMsg)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Portal access granted. Credentials have been sent via email and SMS.",
		"portal_email": supplier.PortalEmail,
		"portal_url":   portalURL,
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

// ── GET /api/v1/supplier-portal/shipments/:id/shipping-label ──
func (h *SupplierPortalHandler) GetShippingLabel(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	asnID := c.Param("id")
	data, err := h.supplierService(c).GetShippingLabelData(supplierID, asnID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// ── POST /api/v1/supplier-portal/catalog/bulk ──
func (h *SupplierPortalHandler) BulkImportCatalog(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}

	var req struct {
		Items []services.BulkCatalogItemInput `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	count, err := h.supplierService(c).BulkImportCatalog(supplierID, req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        fmt.Sprintf("Successfully imported %d catalog items", count),
		"imported_count": count,
	})
}

// ── GET /api/v1/supplier-portal/invoices/:id/three-way-match ──
func (h *SupplierPortalHandler) GetThreeWayMatch(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	invoiceID := c.Param("id")
	audit, err := h.supplierService(c).CalculateThreeWayMatch(supplierID, invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, audit)
}

// ── GET /api/v1/supplier-portal/scorecard ──
func (h *SupplierPortalHandler) GetSupplierScorecard(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	scorecard, err := h.supplierService(c).CalculateSupplierTier(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, scorecard)
}

// ── GET /api/v1/supplier-portal/api-keys ──
func (h *SupplierPortalHandler) ListAPIKeys(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	keys, err := h.supplierService(c).ListSupplierAPIKeys(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, keys)
}

// ── POST /api/v1/supplier-portal/api-keys ──
func (h *SupplierPortalHandler) CreateAPIKey(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	apiKey, plainKey, err := h.supplierService(c).GenerateSupplierAPIKey(supplierID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"api_key":   apiKey,
		"plain_key": plainKey,
		"message":   "API Key generated successfully. Please copy the key now as it will not be shown again.",
	})
}

// ── DELETE /api/v1/supplier-portal/api-keys/:id ──
func (h *SupplierPortalHandler) RevokeAPIKey(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	keyID := c.Param("id")
	if err := h.supplierService(c).RevokeSupplierAPIKey(supplierID, keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "API key revoked"})
}

// ── GET /api/v1/supplier-portal/webhooks ──
func (h *SupplierPortalHandler) ListWebhooks(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	hooks, err := h.supplierService(c).ListSupplierWebhooks(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hooks)
}

// ── POST /api/v1/supplier-portal/webhooks ──
func (h *SupplierPortalHandler) CreateWebhook(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	var req struct {
		URL    string `json:"url" binding:"required,url"`
		Events string `json:"events" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hook, err := h.supplierService(c).CreateSupplierWebhook(supplierID, req.URL, req.Events)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, hook)
}

// ── DELETE /api/v1/supplier-portal/webhooks/:id ──
func (h *SupplierPortalHandler) DeleteWebhook(c *gin.Context) {
	supplierID := h.resolveSupplierID(c)
	if supplierID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier context required"})
		return
	}
	hookID := c.Param("id")
	if err := h.supplierService(c).DeleteSupplierWebhook(supplierID, hookID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Webhook endpoint deleted"})
}
