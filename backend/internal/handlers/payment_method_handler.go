package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type PaymentMethodHandler struct {
	db          *gorm.DB
	paystackCfg *config.PaystackConfig
	httpClient  *http.Client
}

func NewPaymentMethodHandler(db *gorm.DB, paystackCfg *config.PaystackConfig) *PaymentMethodHandler {
	return &PaymentMethodHandler{
		db:          db,
		paystackCfg: paystackCfg,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *PaymentMethodHandler) getDB(c *gin.Context) *gorm.DB {
	return getDB(c, h.db)
}

// List returns all payment methods for the active tenant.
func (h *PaymentMethodHandler) List(c *gin.Context) {
	var methods []models.PaymentMethod
	if err := h.getDB(c).Order("created_at ASC").Find(&methods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment methods"})
		return
	}
	c.JSON(http.StatusOK, methods)
}

// Create creates a new payment method.
func (h *PaymentMethodHandler) Create(c *gin.Context) {
	var method models.PaymentMethod
	if err := c.ShouldBindJSON(&method); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if method.ID == uuid.Nil {
		method.ID = uuid.New()
	}

	if err := h.getDB(c).Create(&method).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment method"})
		return
	}
	c.JSON(http.StatusCreated, method)
}

// Update updates an existing payment method.
func (h *PaymentMethodHandler) Update(c *gin.Context) {
	id := c.Param("id")
	methodID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method ID"})
		return
	}

	var existing models.PaymentMethod
	if err := h.getDB(c).First(&existing, "id = ?", methodID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment method not found"})
		return
	}

	var req struct {
		Name                   *string `json:"name"`
		Provider               *string `json:"provider"`
		IsActive               *bool   `json:"is_active"`
		APIKeyHint             *string `json:"api_key_hint"`
		PaystackSubaccountCode *string `json:"paystack_subaccount_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Provider != nil {
		updates["provider"] = *req.Provider
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.APIKeyHint != nil {
		updates["api_key_hint"] = *req.APIKeyHint
	}
	if req.PaystackSubaccountCode != nil {
		updates["paystack_subaccount_code"] = *req.PaystackSubaccountCode
	}

	if len(updates) > 0 {
		if err := h.getDB(c).Model(&existing).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment method"})
			return
		}
	}

	// Fetch fresh copy
	_ = h.getDB(c).First(&existing, "id = ?", methodID)
	c.JSON(http.StatusOK, existing)
}

// Delete deletes a payment method.
func (h *PaymentMethodHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	methodID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method ID"})
		return
	}
	if err := h.getDB(c).Delete(&models.PaymentMethod{}, "id = ?", methodID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment method not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// PaystackSubaccountItem defines Paystack subaccount details returned from Paystack API.
type PaystackSubaccountItem struct {
	ID               int64   `json:"id"`
	SubaccountCode   string  `json:"subaccount_code"`
	BusinessName     string  `json:"business_name"`
	SettlementBank   string  `json:"settlement_bank"`
	AccountNumber    string  `json:"account_number"`
	PercentageCharge float64 `json:"percentage_charge"`
	Currency         string  `json:"currency"`
	Active           bool    `json:"active"`
	IsVerified       bool    `json:"is_verified"`
}

// ListPaystackSubaccounts calls Paystack API to fetch all available subaccounts.
// GET /api/v1/payment-methods/paystack/subaccounts
func (h *PaymentMethodHandler) ListPaystackSubaccounts(c *gin.Context) {
	if h.paystackCfg == nil || h.paystackCfg.SecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paystack is not configured on the server. Please check PAYSTACK_SECRET_KEY."})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", "https://api.paystack.co/subaccount?perPage=50", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Paystack request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+h.paystackCfg.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach Paystack API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read Paystack response"})
		return
	}

	var paystackResp struct {
		Status  bool                     `json:"status"`
		Message string                   `json:"message"`
		Data    []PaystackSubaccountItem `json:"data"`
	}

	if err := json.Unmarshal(body, &paystackResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Paystack response"})
		return
	}

	if !paystackResp.Status {
		c.JSON(resp.StatusCode, gin.H{"error": paystackResp.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subaccounts": paystackResp.Data,
	})
}

// VerifyPaystackSubaccount verifies a specific subaccount code with Paystack.
// GET /api/v1/payment-methods/paystack/subaccounts/verify/:code
func (h *PaymentMethodHandler) VerifyPaystackSubaccount(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subaccount code is required"})
		return
	}

	if h.paystackCfg == nil || h.paystackCfg.SecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paystack is not configured on the server. Please check PAYSTACK_SECRET_KEY."})
		return
	}

	url := fmt.Sprintf("https://api.paystack.co/subaccount/%s", code)
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Paystack request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+h.paystackCfg.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach Paystack API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read Paystack response"})
		return
	}

	var paystackResp struct {
		Status  bool                   `json:"status"`
		Message string                 `json:"message"`
		Data    PaystackSubaccountItem `json:"data"`
	}

	if err := json.Unmarshal(body, &paystackResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Paystack response"})
		return
	}

	if !paystackResp.Status {
		c.JSON(http.StatusNotFound, gin.H{"error": paystackResp.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subaccount": paystackResp.Data,
	})
}

// ── GUIDED SUBACCOUNT CREATION FLOW ─────────────────────────────────────────

// ListPaystackCountries returns Paystack-supported countries.
// GET /api/v1/payment-methods/paystack/countries
func (h *PaymentMethodHandler) ListPaystackCountries(c *gin.Context) {
	if h.paystackCfg == nil || h.paystackCfg.SecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paystack is not configured"})
		return
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", "https://api.paystack.co/country", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+h.paystackCfg.SecretKey)
	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach Paystack: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Paystack response"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ListPaystackBanks returns banks available for the given country.
// GET /api/v1/payment-methods/paystack/banks?country=ghana&currency=GHS
func (h *PaymentMethodHandler) ListPaystackBanks(c *gin.Context) {
	if h.paystackCfg == nil || h.paystackCfg.SecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paystack is not configured"})
		return
	}
	country := strings.TrimSpace(c.Query("country"))
	currency := strings.TrimSpace(c.Query("currency"))
	if country == "" {
		country = "ghana"
	}
	url := fmt.Sprintf("https://api.paystack.co/bank?country=%s&use_cursor=false&perPage=100", country)
	if currency != "" {
		url += "&currency=" + currency
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+h.paystackCfg.SecretKey)
	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach Paystack: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Paystack response"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ResolvePaystackAccount resolves an account number to the registered account name.
// GET /api/v1/payment-methods/paystack/resolve-account?account_number=xxx&bank_code=xxx
func (h *PaymentMethodHandler) ResolvePaystackAccount(c *gin.Context) {
	if h.paystackCfg == nil || h.paystackCfg.SecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paystack is not configured"})
		return
	}
	accountNumber := strings.TrimSpace(c.Query("account_number"))
	bankCode := strings.TrimSpace(c.Query("bank_code"))
	if accountNumber == "" || bankCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_number and bank_code are required"})
		return
	}
	url := fmt.Sprintf("https://api.paystack.co/bank/resolve?account_number=%s&bank_code=%s", accountNumber, bankCode)
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+h.paystackCfg.SecretKey)
	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach Paystack: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var paystackResp struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AccountNumber string `json:"account_number"`
			AccountName   string `json:"account_name"`
			BankID        int    `json:"bank_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &paystackResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Paystack response"})
		return
	}
	if !paystackResp.Status {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": paystackResp.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"account_number": paystackResp.Data.AccountNumber,
		"account_name":   paystackResp.Data.AccountName,
		"bank_id":        paystackResp.Data.BankID,
	})
}

// CreatePaystackSubaccount creates a subaccount on Paystack and saves it locally.
// POST /api/v1/payment-methods/paystack/create-subaccount
func (h *PaymentMethodHandler) CreatePaystackSubaccount(c *gin.Context) {
	if h.paystackCfg == nil || h.paystackCfg.SecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paystack is not configured"})
		return
	}
	var body struct {
		BusinessName        string  `json:"business_name" binding:"required"`
		SettlementBank      string  `json:"settlement_bank" binding:"required"` // bank code e.g. "058"
		AccountNumber       string  `json:"account_number" binding:"required"`
		PercentageCharge    float64 `json:"percentage_charge"`
		Description         string  `json:"description"`
		PrimaryContactEmail string  `json:"primary_contact_email"`
		LocalName           string  `json:"local_name"`
		IsActive            bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paystackPayload := map[string]interface{}{
		"business_name":     body.BusinessName,
		"settlement_bank":   body.SettlementBank,
		"account_number":    body.AccountNumber,
		"percentage_charge": body.PercentageCharge,
	}
	if body.Description != "" {
		paystackPayload["description"] = body.Description
	}
	if body.PrimaryContactEmail != "" {
		paystackPayload["primary_contact_email"] = body.PrimaryContactEmail
	}

	payloadBytes, _ := json.Marshal(paystackPayload)
	req, err := http.NewRequestWithContext(c.Request.Context(), "POST",
		"https://api.paystack.co/subaccount", bytes.NewBuffer(payloadBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+h.paystackCfg.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach Paystack: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var paystackResp struct {
		Status  bool                   `json:"status"`
		Message string                 `json:"message"`
		Data    PaystackSubaccountItem `json:"data"`
	}
	if err := json.Unmarshal(respBody, &paystackResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Paystack response"})
		return
	}
	if !paystackResp.Status {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": paystackResp.Message})
		return
	}

	sub := paystackResp.Data
	localName := body.LocalName
	if localName == "" {
		localName = fmt.Sprintf("Paystack – %s (%s)", sub.BusinessName, sub.SettlementBank)
	}
	code := sub.SubaccountCode
	method := models.PaymentMethod{
		Name:                   localName,
		Provider:               "paystack_subaccount",
		IsActive:               body.IsActive,
		PaystackSubaccountCode: &code,
	}
	if err := h.getDB(c).Create(&method).Error; err != nil {
		c.JSON(http.StatusCreated, gin.H{
			"subaccount": sub,
			"warning":    "Created on Paystack but not saved locally: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"subaccount":     sub,
		"payment_method": method,
	})
}

// VerifyTransaction verifies any Paystack transaction reference
// GET /api/v1/payment-methods/paystack/verify/:reference or GET /api/v1/pos/verify-payment?reference=...
func (h *PaymentMethodHandler) VerifyTransaction(c *gin.Context) {
	ref := c.Param("reference")
	if ref == "" {
		ref = c.Query("reference")
	}
	if ref == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reference is required"})
		return
	}

	secretKey := h.paystackCfg.SecretKey
	if secretKey == "" {
		secretKey = os.Getenv("PAYSTACK_SECRET_KEY")
	}

	url := fmt.Sprintf("https://api.paystack.co/transaction/verify/%s", ref)
	reqHttp, err := http.NewRequestWithContext(c.Request.Context(), "GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	reqHttp.Header.Set("Authorization", "Bearer "+secretKey)

	resp, err := h.httpClient.Do(reqHttp)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to contact Paystack"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var paystackResp struct {
		Status  bool        `json:"status"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &paystackResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response from Paystack"})
		return
	}

	c.JSON(resp.StatusCode, paystackResp)
}
