package services

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type AdminService struct {
	db          *gorm.DB
	authService *AuthService
}

func NewAdminService(db *gorm.DB, authService *AuthService) *AdminService {
	return &AdminService{db: db, authService: authService}
}

type TenantStats struct {
	Total     int64   `json:"total"`
	Active    int64   `json:"active"`
	Suspended int64   `json:"suspended"`
	Trialing  int64   `json:"trialing"`
	Revenue   float64 `json:"revenue"`
}

func (s *AdminService) ListTenants(limit, offset int) ([]models.Tenant, TenantStats, error) {
	var tenants []models.Tenant
	var stats TenantStats

	query := s.db.Model(&models.Tenant{})
	query.Count(&stats.Total)

	// Query breakdowns using the new Status column and Subscription table
	s.db.Model(&models.Tenant{}).Where("status = ? OR status IS NULL", "active").Count(&stats.Active)
	s.db.Model(&models.Tenant{}).Where("status = ?", "suspended").Count(&stats.Suspended)

	if s.db.Migrator().HasTable(&models.Subscription{}) {
		s.db.Model(&models.Subscription{}).Where("status = ?", "trialing").Count(&stats.Trialing)
	} else {
		stats.Trialing = 0
	}

	// Sum revenue from TenantMetrics
	if s.db.Migrator().HasTable(&models.TenantMetrics{}) {
		s.db.Model(&models.TenantMetrics{}).Select("COALESCE(SUM(total_revenue), 0)").Scan(&stats.Revenue)
	}

	if err := query.Preload("Subscription").Order("created_on desc").Offset(offset).Limit(limit).Find(&tenants).Error; err != nil {
		return nil, stats, err
	}

	// Populate Plan from PricingPlan manually
	for i := range tenants {
		t := &tenants[i]
		if t.Subscription != nil && t.Subscription.PlanID != nil && t.Subscription.Plan == nil {
			var pricingPlan models.PricingPlan
			if err := s.db.Where("id = ?", *t.Subscription.PlanID).First(&pricingPlan).Error; err == nil {
				t.Subscription.Plan = &models.Plan{
					Name:  pricingPlan.Name,
					Price: pricingPlan.PriceMonthly,
				}
			}
		}
	}

	return tenants, stats, nil
}

func (s *AdminService) SuspendTenant(tenantID uuid.UUID) error {
	return s.db.Model(&models.Tenant{}).Where("id = ?", tenantID).Update("status", "suspended").Error
}

func (s *AdminService) ImpersonateTenant(tenantID, adminID uuid.UUID) (string, error) {
	// Verify tenant exists
	var tenant models.Tenant
	if err := s.db.First(&tenant, "id = ?", tenantID).Error; err != nil {
		return "", fmt.Errorf("tenant not found: %w", err)
	}

	// Generate tokens for the admin but scoped to the target tenant
	tokenPair, err := s.authService.GenerateTokenPair(adminID, tenantID, nil, "superadmin", nil, 1)
	if err != nil {
		return "", err
	}

	auditLog := models.CrossTenantAuditLog{
		UserID:           &adminID,
		AccessedTenantID: &tenantID,
		ActionType:       "IMPERSONATE_LOGIN",
	}
	s.db.Create(&auditLog)

	return tokenPair.AccessToken, nil
}

func (s *AdminService) ListPlans() ([]models.Plan, error) {
	var plans []models.Plan
	if err := s.db.Order("price asc").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (s *AdminService) CreatePlan(plan *models.Plan) error {
	return s.db.Create(plan).Error
}

func (s *AdminService) ListBroadcasts() ([]models.Broadcast, error) {
	var broadcasts []models.Broadcast
	if err := s.db.Order("created_at desc").Limit(50).Find(&broadcasts).Error; err != nil {
		return nil, err
	}
	return broadcasts, nil
}

func (s *AdminService) CreateBroadcast(broadcast *models.Broadcast) error {
	return s.db.Create(broadcast).Error
}

// Global Feature flags (Stored in DB for persistence)
func (s *AdminService) UpdateFeatureFlags(flags map[string]interface{}) error {
	// Automatically migrate the FeatureFlag table if it doesn't exist yet
	if err := s.db.AutoMigrate(&models.FeatureFlag{}); err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for key, val := range flags {
			var boolVal bool
			if b, ok := val.(bool); ok {
				boolVal = b
			} else {
				continue // Skip non-boolean flags for now
			}

			flag := models.FeatureFlag{Key: key, Value: boolVal}
			// Upsert the flag
			if err := tx.Save(&flag).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AdminService) ListDomains(tenantID string) ([]models.Domain, error) {
	var domains []models.Domain
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&domains).Error; err != nil {
		return nil, err
	}
	return domains, nil
}

func (s *AdminService) CreateDomain(tenantID, domainStr string) (*models.Domain, error) {
	domain := models.Domain{
		TenantID:   uuid.MustParse(tenantID),
		Domain:     domainStr,
		IsPrimary:  false,
		IsVerified: false,
	}
	if err := s.db.Create(&domain).Error; err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *AdminService) DeleteDomain(id uint) error {
	return s.db.Delete(&models.Domain{}, id).Error
}

func (s *AdminService) VerifyDomain(id uint) error {
	return s.db.Model(&models.Domain{}).Where("id = ?", id).Update("is_verified", true).Error
}

func (s *AdminService) SetPrimaryDomain(tenantID string, domainID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Domain{}).Where("tenant_id = ?", tenantID).Update("is_primary", false).Error; err != nil {
			return err
		}
		return tx.Model(&models.Domain{}).Where("id = ? AND tenant_id = ?", domainID, tenantID).Update("is_primary", true).Error
	})
}

// PricingPlan Methods

func (s *AdminService) ListPricingPlans() ([]models.PricingPlan, error) {
	var plans []models.PricingPlan
	if !s.db.Migrator().HasTable(&models.PricingPlan{}) {
		return plans, nil
	}
	if err := s.db.Preload("Features", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_index ASC")
	}).Order("order_index ASC").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (s *AdminService) CreatePricingPlan(plan *models.PricingPlan) error {
	return s.db.Create(plan).Error
}

func (s *AdminService) UpdatePricingPlan(plan *models.PricingPlan) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(plan).Error; err != nil {
			return err
		}

		if err := tx.Where("plan_id = ?", plan.ID).Delete(&models.PlanFeature{}).Error; err != nil {
			return err
		}

		if len(plan.Features) > 0 {
			for i := range plan.Features {
				plan.Features[i].PlanID = plan.ID
				plan.Features[i].ID = uuid.Nil
			}
			if err := tx.Create(&plan.Features).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AdminService) DeletePricingPlan(id uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("plan_id = ?", id).Delete(&models.PlanFeature{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.PricingPlan{}, "id = ?", id).Error
	})
}

// Subscriptions
func (s *AdminService) ListSubscriptions() ([]models.Subscription, map[string]interface{}, error) {
	var subs []models.Subscription
	stats := map[string]interface{}{
		"total":    0,
		"active":   0,
		"trialing": 0,
		"past_due": 0,
		"mrr":      0.0,
	}

	if !s.db.Migrator().HasTable(&models.Subscription{}) {
		return subs, stats, nil
	}
	err := s.db.Preload("Tenant").Find(&subs).Error

	// Calculate stats and populate PricingPlans
	for i := range subs {
		sub := &subs[i]
		stats["total"] = stats["total"].(int) + 1

		if sub.PlanID != nil && sub.Plan == nil {
			var pricingPlan models.PricingPlan
			if err := s.db.Where("id = ?", *sub.PlanID).First(&pricingPlan).Error; err == nil {
				sub.Plan = &models.Plan{
					Name:  pricingPlan.Name,
					Price: pricingPlan.PriceMonthly,
				}
			}
		}

		switch sub.Status {
		case "active":
			stats["active"] = stats["active"].(int) + 1
			if sub.Plan != nil {
				stats["mrr"] = stats["mrr"].(float64) + sub.Plan.Price
			}
		case "trialing":
			stats["trialing"] = stats["trialing"].(int) + 1
		case "past_due":
			stats["past_due"] = stats["past_due"].(int) + 1
		}
	}

	return subs, stats, err
}

func (s *AdminService) OverrideSubscription(subID uuid.UUID, status string) error {
	return s.db.Model(&models.Subscription{}).Where("id = ?", subID).Update("status", status).Error
}

// PromoCodes
func (s *AdminService) ListPromoCodes() ([]models.PromoCode, map[string]interface{}, error) {
	var codes []models.PromoCode
	stats := map[string]interface{}{
		"active_codes":      0,
		"total_redemptions": 0,
		"top_code":          "-",
	}

	if !s.db.Migrator().HasTable(&models.PromoCode{}) {
		return codes, stats, nil
	}

	err := s.db.Find(&codes).Error

	// Calculate stats
	maxUses := uint(0)
	for _, c := range codes {
		if c.IsActive {
			stats["active_codes"] = stats["active_codes"].(int) + 1
		}
		stats["total_redemptions"] = stats["total_redemptions"].(int) + int(c.CurrentUses)

		if c.CurrentUses > maxUses {
			maxUses = c.CurrentUses
			stats["top_code"] = c.Code
		}
	}

	return codes, stats, err
}

func (s *AdminService) CreatePromoCode(code *models.PromoCode) error {
	return s.db.Create(code).Error
}

func (s *AdminService) TogglePromoCode(id uuid.UUID) error {
	var code models.PromoCode
	if err := s.db.First(&code, "id = ?", id).Error; err != nil {
		return err
	}
	return s.db.Model(&code).Update("is_active", !code.IsActive).Error
}

type PaymentFilterParams struct {
	Search           string `form:"search"`
	Status           string `form:"status"`
	PaymentType      string `form:"payment_type"`
	SubaccountRouted string `form:"subaccount_routed"` // "all", "true", "false"
	DisputeStatus    string `form:"dispute_status"`
	TenantID         string `form:"tenant_id"`
	FromDate         string `form:"from_date"`
	ToDate           string `form:"to_date"`
	Page             int    `form:"page,default=1"`
	Limit            int    `form:"limit,default=50"`
}

type PaymentLogInput struct {
	TenantID           *uuid.UUID `json:"tenant_id"`
	TenantName         string     `json:"tenant_name"`
	PaymentType        string     `json:"payment_type"`
	Reference          string     `json:"reference"`
	OrderID            *uuid.UUID `json:"order_id"`
	OrderNumber        string     `json:"order_number"`
	Amount             float64    `json:"amount"`
	Currency           string     `json:"currency"`
	PaymentMethod      string     `json:"payment_method"`
	Gateway            string     `json:"gateway"`
	SubaccountCode     *string    `json:"subaccount_code"`
	IsSubaccountRouted bool       `json:"is_subaccount_routed"`
	SubaccountShare    float64    `json:"subaccount_share"`
	PlatformFee        float64    `json:"platform_fee"`
	CustomerName       string     `json:"customer_name"`
	CustomerEmail      string     `json:"customer_email"`
	CustomerPhone      string     `json:"customer_phone"`
	Status             string     `json:"status"`
	DisputeStatus      string     `json:"dispute_status"`
	Notes              string     `json:"notes"`
}

type PaymentLogUpdateInput struct {
	Status        *string `json:"status"`
	DisputeStatus *string `json:"dispute_status"`
	Notes         *string `json:"notes"`
}

// ListPaymentLogs retrieves paginated, searchable payment logs with rich metrics
func (s *AdminService) ListPaymentLogs(params PaymentFilterParams) ([]models.PaymentLog, map[string]interface{}, int64, error) {
	_ = s.db.AutoMigrate(&models.PaymentLog{})

	query := s.db.Table("public.payment_logs").Preload("Tenant")

	if params.Search != "" {
		sTerm := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("reference ILIKE ? OR order_number ILIKE ? OR tenant_name ILIKE ? OR customer_name ILIKE ? OR customer_phone ILIKE ? OR customer_email ILIKE ? OR subaccount_code ILIKE ?",
			sTerm, sTerm, sTerm, sTerm, sTerm, sTerm, sTerm)
	}

	if params.Status != "" && params.Status != "all" {
		query = query.Where("status = ?", strings.ToLower(params.Status))
	}

	if params.PaymentType != "" && params.PaymentType != "all" {
		query = query.Where("payment_type = ?", strings.ToLower(params.PaymentType))
	}

	if params.SubaccountRouted == "true" {
		query = query.Where("is_subaccount_routed = ? OR (subaccount_code IS NOT NULL AND subaccount_code != '')", true)
	} else if params.SubaccountRouted == "false" {
		query = query.Where("is_subaccount_routed = ? AND (subaccount_code IS NULL OR subaccount_code = '')", false)
	}

	if params.DisputeStatus != "" && params.DisputeStatus != "all" {
		query = query.Where("dispute_status = ?", strings.ToLower(params.DisputeStatus))
	}

	if params.TenantID != "" {
		if tID, err := uuid.Parse(params.TenantID); err == nil {
			query = query.Where("tenant_id = ?", tID)
		}
	}

	if params.FromDate != "" {
		query = query.Where("created_at >= ?", params.FromDate)
	}
	if params.ToDate != "" {
		query = query.Where("created_at <= ?", params.ToDate)
	}

	var total int64
	query.Count(&total)

	// Compute aggregated stats
	stats := map[string]interface{}{
		"total_volume":             0.0,
		"successful_count":         int64(0),
		"failed_count":             int64(0),
		"disputed_count":           int64(0),
		"subaccount_routed_volume": 0.0,
		"subaccount_routed_count":  int64(0),
		"platform_fee_total":       0.0,
	}

	var statsRows []struct {
		Status             string
		DisputeStatus      string
		IsSubaccountRouted bool
		Amount             float64
		PlatformFee        float64
		Count              int64
	}

	s.db.Table("public.payment_logs").
		Select("status, dispute_status, is_subaccount_routed, SUM(amount) as amount, SUM(platform_fee) as platform_fee, COUNT(id) as count").
		Group("status, dispute_status, is_subaccount_routed").
		Scan(&statsRows)

	for _, r := range statsRows {
		if r.Status == "successful" || r.Status == "paid" || r.Status == "succeeded" {
			stats["total_volume"] = stats["total_volume"].(float64) + r.Amount
			stats["successful_count"] = stats["successful_count"].(int64) + r.Count
			stats["platform_fee_total"] = stats["platform_fee_total"].(float64) + r.PlatformFee
			if r.IsSubaccountRouted {
				stats["subaccount_routed_volume"] = stats["subaccount_routed_volume"].(float64) + r.Amount
				stats["subaccount_routed_count"] = stats["subaccount_routed_count"].(int64) + r.Count
			}
		} else if r.Status == "failed" || r.Status == "declined" {
			stats["failed_count"] = stats["failed_count"].(int64) + r.Count
		}

		if r.DisputeStatus == "under_review" || r.DisputeStatus == "disputed" || r.DisputeStatus == "chargeback" {
			stats["disputed_count"] = stats["disputed_count"].(int64) + r.Count
		}
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}
	page := params.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var logs []models.PaymentLog
	err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error

	return logs, stats, total, err
}

// CreatePaymentLog manually records a payment / settlement / dispute
func (s *AdminService) CreatePaymentLog(input PaymentLogInput) (*models.PaymentLog, error) {
	_ = s.db.AutoMigrate(&models.PaymentLog{})

	if input.Reference == "" {
		input.Reference = fmt.Sprintf("PAY-%d", time.Now().UnixNano()/1e6)
	}
	if input.Currency == "" {
		input.Currency = "GHS"
	}
	if input.Status == "" {
		input.Status = "successful"
	}
	if input.DisputeStatus == "" {
		input.DisputeStatus = "none"
	}
	if input.PaymentType == "" {
		input.PaymentType = "manual_entry"
	}
	if input.PaymentMethod == "" {
		input.PaymentMethod = "paystack"
	}
	if input.Gateway == "" {
		input.Gateway = "manual"
	}

	if input.TenantID != nil && input.TenantName == "" {
		var tenant models.Tenant
		if err := s.db.Table("public.tenants").Where("id = ?", *input.TenantID).First(&tenant).Error; err == nil {
			input.TenantName = tenant.Name
		}
	}

	if input.SubaccountCode != nil && *input.SubaccountCode != "" {
		input.IsSubaccountRouted = true
	}

	paymentLog := models.PaymentLog{
		TenantID:           input.TenantID,
		TenantName:         input.TenantName,
		PaymentType:        input.PaymentType,
		Reference:          input.Reference,
		OrderID:            input.OrderID,
		OrderNumber:        input.OrderNumber,
		Amount:             input.Amount,
		Currency:           input.Currency,
		PaymentMethod:      input.PaymentMethod,
		Gateway:            input.Gateway,
		SubaccountCode:     input.SubaccountCode,
		IsSubaccountRouted: input.IsSubaccountRouted,
		SubaccountShare:    input.SubaccountShare,
		PlatformFee:        input.PlatformFee,
		CustomerName:       input.CustomerName,
		CustomerEmail:      input.CustomerEmail,
		CustomerPhone:      input.CustomerPhone,
		Status:             input.Status,
		DisputeStatus:      input.DisputeStatus,
		Notes:              input.Notes,
	}

	if err := s.db.Table("public.payment_logs").Create(&paymentLog).Error; err != nil {
		return nil, err
	}

	return &paymentLog, nil
}

// GetPaymentLog retrieves a single payment log by ID
func (s *AdminService) GetPaymentLog(id string) (*models.PaymentLog, error) {
	pID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid payment log ID")
	}

	var log models.PaymentLog
	if err := s.db.Table("public.payment_logs").Preload("Tenant").Where("id = ?", pID).First(&log).Error; err != nil {
		return nil, errors.New("payment log not found")
	}

	return &log, nil
}

// UpdatePaymentLog updates status, dispute status, and notes of a payment log
func (s *AdminService) UpdatePaymentLog(id string, input PaymentLogUpdateInput) (*models.PaymentLog, error) {
	pID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid payment log ID")
	}

	updates := map[string]interface{}{}
	if input.Status != nil {
		updates["status"] = *input.Status
	}
	if input.DisputeStatus != nil {
		updates["dispute_status"] = *input.DisputeStatus
	}
	if input.Notes != nil {
		updates["notes"] = *input.Notes
	}

	if len(updates) > 0 {
		if err := s.db.Table("public.payment_logs").Where("id = ?", pID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.GetPaymentLog(id)
}

// DeletePaymentLog removes a payment log entry
func (s *AdminService) DeletePaymentLog(id string) error {
	pID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid payment log ID")
	}
	return s.db.Table("public.payment_logs").Where("id = ?", pID).Delete(&models.PaymentLog{}).Error
}

// BillingPayments (Legacy backward-compatible wrapper)
func (s *AdminService) ListBillingPayments() ([]models.BillingPayment, map[string]interface{}, error) {
	var payments []models.BillingPayment
	stats := map[string]interface{}{
		"total_revenue":    0.0,
		"successful_count": 0,
		"failed_count":     0,
	}

	if !s.db.Migrator().HasTable(&models.BillingPayment{}) {
		return payments, stats, nil
	}

	err := s.db.Preload("Subscription.Tenant").Find(&payments).Error

	// Calculate stats
	for _, p := range payments {
		switch p.Status {
		case "succeeded", "paid":
			stats["successful_count"] = stats["successful_count"].(int) + 1
			stats["total_revenue"] = stats["total_revenue"].(float64) + p.Amount
		case "failed":
			stats["failed_count"] = stats["failed_count"].(int) + 1
		}
	}

	return payments, stats, err
}

// Audit Logs
func (s *AdminService) ListAuditLogs() ([]models.ActivityLog, map[string]interface{}, error) {
	var logs []models.ActivityLog
	stats := map[string]interface{}{
		"total_events":      0,
		"critical_errors":   0,
		"high_risk_actions": 0,
	}

	if !s.db.Migrator().HasTable(&models.ActivityLog{}) {
		return logs, stats, nil
	}

	err := s.db.Preload("Tenant").Preload("Actor").Order("created_at desc").Limit(100).Find(&logs).Error

	// Calculate stats
	for _, l := range logs {
		stats["total_events"] = stats["total_events"].(int) + 1

		// High risk actions (delete, suspend, override, force)
		action := strings.ToLower(l.ActionType)
		if strings.Contains(action, "delete") || strings.Contains(action, "override") || strings.Contains(action, "force") || strings.Contains(action, "suspend") {
			stats["high_risk_actions"] = stats["high_risk_actions"].(int) + 1
		}

		// Count error-like events as critical
		if strings.Contains(action, "error") || strings.Contains(action, "fail") || strings.Contains(action, "denied") {
			stats["critical_errors"] = stats["critical_errors"].(int) + 1
		}
	}

	return logs, stats, err
}

// FAQs
func (s *AdminService) ListFAQs() ([]models.FAQ, error) {
	var faqs []models.FAQ
	if !s.db.Migrator().HasTable(&models.FAQ{}) {
		return faqs, nil
	}
	err := s.db.Order("order_index asc").Find(&faqs).Error
	return faqs, err
}

func (s *AdminService) CreateFAQ(faq *models.FAQ) error {
	return s.db.Create(faq).Error
}

func (s *AdminService) ToggleFAQ(id uint) error {
	var faq models.FAQ
	if err := s.db.First(&faq, "id = ?", id).Error; err != nil {
		return err
	}
	return s.db.Model(&faq).Update("is_published", !faq.IsPublished).Error
}

// App Marketplace (External Systems)
func (s *AdminService) ListExternalSystems() ([]models.ExternalSystem, error) {
	var systems []models.ExternalSystem
	if !s.db.Migrator().HasTable(&models.ExternalSystem{}) {
		return systems, nil
	}
	err := s.db.Find(&systems).Error
	return systems, err
}

func (s *AdminService) ToggleExternalSystem(id uuid.UUID, field string) error {
	var sys models.ExternalSystem
	if err := s.db.First(&sys, "id = ?", id).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{}
	switch field {
	case "active":
		updates["is_active"] = !sys.IsActive
	case "public":
		updates["is_public"] = !sys.IsPublic
	}

	return s.db.Model(&sys).Updates(updates).Error
}

// Database Backups
func (s *AdminService) ListDatabaseBackups() ([]models.DatabaseBackup, error) {
	var backups []models.DatabaseBackup
	if !s.db.Migrator().HasTable(&models.DatabaseBackup{}) {
		return backups, nil
	}
	err := s.db.Order("created_at desc").Find(&backups).Error
	return backups, err
}

func (s *AdminService) GetDatabaseBackup(id uint) (*models.DatabaseBackup, error) {
	var backup models.DatabaseBackup
	if err := s.db.First(&backup, id).Error; err != nil {
		return nil, err
	}
	return &backup, nil
}

func (s *AdminService) TriggerDatabaseBackup() (*models.DatabaseBackup, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	backupDir := "./backups"
	if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.sql", timestamp)
	filepath := filepath.Join(backupDir, filename)

	// Build pg_dump command
	host := cfg.Database.Host
	port := cfg.Database.Port
	user := cfg.Database.User
	dbname := cfg.Database.DBName

	cmd := exec.Command("pg_dump", "-h", host, "-p", port, "-U", user, "-F", "c", "-f", filepath, dbname)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.Database.Password))

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pg_dump failed: %w", err)
	}

	info, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat backup file: %w", err)
	}

	backup := models.DatabaseBackup{
		Filename:  filename,
		FilePath:  filepath,
		SizeBytes: info.Size(),
	}

	if err := s.db.Create(&backup).Error; err != nil {
		return nil, fmt.Errorf("failed to save backup record: %w", err)
	}

	return &backup, nil
}

// ───── Phase 1: Tenant Command Center ─────

// GetTenantDetail returns a single tenant with subscription and recent activity.
func (s *AdminService) GetTenantDetail(id string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := s.db.Preload("Subscription").Preload("Domains").First(&tenant, "id = ?", id).Error
	return &tenant, err
}

// CreateTenant manually creates a new tenant from the admin panel.
func (s *AdminService) CreateTenant(t *models.Tenant) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		trialEnd := time.Now().AddDate(0, 0, 7)
		subscription := models.Subscription{
			TenantID:         t.ID,
			Status:           "trialing",
			CurrentPeriodEnd: &trialEnd,
		}
		return tx.Create(&subscription).Error
	})
}

// SearchTenants filters tenants by name/subdomain and/or status.
func (s *AdminService) SearchTenants(search, status string, limit, offset int) ([]models.Tenant, int64, error) {
	var tenants []models.Tenant
	var total int64
	query := s.db.Model(&models.Tenant{})

	if search != "" {
		query = query.Where("name ILIKE ? OR subdomain ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Preload("Subscription").Order("created_on desc").Limit(limit).Offset(offset).Find(&tenants).Error

	// Populate Plan from PricingPlan manually
	for i := range tenants {
		t := &tenants[i]
		if t.Subscription != nil && t.Subscription.PlanID != nil && t.Subscription.Plan == nil {
			var pricingPlan models.PricingPlan
			if err := s.db.Where("id = ?", *t.Subscription.PlanID).First(&pricingPlan).Error; err == nil {
				t.Subscription.Plan = &models.Plan{
					Name:  pricingPlan.Name,
					Price: pricingPlan.PriceMonthly,
				}
			}
		}
	}

	return tenants, total, err
}

// UpdateTenantNotes saves a note string to the tenant's Metadata JSONB field.
func (s *AdminService) UpdateTenantNotes(id, notes string) error {
	return s.db.Model(&models.Tenant{}).Where("id = ?", id).
		UpdateColumn("metadata", s.db.Raw("jsonb_set(COALESCE(metadata, '{}'), '{admin_notes}', ?::jsonb)", `"`+notes+`"`)).Error
}

// ───── Phase 2: Billing & Revenue Intelligence ─────

// ListUpcomingRenewals returns active subscriptions renewing within the next N days.
func (s *AdminService) ListUpcomingRenewals(days int) ([]models.Subscription, error) {
	var subs []models.Subscription
	if !s.db.Migrator().HasTable(&models.Subscription{}) {
		return subs, nil
	}
	err := s.db.Preload("Tenant").
		Where(fmt.Sprintf("status = 'active' AND current_period_end BETWEEN NOW() AND NOW() + INTERVAL '%d days'", days)).
		Order("current_period_end asc").
		Find(&subs).Error

	// Populate Plan from PricingPlan manually
	for i := range subs {
		sub := &subs[i]
		if sub.PlanID != nil && sub.Plan == nil {
			var pricingPlan models.PricingPlan
			if err := s.db.Where("id = ?", *sub.PlanID).First(&pricingPlan).Error; err == nil {
				sub.Plan = &models.Plan{
					Name:  pricingPlan.Name,
					Price: pricingPlan.PriceMonthly,
				}
			}
		}
	}
	return subs, err
}

// ListFailedPayments returns past-due subscriptions and failed payment records.
func (s *AdminService) ListFailedPayments() ([]models.BillingPayment, error) {
	var payments []models.BillingPayment
	if !s.db.Migrator().HasTable(&models.BillingPayment{}) {
		return payments, nil
	}
	err := s.db.Preload("Subscription.Tenant").Where("status = 'failed' OR status = 'past_due'").
		Order("created_at desc").Find(&payments).Error
	return payments, err
}

// ───── Phase 3: Growth ─────

// ListReferrals returns all referral reward records.
func (s *AdminService) ListReferrals() ([]models.ReferralReward, error) {
	var rewards []models.ReferralReward
	if !s.db.Migrator().HasTable(&models.ReferralReward{}) {
		return rewards, nil
	}
	err := s.db.Preload("Referrer").Preload("ReferredTenant").Order("created_at desc").Find(&rewards).Error
	return rewards, err
}

// ───── Phase 5: Webhook Event Log ─────

// ListWebhookEvents returns the most recent webhook delivery attempts.
func (s *AdminService) ListWebhookEvents() ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	if !s.db.Migrator().HasTable(&models.WebhookEvent{}) {
		return events, nil
	}
	err := s.db.Order("created_at desc").Limit(100).Find(&events).Error
	return events, err
}

// ───── Phase 4: Security & Access Control ─────

func (s *AdminService) ListAdminRoles() ([]models.AdminRole, error) {
	var roles []models.AdminRole
	if !s.db.Migrator().HasTable(&models.AdminRole{}) {
		return roles, nil
	}
	err := s.db.Find(&roles).Error
	return roles, err
}

func (s *AdminService) CreateAdminRole(role *models.AdminRole) error {
	return s.db.Create(role).Error
}

func (s *AdminService) UpdateAdminRole(id string, permissions string) error {
	return s.db.Model(&models.AdminRole{}).Where("id = ?", id).Update("permissions", permissions).Error
}

func (s *AdminService) ListIPAllowlist() ([]models.IPAllowlist, error) {
	var ips []models.IPAllowlist
	if !s.db.Migrator().HasTable(&models.IPAllowlist{}) {
		return ips, nil
	}
	err := s.db.Find(&ips).Error
	return ips, err
}

func (s *AdminService) AddIPToAllowlist(ip *models.IPAllowlist) error {
	return s.db.Create(ip).Error
}

func (s *AdminService) RemoveIPFromAllowlist(id string) error {
	return s.db.Delete(&models.IPAllowlist{}, "id = ?", id).Error
}

func (s *AdminService) ListMasterAPIKeys() ([]models.MasterAPIKey, error) {
	var keys []models.MasterAPIKey
	if !s.db.Migrator().HasTable(&models.MasterAPIKey{}) {
		return keys, nil
	}
	err := s.db.Find(&keys).Error
	return keys, err
}

func (s *AdminService) CreateMasterAPIKey(key *models.MasterAPIKey) error {
	return s.db.Create(key).Error
}

func (s *AdminService) RevokeMasterAPIKey(id string) error {
	return s.db.Model(&models.MasterAPIKey{}).Where("id = ?", id).Update("is_active", false).Error
}

func (s *AdminService) SearchAllDomains(search, status string, limit, offset int) ([]models.Domain, int64, error) {
	var domains []models.Domain
	var total int64
	if !s.db.Migrator().HasTable(&models.Domain{}) {
		return domains, total, nil
	}

	query := s.db.Model(&models.Domain{})
	if search != "" {
		query = query.Where("domain ILIKE ?", "%"+search+"%")
	}
	switch status {
	case "verified":
		query = query.Where("is_verified = ?", true)
	case "pending":
		query = query.Where("is_verified = ?", false)
	}

	query.Count(&total)
	err := query.Preload("Tenant").Order("created_at desc").Limit(limit).Offset(offset).Find(&domains).Error
	return domains, total, err
}

func (s *AdminService) GetDomainDiagnostics(id uint) (map[string]interface{}, error) {
	var domain models.Domain
	if err := s.db.First(&domain, id).Error; err != nil {
		return nil, err
	}

	cname, errCname := net.LookupCNAME(domain.Domain)
	ips, errIPs := net.LookupHost(domain.Domain)

	status := "resolved"
	if errCname != nil && errIPs != nil {
		status = "failed"
	}

	return map[string]interface{}{
		"domain": domain.Domain,
		"cname":  cname,
		"cname_error": func() string {
			if errCname != nil {
				return errCname.Error()
			}
			return ""
		}(),
		"ips": ips,
		"ips_error": func() string {
			if errIPs != nil {
				return errIPs.Error()
			}
			return ""
		}(),
		"status": status,
	}, nil
}

func (s *AdminService) BulkVerifyDomains(ids []uint) error {
	var domains []models.Domain
	if err := s.db.Where("id IN ?", ids).Find(&domains).Error; err != nil {
		return err
	}

	for _, domain := range domains {
		cname, errCname := net.LookupCNAME(domain.Domain)
		_, errIPs := net.LookupHost(domain.Domain)

		isVerified := false
		if errCname == nil && errIPs == nil {
			// Basic check: If it resolves to our proxy or returns IPs, we consider it verified
			// You can refine this logic to match specific proxy CNAME targets (e.g., proxy.puxbay.com)
			if strings.Contains(cname, "proxy.puxbay.com") || len(cname) > 0 {
				isVerified = true
			}
		}

		now := time.Now()
		s.db.Model(&domain).Updates(map[string]interface{}{
			"is_verified":    isVerified,
			"dns_checked_at": &now,
		})
	}
	return nil
}
