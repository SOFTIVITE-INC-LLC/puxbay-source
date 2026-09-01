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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db      *gorm.DB
	service *services.AdminService
}

func NewAdminHandler(db *gorm.DB, authService *services.AuthService) *AdminHandler {
	return &AdminHandler{
		db:      db,
		service: services.NewAdminService(db, authService),
	}
}

type DashboardStatsResponse struct {
	MRR          float64 `json:"mrr"`
	TotalTenants int64   `json:"total_tenants"`
	ActiveUsers  int64   `json:"active_users"`
	TotalOrders  int64   `json:"total_orders"`

	// SaaS Metrics
	ActiveTrials        int64   `json:"active_trials"`
	TrialConversionRate float64 `json:"trial_conversion_rate"`
	ChurnRate           float64 `json:"churn_rate"`
	FailedPaymentsCount int64   `json:"failed_payments_count"`
	RevenueThisMonth    float64 `json:"revenue_this_month"`

	// Real data fields for UI
	PlatformGrowth   []GrowthData   `json:"platform_growth"`
	RecentActivities []ActivityData `json:"recent_activities"`
}

type GrowthData struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type ActivityData struct {
	Icon  string `json:"icon"`
	Title string `json:"title"`
	Time  string `json:"time"`
	Desc  string `json:"desc"`
	Color string `json:"color"`
	Bg    string `json:"bg"`
}

// GetDashboardStats returns aggregated metrics for the admin dashboard
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	var stats DashboardStatsResponse

	h.db.Model(&models.Tenant{}).Count(&stats.TotalTenants)

	// Count non-superuser active users
	h.db.Model(&models.User{}).Where("is_superuser = ?", false).Count(&stats.ActiveUsers)

	// Count total orders across all tenants (ignore if table doesn't exist yet)
	if h.db.Migrator().HasTable(&models.Order{}) {
		h.db.Model(&models.Order{}).Count(&stats.TotalOrders)
	}

	// Calculate MRR from active subscriptions
	stats.MRR = 0
	if h.db.Migrator().HasTable(&models.Subscription{}) && h.db.Migrator().HasTable(&models.PricingPlan{}) {
		h.db.Table("subscriptions").
			Joins("JOIN pricing_plans ON subscriptions.plan_id = pricing_plans.id").
			Where("subscriptions.status = ?", "active").
			Select("COALESCE(SUM(pricing_plans.price_monthly), 0)").
			Scan(&stats.MRR)

		// 1. Active Trials
		h.db.Model(&models.Subscription{}).Where("status = ?", "trialing").Count(&stats.ActiveTrials)

		// 2. Trial Conversion Rate
		var totalUsedTrials int64
		var convertedFromTrial int64
		h.db.Model(&models.Tenant{}).Where("has_used_trial = ?", true).Count(&totalUsedTrials)
		if totalUsedTrials > 0 {
			h.db.Model(&models.Tenant{}).
				Joins("JOIN subscriptions ON subscriptions.tenant_id = tenants.id").
				Where("tenants.has_used_trial = ? AND subscriptions.status = ?", true, "active").
				Count(&convertedFromTrial)
			stats.TrialConversionRate = float64(convertedFromTrial) / float64(totalUsedTrials) * 100.0
		}

		// 3. Churn Rate (Lifetime: total canceled / total ever active)
		var everActive int64
		var currentlyChurned int64
		// Anyone who is active, past_due, or canceled was "ever active" (or converted)
		h.db.Model(&models.Subscription{}).Where("status IN ?", []string{"active", "past_due", "canceled"}).Count(&everActive)
		if everActive > 0 {
			h.db.Model(&models.Subscription{}).Where("status IN ?", []string{"past_due", "canceled"}).Count(&currentlyChurned)
			stats.ChurnRate = float64(currentlyChurned) / float64(everActive) * 100.0
		}

		// 4. Failed Payments Count
		h.db.Model(&models.BillingPayment{}).Where("status IN ?", []string{"failed", "declined"}).Count(&stats.FailedPaymentsCount)

		// 5. Revenue This Month
		startOfMonth := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -time.Now().Day()+1)
		h.db.Model(&models.BillingPayment{}).
			Where("status IN ? AND (created_at >= ? OR date >= ?)", []string{"succeeded", "successful", "paid"}, startOfMonth, startOfMonth).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&stats.RevenueThisMonth)
	}

	// 1. Fetch System Activities (last 5 audit logs)
	var logs []models.AuditLog
	if h.db.Migrator().HasTable(&models.AuditLog{}) {
		h.db.Order("created_at desc").Limit(5).Find(&logs)
	}

	for _, log := range logs {
		color := "text-slate-500"
		bg := "bg-slate-100"
		icon := "info"

		switch log.Action {
		case "create":
			color = "text-emerald-500"
			bg = "bg-emerald-100"
			icon = "add_circle"
		case "delete":
			color = "text-red-500"
			bg = "bg-red-100"
			icon = "delete"
		case "login":
			color = "text-indigo-500"
			bg = "bg-indigo-100"
			icon = "login"
		}

		desc := "System action"
		if log.ObjectID != nil {
			desc = "ID: " + *log.ObjectID
		}

		stats.RecentActivities = append(stats.RecentActivities, ActivityData{
			Icon:  icon,
			Title: log.ModelName + " " + log.Action,
			Time:  log.CreatedAt.Format("02 Jan 15:04"),
			Desc:  desc,
			Color: color,
			Bg:    bg,
		})
	}

	// 2. Fetch Platform Growth (Tenant signups over last 6 months)
	sixMonthsAgo := time.Now().AddDate(0, -5, 0)
	var tenants []models.Tenant
	h.db.Select("created_on").Where("created_on >= ?", sixMonthsAgo).Find(&tenants)

	// Group by month in Go to remain DB agnostic
	countsByMonth := make(map[string]int)
	months := make([]string, 6)

	for i := 5; i >= 0; i-- {
		m := time.Now().AddDate(0, -i, 0).Format("Jan")
		months[5-i] = m
		countsByMonth[m] = 0 // initialize
	}

	for _, t := range tenants {
		m := t.CreatedOn.Format("Jan")
		if _, exists := countsByMonth[m]; exists {
			countsByMonth[m]++
		}
	}

	// Calculate cumulative growth or just monthly signups
	// We'll show cumulative active tenants over the 6 months for a nice upward chart
	var baseCount int64
	h.db.Model(&models.Tenant{}).Where("created_on < ?", sixMonthsAgo).Count(&baseCount)

	runningTotal := int(baseCount)
	stats.PlatformGrowth = make([]GrowthData, 6)

	for i, m := range months {
		runningTotal += countsByMonth[m]
		stats.PlatformGrowth[i] = GrowthData{
			Label: m,
			Value: runningTotal,
		}
	}

	c.JSON(http.StatusOK, stats)
}

// SystemHealth checks system status
func (h *AdminHandler) SystemHealth(c *gin.Context) {
	start := time.Now()
	sqlDB, err := h.db.DB()

	status := "healthy"
	if err != nil || sqlDB.Ping() != nil {
		status = "degraded"
	}

	latencyMs := time.Since(start).Milliseconds()

	c.JSON(http.StatusOK, gin.H{
		"status":     status,
		"version":    "1.0.0",
		"latency_ms": latencyMs,
	})
}

// ListTenants returns a list of tenants (Superadmin only)
func (h *AdminHandler) ListTenants(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	tenants, stats, err := h.service.ListTenants(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tenants,
		"stats": stats,
	})
}

func (h *AdminHandler) ListDomains(c *gin.Context) {
	tenantIDCtx, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID := tenantIDCtx.(uuid.UUID).String()
	domains, err := h.service.ListDomains(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch domains"})
		return
	}
	c.JSON(http.StatusOK, domains)
}

type CreateDomainRequest struct {
	Domain string `json:"domain" binding:"required"`
}

func (h *AdminHandler) CreateDomain(c *gin.Context) {
	tenantIDCtx, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID := tenantIDCtx.(uuid.UUID).String()
	var req CreateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain, err := h.service.CreateDomain(tenantID, req.Domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create domain"})
		return
	}
	c.JSON(http.StatusCreated, domain)
}

func (h *AdminHandler) DeleteDomain(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteDomain(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete domain"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *AdminHandler) VerifyDomain(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.VerifyDomain(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify domain"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "verified"})
}

func (h *AdminHandler) SetPrimaryDomain(c *gin.Context) {
	tenantIDCtx, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID := tenantIDCtx.(uuid.UUID).String()
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.SetPrimaryDomain(tenantID, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set primary domain"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "set as primary"})
}

func (h *AdminHandler) SuspendTenant(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	if err := h.service.SuspendTenant(tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to suspend tenant"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "suspended"})
}

func (h *AdminHandler) ImpersonateTenant(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	adminIDCtx, _ := c.Get(middleware.ContextKeyUserID)
	adminID := adminIDCtx.(uuid.UUID)

	token, err := h.service.ImpersonateTenant(tenantID, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to impersonate tenant"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AdminHandler) ListPlans(c *gin.Context) {
	plans, err := h.service.ListPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plans"})
		return
	}
	c.JSON(http.StatusOK, plans)
}

func (h *AdminHandler) CreatePlan(c *gin.Context) {
	var plan models.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreatePlan(&plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create plan"})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h *AdminHandler) ListBroadcasts(c *gin.Context) {
	broadcasts, err := h.service.ListBroadcasts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch broadcasts"})
		return
	}
	c.JSON(http.StatusOK, broadcasts)
}

func (h *AdminHandler) CreateBroadcast(c *gin.Context) {
	var broadcast models.Broadcast
	if err := c.ShouldBindJSON(&broadcast); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminIDCtx, _ := c.Get(middleware.ContextKeyUserID)
	broadcast.CreatedBy = adminIDCtx.(uuid.UUID)

	if err := h.service.CreateBroadcast(&broadcast); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create broadcast"})
		return
	}
	c.JSON(http.StatusCreated, broadcast)
}

func (h *AdminHandler) UpdateFeatureFlags(c *gin.Context) {
	var flags map[string]interface{}
	if err := c.ShouldBindJSON(&flags); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateFeatureFlags(flags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature flags"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// PricingPlan endpoints

func (h *AdminHandler) ListPricingPlans(c *gin.Context) {
	plans, err := h.service.ListPricingPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pricing plans"})
		return
	}
	c.JSON(http.StatusOK, plans)
}

func (h *AdminHandler) CreatePricingPlan(c *gin.Context) {
	var plan models.PricingPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreatePricingPlan(&plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pricing plan"})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h *AdminHandler) UpdatePricingPlan(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pricing plan ID"})
		return
	}

	var plan models.PricingPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan.ID = id

	if err := h.service.UpdatePricingPlan(&plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pricing plan"})
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *AdminHandler) DeletePricingPlan(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pricing plan ID"})
		return
	}

	if err := h.service.DeletePricingPlan(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pricing plan"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Subscriptions
func (h *AdminHandler) ListSubscriptions(c *gin.Context) {
	subs, stats, err := h.service.ListSubscriptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  subs,
		"stats": stats,
	})
}

func (h *AdminHandler) OverrideSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.OverrideSubscription(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to override subscription"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "overridden"})
}

// PromoCodes
func (h *AdminHandler) ListPromoCodes(c *gin.Context) {
	codes, stats, err := h.service.ListPromoCodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch promo codes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  codes,
		"stats": stats,
	})
}

func (h *AdminHandler) CreatePromoCode(c *gin.Context) {
	var code models.PromoCode
	if err := c.ShouldBindJSON(&code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.CreatePromoCode(&code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create promo code"})
		return
	}
	c.JSON(http.StatusCreated, code)
}

func (h *AdminHandler) TogglePromoCode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid promo code ID"})
		return
	}
	if err := h.service.TogglePromoCode(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle promo code"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "toggled"})
}

// BillingPayments
func (h *AdminHandler) ListBillingPayments(c *gin.Context) {
	payments, stats, err := h.service.ListBillingPayments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  payments,
		"stats": stats,
	})
}

// AuditLogs
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	logs, stats, err := h.service.ListAuditLogs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"stats": stats,
	})
}

// FAQs
func (h *AdminHandler) ListFAQs(c *gin.Context) {
	faqs, err := h.service.ListFAQs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FAQs"})
		return
	}
	c.JSON(http.StatusOK, faqs)
}

func (h *AdminHandler) CreateFAQ(c *gin.Context) {
	var faq models.FAQ
	if err := c.ShouldBindJSON(&faq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.CreateFAQ(&faq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create FAQ"})
		return
	}
	c.JSON(http.StatusCreated, faq)
}

func (h *AdminHandler) ToggleFAQ(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid FAQ ID"})
		return
	}
	if err := h.service.ToggleFAQ(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle FAQ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "toggled"})
}

// App Marketplace
func (h *AdminHandler) ListExternalSystems(c *gin.Context) {
	systems, err := h.service.ListExternalSystems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch apps"})
		return
	}
	c.JSON(http.StatusOK, systems)
}

func (h *AdminHandler) ToggleExternalSystem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid App ID"})
		return
	}
	field := c.Query("field") // active or public
	if err := h.service.ToggleExternalSystem(id, field); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle app"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "toggled"})
}

// System Backups
func (h *AdminHandler) ListDatabaseBackups(c *gin.Context) {
	backups, err := h.service.ListDatabaseBackups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch backups"})
		return
	}
	c.JSON(http.StatusOK, backups)
}

func (h *AdminHandler) TriggerDatabaseBackup(c *gin.Context) {
	backup, err := h.service.TriggerDatabaseBackup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, backup)
}

func (h *AdminHandler) DownloadDatabaseBackup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}

	backup, err := h.service.GetDatabaseBackup(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}

	// Serve the file
	c.FileAttachment(backup.FilePath, backup.Filename)
}

// ───── Phase 1: Tenant Command Center ─────

// GetTenantDetail returns full details for a single tenant.
func (h *AdminHandler) GetTenantDetail(c *gin.Context) {
	id := c.Param("id")
	tenant, err := h.service.GetTenantDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}
	c.JSON(http.StatusOK, tenant)
}

// SearchTenants returns filtered/searched tenants.
func (h *AdminHandler) SearchTenants(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")
	page := 1
	tenants, total, err := h.service.SearchTenants(search, status, 50, (page-1)*50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search tenants"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tenants, "total": total})
}

// CreateTenant manually creates a new tenant.
func (h *AdminHandler) CreateTenant(c *gin.Context) {
	var t struct {
		Name      string `json:"name" binding:"required"`
		Subdomain string `json:"subdomain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenant := &models.Tenant{
		Name:      t.Name,
		Subdomain: t.Subdomain,
		Status:    "active",
	}
	if err := h.service.CreateTenant(tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tenant)
}

// UpdateTenantNotes saves an internal note on a tenant.
func (h *AdminHandler) UpdateTenantNotes(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateTenantNotes(id, body.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ───── Phase 2: Billing & Revenue Intelligence ─────

// ListUpcomingRenewals returns subscriptions renewing in the next 7 days.
func (h *AdminHandler) ListUpcomingRenewals(c *gin.Context) {
	subs, err := h.service.ListUpcomingRenewals(7)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch upcoming renewals"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": subs})
}

// ListFailedPayments returns failed payment records.
func (h *AdminHandler) ListFailedPayments(c *gin.Context) {
	payments, err := h.service.ListFailedPayments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch failed payments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": payments})
}

// ───── Phase 3: Growth ─────

// ListReferrals returns all referral records.
func (h *AdminHandler) ListReferrals(c *gin.Context) {
	rewards, err := h.service.ListReferrals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch referrals"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rewards})
}

// ───── Phase 5: Webhook Event Log ─────

// ListWebhookEvents returns the webhook delivery log.
func (h *AdminHandler) ListWebhookEvents(c *gin.Context) {
	events, err := h.service.ListWebhookEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch webhook events"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events})
}

// ───── Phase 4: Security & Access Control ─────

func (h *AdminHandler) ListAdminRoles(c *gin.Context) {
	roles, err := h.service.ListAdminRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *AdminHandler) CreateAdminRole(c *gin.Context) {
	var role models.AdminRole
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.CreateAdminRole(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *AdminHandler) UpdateAdminRole(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Permissions string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateAdminRole(id, body.Permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminHandler) ListIPAllowlist(c *gin.Context) {
	ips, err := h.service.ListIPAllowlist()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch IPs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ips})
}

func (h *AdminHandler) AddIPToAllowlist(c *gin.Context) {
	var ip models.IPAllowlist
	if err := c.ShouldBindJSON(&ip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.AddIPToAllowlist(&ip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add IP"})
		return
	}
	c.JSON(http.StatusCreated, ip)
}

func (h *AdminHandler) RemoveIPFromAllowlist(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.RemoveIPFromAllowlist(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove IP"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ───── Phase 5: Operations & Integrations (API Keys) ─────

func (h *AdminHandler) ListMasterAPIKeys(c *gin.Context) {
	keys, err := h.service.ListMasterAPIKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch API keys"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": keys})
}

func (h *AdminHandler) CreateMasterAPIKey(c *gin.Context) {
	var key models.MasterAPIKey
	if err := c.ShouldBindJSON(&key); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a random key
	key.Key = "pux_" + uuid.New().String()

	if err := h.service.CreateMasterAPIKey(&key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}
	c.JSON(http.StatusCreated, key)
}

func (h *AdminHandler) RevokeMasterAPIKey(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.RevokeMasterAPIKey(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke API key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ───── Domains ─────

func (h *AdminHandler) SearchAllDomains(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	search := c.Query("search")
	status := c.Query("status")

	domains, total, err := h.service.SearchAllDomains(search, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch domains"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": domains, "total": total})
}

func (h *AdminHandler) BulkVerifyDomains(c *gin.Context) {
	var input struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.BulkVerifyDomains(input.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk verify domains"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminHandler) GetDomainDiagnostics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID"})
		return
	}

	diagnostics, err := h.service.GetDomainDiagnostics(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run diagnostics"})
		return
	}
	c.JSON(http.StatusOK, diagnostics)
}

// ListLegalDocuments returns all legal documents (admin only).
func (h *AdminHandler) ListLegalDocuments(c *gin.Context) {
	var docs []models.LegalDocument
	if err := h.db.Find(&docs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch legal documents"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"documents": docs})
}

// UpsertLegalDocument creates or updates a legal document by type (terms | privacy | cookie).
func (h *AdminHandler) UpsertLegalDocument(c *gin.Context) {
	docType := c.Param("type")
	validTypes := map[string]bool{"terms": true, "privacy": true, "cookie": true}
	if !validTypes[docType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document type. Must be one of: terms, privacy, cookie"})
		return
	}

	var input struct {
		Title         string  `json:"title"`
		Content       string  `json:"content" binding:"required"`
		Version       string  `json:"version"`
		EffectiveDate *string `json:"effective_date"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var doc models.LegalDocument
	result := h.db.Where("type = ?", docType).First(&doc)

	if result.Error != nil {
		// Create new
		doc = models.LegalDocument{
			Type:    docType,
			Title:   input.Title,
			Content: input.Content,
			Version: input.Version,
		}
		if input.Title == "" {
			titles := map[string]string{
				"terms":   "Terms of Service",
				"privacy": "Privacy Policy",
				"cookie":  "Cookie Policy",
			}
			doc.Title = titles[docType]
		}
		if input.Version == "" {
			doc.Version = "1.0"
		}
		if input.EffectiveDate != nil {
			t, err := time.Parse("2006-01-02", *input.EffectiveDate)
			if err == nil {
				doc.EffectiveDate = &t
			}
		}
		if err := h.db.Create(&doc).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create legal document"})
			return
		}
	} else {
		// Update existing
		updates := map[string]interface{}{
			"content": input.Content,
		}
		if input.Title != "" {
			updates["title"] = input.Title
		}
		if input.Version != "" {
			updates["version"] = input.Version
		}
		if input.EffectiveDate != nil {
			t, err := time.Parse("2006-01-02", *input.EffectiveDate)
			if err == nil {
				updates["effective_date"] = t
			}
		}
		if err := h.db.Model(&doc).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update legal document"})
			return
		}
	}

	// Re-fetch to return full updated record
	h.db.Where("type = ?", docType).First(&doc)
	c.JSON(http.StatusOK, doc)
}

// GetPublicLegalDocument returns a legal document by type — no authentication required.
func (h *AdminHandler) GetPublicLegalDocument(c *gin.Context) {
	docType := c.Param("type")
	validTypes := map[string]bool{"terms": true, "privacy": true, "cookie": true}
	if !validTypes[docType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document type"})
		return
	}

	var doc models.LegalDocument
	if err := h.db.Where("type = ?", docType).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	c.JSON(http.StatusOK, doc)
}

// ListAdminBlogPosts returns all blog posts for the admin panel.
func (h *AdminHandler) ListAdminBlogPosts(c *gin.Context) {
	var posts []models.BlogPost
	if err := h.db.Order("created_at desc").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blog posts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// GetAdminBlogPost returns a single blog post for the admin panel by ID.
func (h *AdminHandler) GetAdminBlogPost(c *gin.Context) {
	id := c.Param("id")
	var post models.BlogPost
	if err := h.db.First(&post, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
		return
	}
	c.JSON(http.StatusOK, post)
}

// CreateBlogPost creates a new blog post.
func (h *AdminHandler) CreateBlogPost(c *gin.Context) {
	var post models.BlogPost
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if post.Status == "published" && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
	}

	if err := h.db.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create blog post"})
		return
	}
	c.JSON(http.StatusCreated, post)
}

// UpdateBlogPost updates an existing blog post.
func (h *AdminHandler) UpdateBlogPost(c *gin.Context) {
	id := c.Param("id")
	var post models.BlogPost
	if err := h.db.First(&post, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
		return
	}

	var input models.BlogPost
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle publish date
	if input.Status == "published" && post.Status != "published" && post.PublishedAt == nil {
		now := time.Now()
		input.PublishedAt = &now
	}

	if err := h.db.Model(&post).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update blog post"})
		return
	}
	
	h.db.First(&post, "id = ?", id)
	c.JSON(http.StatusOK, post)
}

// DeleteBlogPost deletes a blog post.
func (h *AdminHandler) DeleteBlogPost(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&models.BlogPost{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete blog post"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Blog post deleted"})
}

// ListPublicBlogPosts returns published blog posts for the public frontend.
func (h *AdminHandler) ListPublicBlogPosts(c *gin.Context) {
	var posts []models.BlogPost
	if err := h.db.Where("status = ?", "published").Order("published_at desc").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blog posts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// GetPublicBlogPost returns a single published blog post by slug.
func (h *AdminHandler) GetPublicBlogPost(c *gin.Context) {
	slug := c.Param("slug")
	var post models.BlogPost
	if err := h.db.Where("slug = ? AND status = ?", slug, "published").First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
		return
	}
	c.JSON(http.StatusOK, post)
}

// ListTelemetryLogs returns a paginated list of telemetry logs (OpenTelemetry traces).
func (h *AdminHandler) ListTelemetryLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var logs []models.TelemetryLog
	var total int64

	// Count total records
	if err := h.db.Model(&models.TelemetryLog{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count telemetry logs"})
		return
	}

	// Fetch paginated logs, ordered by newest first
	if err := h.db.Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch telemetry logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
	})
}

// ───── Admin Users Management ─────

func (h *AdminHandler) ListAdminUsers(c *gin.Context) {
	// We want to return users with their AdminUser information
	type AdminUserResponse struct {
		models.User
		AdminRoleID *uuid.UUID `json:"admin_role_id"`
		AdminRole   *string    `json:"admin_role_name,omitempty"`
		Permissions string     `json:"permissions"`
	}

	var users []models.User
	if err := h.db.Where("is_superuser = ?", true).Or("id IN (SELECT user_id FROM admin_users)").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch admin users"})
		return
	}

	// Fetch roles
	var adminUsers []models.AdminUser
	h.db.Preload("Role").Find(&adminUsers)
	
	roleMap := make(map[uuid.UUID]models.AdminUser)
	for _, au := range adminUsers {
		roleMap[au.UserID] = au
	}

	response := make([]AdminUserResponse, len(users))
	for i, u := range users {
		res := AdminUserResponse{User: u, Permissions: "[]"}
		if au, ok := roleMap[u.ID]; ok {
			res.AdminRoleID = au.AdminRoleID
			if au.Role != nil {
				res.AdminRole = &au.Role.Name
			}
			if au.Permissions != "" {
				res.Permissions = au.Permissions
			}
		}
		response[i] = res
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *AdminHandler) CreateAdminUser(c *gin.Context) {
	var req struct {
		FirstName   string     `json:"first_name" binding:"required"`
		LastName    string     `json:"last_name" binding:"required"`
		Email       string     `json:"email" binding:"required,email"`
		Username    string     `json:"username" binding:"required"`
		Password    string     `json:"password" binding:"required"`
		AdminRoleID *uuid.UUID `json:"admin_role_id"`
		IsSuperuser bool       `json:"is_superuser"`
		Permissions string     `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Username:    req.Username,
		Password:    string(hashedPassword),
		IsSuperuser: req.IsSuperuser,
	}

	// Start a transaction
	tx := h.db.Begin()

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	if req.AdminRoleID != nil || req.Permissions != "" {
		adminUser := models.AdminUser{
			UserID:      user.ID,
			AdminRoleID: req.AdminRoleID,
			Permissions: req.Permissions,
		}
		if err := tx.Create(&adminUser).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign admin role/permissions"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusCreated, user)
}

func (h *AdminHandler) UpdateAdminUserRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		AdminRoleID *uuid.UUID `json:"admin_role_id"`
		IsSuperuser bool       `json:"is_superuser"`
		Permissions string     `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := h.db.Begin()

	if err := tx.Model(&models.User{}).Where("id = ?", id).Update("is_superuser", req.IsSuperuser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	// Update AdminUser record
	if req.AdminRoleID != nil || req.Permissions != "" {
		var au models.AdminUser
		if err := tx.Where("user_id = ?", id).First(&au).Error; err != nil {
			// Doesn't exist, create it
			au = models.AdminUser{UserID: id, AdminRoleID: req.AdminRoleID, Permissions: req.Permissions}
			if err := tx.Create(&au).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign admin role/permissions"})
				return
			}
		} else {
			// Update existing
			updates := map[string]interface{}{
				"admin_role_id": req.AdminRoleID,
				"permissions":   req.Permissions,
			}
			if err := tx.Model(&au).Updates(updates).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin role/permissions"})
				return
			}
		}
	} else {
		// Remove role
		tx.Where("user_id = ?", id).Delete(&models.AdminUser{})
	}

	tx.Commit()

	// Audit log
	h.writeAdminAudit(c, "update_role", "AdminUser", id.String(), nil)

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// ───── writeAdminAudit ─────

func (h *AdminHandler) writeAdminAudit(c *gin.Context, action, modelName, objectID string, changes map[string]interface{}) {
	userIDVal, _ := c.Get(middleware.ContextKeyUserID)
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	log := models.AuditLog{
		Action:    action,
		ModelName: modelName,
	}
	if objectID != "" {
		log.ObjectID = &objectID
	}
	log.IPAddress = &ip
	if ua != "" {
		log.UserAgent = &ua
	}
	if uid, ok := userIDVal.(uuid.UUID); ok {
		log.UserID = &uid
	}
	h.db.Create(&log)
}

// ───── DeleteAdminUser ─────

func (h *AdminHandler) DeleteAdminUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	tx := h.db.Begin()

	// Remove from admin_users table
	tx.Where("user_id = ?", id).Delete(&models.AdminUser{})

	// Revoke superuser flag and bump token_version to invalidate sessions
	if err := tx.Model(&models.User{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_superuser": false, "token_version": h.db.Raw("token_version + 1")}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke admin access"})
		return
	}

	tx.Commit()

	h.writeAdminAudit(c, "delete", "AdminUser", id.String(), nil)
	c.JSON(http.StatusOK, gin.H{"status": "admin access revoked"})
}

// ───── RetryWebhookEvent ─────

func (h *AdminHandler) RetryWebhookEvent(c *gin.Context) {
	eventID := c.Param("id")

	var event models.WebhookEvent
	if err := h.db.First(&event, "id = ?", eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook event not found"})
		return
	}

	// Reset attempts counter and status so the worker will pick it up again
	if err := h.db.Model(&event).Updates(map[string]interface{}{
		"status":   "pending",
		"attempts": 0,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to requeue webhook event"})
		return
	}

	h.writeAdminAudit(c, "retry", "WebhookEvent", eventID, nil)
	c.JSON(http.StatusOK, gin.H{"status": "queued for retry"})
}

// ───── Gift Cards (Admin) ─────

func (h *AdminHandler) ListAllGiftCards(c *gin.Context) {
	var cards []models.GiftCard
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 50
	offset := (page - 1) * limit

	var total int64
	h.db.Model(&models.GiftCard{}).Count(&total)
	h.db.Order("created_at desc").Limit(limit).Offset(offset).Find(&cards)

	c.JSON(http.StatusOK, gin.H{"data": cards, "total": total})
}

func (h *AdminHandler) CreateGiftCardAdmin(c *gin.Context) {
	var req struct {
		InitialBalance float64 `json:"initial_balance" binding:"required"`
		CustomCode     string  `json:"custom_code"`
		ExpiresAt      string  `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card := models.GiftCard{
		InitialBalance:  req.InitialBalance,
		CurrentBalance:  req.InitialBalance,
		IsActive:        true,
		Status:          "active",
	}

	if req.CustomCode != "" {
		card.Code = req.CustomCode
	} else {
		card.Code = "GC-" + uuid.New().String()[:8]
	}

	if req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiresAt)
		if err == nil {
			card.ExpiresAt = &t
		}
	}

	if err := h.db.Create(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gift card"})
		return
	}

	h.writeAdminAudit(c, "create", "GiftCard", card.ID.String(), nil)
	c.JSON(http.StatusCreated, card)
}

func (h *AdminHandler) DisableGiftCardAdmin(c *gin.Context) {
	var card models.GiftCard
	if err := h.db.First(&card, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gift card not found"})
		return
	}

	card.Status = "disabled"
	card.IsActive = false

	if err := h.db.Save(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable gift card"})
		return
	}

	h.writeAdminAudit(c, "disable", "GiftCard", card.ID.String(), nil)
	c.JSON(http.StatusOK, card)
}

func (h *AdminHandler) EnableGiftCardAdmin(c *gin.Context) {
	var card models.GiftCard
	if err := h.db.First(&card, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gift card not found"})
		return
	}

	card.Status = "active"
	card.IsActive = true

	if err := h.db.Save(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable gift card"})
		return
	}

	h.writeAdminAudit(c, "enable", "GiftCard", card.ID.String(), nil)
	c.JSON(http.StatusOK, card)
}
