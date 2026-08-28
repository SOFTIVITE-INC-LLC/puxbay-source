package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

// SMSHandler handles SMS wallet, sender ID, and admin SMS gateway management.
type SMSHandler struct {
	db          *gorm.DB
	paystackKey string // platform-level Paystack key for wallet top-ups
}

func NewSMSHandler(db *gorm.DB, paystackSecretKey string) *SMSHandler {
	return &SMSHandler{db: db, paystackKey: paystackSecretKey}
}

// ──────────────────────────────────────────────────────────────────────────────
// TENANT ENDPOINTS
// ──────────────────────────────────────────────────────────────────────────────

// GetSMSWallet returns the tenant's current SMS wallet balance, active sender ID, and pricing.
func (h *SMSHandler) GetSMSWallet(c *gin.Context) {
	db := getDB(c, h.db)

	// Ensure wallet exists
	wallet := h.ensureWallet(c, db)
	if wallet == nil {
		return
	}

	// Get active sender ID
	var senderID *models.SMSSenderID
	db.Where("status = ?", "approved").Order("updated_at desc").First(&senderID)

	// Get platform pricing
	cfg := h.getGatewayConfig()

	c.JSON(200, gin.H{
		"wallet":    wallet,
		"sender_id": senderID,
		"rate":      cfg.PricePerSMS,
		"currency":  cfg.PriceCurrency,
	})
}

// GetSMSTransactions returns a paginated list of SMS top-up and usage transactions.
func (h *SMSHandler) GetSMSTransactions(c *gin.Context) {
	db := getDB(c, h.db)

	var txns []models.SMSTransaction
	db.Order("created_at desc").Limit(100).Find(&txns)
	c.JSON(200, txns)
}

// ListSenderIDs returns all Sender IDs submitted by the tenant.
func (h *SMSHandler) ListSenderIDs(c *gin.Context) {
	db := getDB(c, h.db)
	var ids []models.SMSSenderID
	db.Order("created_at desc").Find(&ids)
	c.JSON(200, ids)
}

// SubmitSenderID creates a new Sender ID registration request (status=pending).
func (h *SMSHandler) SubmitSenderID(c *gin.Context) {
	db := getDB(c, h.db)

	var req struct {
		SenderID string `json:"sender_id" binding:"required"`
		Purpose  string `json:"purpose"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if len(req.SenderID) < 3 || len(req.SenderID) > 11 {
		c.JSON(400, gin.H{"error": "Sender ID must be 3–11 characters"})
		return
	}

	tenantID := c.GetString("tenant_id")
	senderID := models.SMSSenderID{
		SenderID: req.SenderID,
		Purpose:  req.Purpose,
		Status:   "pending",
		TenantID: tenantID,
	}
	if err := db.Create(&senderID).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to submit Sender ID: " + err.Error()})
		return
	}
	c.JSON(201, senderID)
}

// InitiateSMSTopup initializes a Paystack payment to top-up the SMS wallet.
func (h *SMSHandler) InitiateSMSTopup(c *gin.Context) {
	db := getDB(c, h.db)

	var req struct {
		Amount float64 `json:"amount" binding:"required"` // monetary amount e.g. 20.00 GHS
		Email  string  `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Amount < 1 {
		c.JSON(400, gin.H{"error": "Minimum top-up amount is 1.00"})
		return
	}

	cfg := h.getGatewayConfig()
	credits := int64(req.Amount / cfg.PricePerSMS)

	// Create pending SMS transaction
	tenantID := c.GetString("tenant_id")
	ref := fmt.Sprintf("SMS-TOPUP-%s-%d", tenantID[:8], time.Now().UnixMilli())
	txn := models.SMSTransaction{
		TenantID:      tenantID,
		Type:          "topup",
		Amount:        req.Amount,
		CreditsAdded:  credits,
		PricePerSMS:   cfg.PricePerSMS,
		Reference:     ref,
		PaymentMethod: "paystack",
		Status:        "pending",
		Description:   fmt.Sprintf("SMS Wallet Top-up: %.2f %s → %d SMS credits", req.Amount, cfg.PriceCurrency, credits),
	}
	if err := db.Create(&txn).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create transaction: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"reference":    ref,
		"amount":       req.Amount,
		"credits":      credits,
		"price_per_sms": cfg.PricePerSMS,
		"currency":     cfg.PriceCurrency,
		"transaction_id": txn.ID,
	})
}

// VerifySMSTopup verifies the Paystack payment and credits the wallet.
func (h *SMSHandler) VerifySMSTopup(c *gin.Context) {
	db := getDB(c, h.db)

	var req struct {
		Reference string `json:"reference" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Find the pending transaction
	var txn models.SMSTransaction
	if err := db.Where("reference = ? AND status = ?", req.Reference, "pending").First(&txn).Error; err != nil {
		c.JSON(404, gin.H{"error": "Transaction not found or already processed"})
		return
	}

	// Verify with Paystack
	if h.paystackKey != "" {
		url := fmt.Sprintf("https://api.paystack.co/transaction/verify/%s", req.Reference)
		httpReq, _ := http.NewRequest("GET", url, nil)
		httpReq.Header.Set("Authorization", "Bearer "+h.paystackKey)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			c.JSON(500, gin.H{"error": "Could not verify payment"})
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var psResp struct {
			Status bool `json:"status"`
			Data   struct {
				Status string  `json:"status"`
				Amount float64 `json:"amount"`
			} `json:"data"`
		}
		if json.Unmarshal(body, &psResp) == nil {
			if !psResp.Status || psResp.Data.Status != "success" {
				db.Model(&txn).Update("status", "failed")
				c.JSON(400, gin.H{"error": "Payment verification failed"})
				return
			}
		}
	}

	// Credit the wallet within a transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		txn.Status = "completed"
		tx.Save(&txn)

		var wallet models.SMSWallet
		if err := tx.Where("tenant_id = ?", txn.TenantID).First(&wallet).Error; err != nil {
			wallet = models.SMSWallet{
				TenantID:    txn.TenantID,
				PricePerSMS: txn.PricePerSMS,
			}
			tx.Create(&wallet)
		}
		wallet.BalanceAmount += txn.Amount
		wallet.CreditsTotal += txn.CreditsAdded
		wallet.CreditsBalance += txn.CreditsAdded
		wallet.PricePerSMS = txn.PricePerSMS
		return tx.Save(&wallet).Error
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to credit wallet: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success", "message": fmt.Sprintf("%d SMS credits added to your wallet", txn.CreditsAdded)})
}

// ──────────────────────────────────────────────────────────────────────────────
// SUPERADMIN ENDPOINTS
// ──────────────────────────────────────────────────────────────────────────────

// AdminGetSMSConfig returns the current platform SMS gateway configuration.
func (h *SMSHandler) AdminGetSMSConfig(c *gin.Context) {
	cfg := h.getGatewayConfig()
	c.JSON(200, cfg)
}

// AdminUpdateSMSConfig updates the platform SMS gateway configuration.
func (h *SMSHandler) AdminUpdateSMSConfig(c *gin.Context) {
	var req struct {
		DefaultSenderID string  `json:"default_sender_id"`
		PricePerSMS     float64 `json:"price_per_sms"`
		PriceCurrency   string  `json:"price_currency"`
		IsActive        *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	cfg := h.getGatewayConfig()

	if req.DefaultSenderID != "" {
		cfg.DefaultSenderID = req.DefaultSenderID
	}
	if req.PricePerSMS > 0 {
		cfg.PricePerSMS = req.PricePerSMS
	}
	if req.PriceCurrency != "" {
		cfg.PriceCurrency = req.PriceCurrency
	}
	if req.IsActive != nil {
		cfg.IsActive = *req.IsActive
	}

	if err := h.db.Table("public.sms_gateway_configs").Save(&cfg).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update SMS config: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "updated", "price_per_sms": cfg.PricePerSMS})
}

// AdminListSenderIDs returns ALL tenant Sender ID requests across the platform.
func (h *SMSHandler) AdminListSenderIDs(c *gin.Context) {
	status := c.Query("status") // pending, approved, rejected, or empty for all

	type SenderIDWithSchema struct {
		models.SMSSenderID
		SchemaName string `json:"schema_name"`
	}

	// We query the public tenant table to get all schemas
	var tenants []models.Tenant
	h.db.Where("is_active = ?", true).Find(&tenants)

	var results []map[string]interface{}
	for _, tenant := range tenants {
		tenantDB := h.db.Exec(fmt.Sprintf("SET search_path TO %s", tenant.SchemaName)).Session(&gorm.Session{NewDB: true})
		tenantDB = tenantDB.Table("s_m_s_sender_i_ds")

		var senderIDs []models.SMSSenderID
		q := tenantDB.Order("created_at desc")
		if status != "" {
			q = q.Where("status = ?", status)
		}
		q.Find(&senderIDs)

		for _, sid := range senderIDs {
			results = append(results, map[string]interface{}{
				"id":               sid.ID,
				"tenant_id":        tenant.SchemaName,
				"tenant_name":      tenant.Name,
				"sender_id":        sid.SenderID,
				"purpose":          sid.Purpose,
				"status":           sid.Status,
				"rejection_reason": sid.RejectionReason,
				"approved_at":      sid.ApprovedAt,
				"created_at":       sid.CreatedAt,
			})
		}
	}

	if results == nil {
		results = []map[string]interface{}{}
	}
	c.JSON(200, results)
}

// AdminApproveSenderID approves a tenant's Sender ID request.
func (h *SMSHandler) AdminApproveSenderID(c *gin.Context) {
	id := c.Param("id")
	tenantSchema := c.Query("schema")
	if tenantSchema == "" {
		c.JSON(400, gin.H{"error": "schema query param required (e.g. ?schema=tenant_acme)"})
		return
	}

	tenantDB := h.db.Exec(fmt.Sprintf("SET search_path TO %s", tenantSchema)).Session(&gorm.Session{NewDB: true})
	parsedID, _ := uuid.Parse(id)
	now := time.Now()

	result := tenantDB.Model(&models.SMSSenderID{}).Where("id = ?", parsedID).Updates(map[string]interface{}{
		"status":      "approved",
		"approved_at": &now,
	})
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Sender ID not found"})
		return
	}
	c.JSON(200, gin.H{"status": "approved"})
}

// AdminRejectSenderID rejects a tenant's Sender ID request with a reason.
func (h *SMSHandler) AdminRejectSenderID(c *gin.Context) {
	id := c.Param("id")
	tenantSchema := c.Query("schema")
	if tenantSchema == "" {
		c.JSON(400, gin.H{"error": "schema query param required"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tenantDB := h.db.Exec(fmt.Sprintf("SET search_path TO %s", tenantSchema)).Session(&gorm.Session{NewDB: true})
	parsedID, _ := uuid.Parse(id)

	result := tenantDB.Model(&models.SMSSenderID{}).Where("id = ?", parsedID).Updates(map[string]interface{}{
		"status":           "rejected",
		"rejection_reason": req.Reason,
	})
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Sender ID not found"})
		return
	}
	c.JSON(200, gin.H{"status": "rejected"})
}

// ──────────────────────────────────────────────────────────────────────────────
// HELPERS
// ──────────────────────────────────────────────────────────────────────────────

func (h *SMSHandler) getGatewayConfig() *models.SMSGatewayConfig {
	var cfg models.SMSGatewayConfig
	// SMSGatewayConfig is a public-schema model. Always query public schema explicitly.
	if err := h.db.Table("public.sms_gateway_configs").Where("deleted_at IS NULL").First(&cfg).Error; err != nil {
		// Create default config
		cfg = models.SMSGatewayConfig{
			Provider:        "arkesel",
			DefaultSenderID: "PUXBAY",
			PricePerSMS:     0.20,
			PriceCurrency:   "GHS",
			IsActive:        true,
		}
		_ = h.db.Table("public.sms_gateway_configs").Create(&cfg).Error
	}
	return &cfg
}

func (h *SMSHandler) ensureWallet(c *gin.Context, db *gorm.DB) *models.SMSWallet {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		tenantID = c.GetHeader("X-Tenant-ID")
	}

	var wallet models.SMSWallet
	if err := db.Where("tenant_id = ?", tenantID).First(&wallet).Error; err != nil {
		// Auto-create wallet
		cfg := h.getGatewayConfig()
		wallet = models.SMSWallet{
			TenantID:    tenantID,
			PricePerSMS: cfg.PricePerSMS,
			Currency:    cfg.PriceCurrency,
		}
		if err := db.Create(&wallet).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to initialize SMS wallet"})
			return nil
		}
	}
	return &wallet
}
