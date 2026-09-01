package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/softivite/puxbay/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// StorefrontAPIHandler handles all storefront-related API endpoints.
type StorefrontAPIHandler struct {
	db          *gorm.DB
	authService *services.AuthService
	paystackCfg *config.PaystackConfig
	redis       *redis.Client
	smsService  *services.SMSService
	pushService *services.PushService
}

func NewStorefrontAPIHandler(db *gorm.DB, authService *services.AuthService, paystackCfg *config.PaystackConfig, rdb *redis.Client, sms *services.SMSService, push *services.PushService) *StorefrontAPIHandler {
	return &StorefrontAPIHandler{db: db, authService: authService, paystackCfg: paystackCfg, redis: rdb, smsService: sms, pushService: push}
}

func (h *StorefrontAPIHandler) service(c *gin.Context) *services.StorefrontService {
	tenantID, _ := c.Get(middleware.ContextKeyTenantID)
	return services.NewStorefrontService(getDB(c, h.db), h.redis, tenantID.(uuid.UUID))
}

type RegisterCustomerReq struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *StorefrontAPIHandler) RegisterCustomer(c *gin.Context) {
	var req RegisterCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantIDVal, exists := c.Get(middleware.ContextKeyTenantID)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No tenant found"})
		return
	}
	tenantID := tenantIDVal.(uuid.UUID)

	db := getDB(c, h.db)

	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure password"})
		return
	}

	var customer models.Customer
	if err := db.Where("email = ?", req.Email).First(&customer).Error; err == nil {
		if customer.IsRegistered {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		customer.Name = req.Name
		customer.PasswordHash = &hashedPassword
		customer.IsRegistered = true
		if err := db.Save(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register customer"})
			return
		}
	} else {
		customer = models.Customer{
			Name:         req.Name,
			Email:        &req.Email,
			PasswordHash: &hashedPassword,
			IsRegistered: true,
			CustomerType: "retail",
		}
		if err := db.Create(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
			return
		}
	}

	tokens, err := h.authService.GenerateTokenPair(customer.ID, tenantID, nil, "customer", nil, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":    tokens.AccessToken,
		"customer": customer,
	})
}

type LoginCustomerReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *StorefrontAPIHandler) LoginCustomer(c *gin.Context) {
	var req LoginCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantIDVal, exists := c.Get(middleware.ContextKeyTenantID)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No tenant found"})
		return
	}
	tenantID := tenantIDVal.(uuid.UUID)

	db := getDB(c, h.db)

	var customer models.Customer
	if err := db.Where("email = ?", req.Email).First(&customer).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !customer.IsRegistered || customer.PasswordHash == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !h.authService.CheckPassword(req.Password, *customer.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	tokens, err := h.authService.GenerateTokenPair(customer.ID, tenantID, nil, "customer", nil, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    tokens.AccessToken,
		"customer": customer,
	})
}

func (h *StorefrontAPIHandler) GetCustomerMe(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	customerID := userIDVal.(uuid.UUID)

	db := getDB(c, h.db)
	var customer models.Customer
	if err := db.First(&customer, customerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

type UpdateCustomerReq struct {
	Name    string `json:"name" binding:"required"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

func (h *StorefrontAPIHandler) UpdateCustomerMe(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	customerID := userIDVal.(uuid.UUID)

	var req UpdateCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := getDB(c, h.db)
	var customer models.Customer
	if err := db.First(&customer, customerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	customer.Name = req.Name
	if req.Phone != "" {
		customer.Phone = &req.Phone
	} else {
		customer.Phone = nil
	}
	if req.Address != "" {
		customer.Address = &req.Address
	} else {
		customer.Address = nil
	}

	if err := db.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func (h *StorefrontAPIHandler) GetCustomerOrders(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	customerID := userIDVal.(uuid.UUID)

	db := getDB(c, h.db)
	var orders []models.Order
	if err := db.Where("customer_id = ?", customerID).Preload("Items").Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetSettings returns the storefront configuration.
func (h *StorefrontAPIHandler) GetSettings(c *gin.Context) {
	settings, _ := h.service(c).GetSettings()
	c.JSON(http.StatusOK, settings)
}

// UpdateSettings updates the storefront configuration (admin only).
func (h *StorefrontAPIHandler) UpdateSettings(c *gin.Context) {
	var req models.StorefrontSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service(c).UpdateSettings(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	if settings == &req { // Created instead of updated
		c.JSON(http.StatusCreated, settings)
		return
	}

	c.JSON(http.StatusOK, settings)
}

// SearchProducts is the public storefront product search/list API.
func (h *StorefrontAPIHandler) SearchProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	params := services.ProductSearchParams{
		BranchID:   middleware.ResolveBranchID(c, c.Query("branch_id")),
		Search:     c.Query("search"),
		CategoryID: c.Query("category_id"),
		MinPrice:   c.Query("min_price"),
		MaxPrice:   c.Query("max_price"),
		InStock:    c.Query("in_stock"),
		SortBy:     c.DefaultQuery("sort", "latest"),
		Page:       page,
		PageSize:   pageSize,
	}

	result, err := h.service(c).SearchProducts(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products":    result.Products,
		"total":       result.Total,
		"page":        result.Page,
		"page_size":   result.PageSize,
		"total_pages": result.TotalPages,
	})
}

// ListCategories returns all categories.
func (h *StorefrontAPIHandler) ListCategories(c *gin.Context) {
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	categories, err := h.service(c).ListCategories(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// GetProduct returns a single public product detail.
func (h *StorefrontAPIHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")

	detail, err := h.service(c).GetProduct(id)
	if err != nil {
		if err.Error() == "invalid product ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product":          detail.Product,
		"images":           detail.Images,
		"reviews":          detail.Reviews,
		"avg_rating":       detail.AvgRating,
		"related_products": detail.RelatedProducts,
	})
}

// TrackOrder allows public order tracking by 8-character tracking code.
func (h *StorefrontAPIHandler) TrackOrder(c *gin.Context) {
	orderNumber := c.Query("order_number")
	if orderNumber == "" {
		orderNumber = c.Query("code")
	}
	if orderNumber == "" {
		orderNumber = c.Param("code")
	}

	order, err := h.service(c).TrackOrder(orderNumber)
	if err != nil {
		if err.Error() == "order_number is required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please enter a valid 8-character order tracking code."})
			return
		}
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "No order found matching tracking code '" + strings.ToUpper(strings.TrimSpace(orderNumber)) + "'."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order": gin.H{
			"order_number":   order.OrderNumber,
			"tracking_code":  order.OrderNumber,
			"status":         order.Status,
			"payment_status": order.PaymentStatus,
			"payment_method": order.PaymentMethod,
			"order_type":     order.OrderType,
			"created_at":     order.CreatedAt,
			"updated_at":     order.UpdatedAt,
			"subtotal":       order.Subtotal,
			"total":          order.Total,
			"total_amount":   order.Total,
			"notes":          order.Notes,
			"receipt_token":  order.ReceiptToken,
			"branch":         order.Branch,
			"customer":       order.Customer,
			"items":          order.Items,
		},
		"order_number":  order.OrderNumber,
		"tracking_code": order.OrderNumber,
		"status":        order.Status,
		"created_at":    order.CreatedAt,
		"total_amount":  order.Total,
		"items":         order.Items,
	})
}

// SubmitReview allows a customer to submit a product review.
func (h *StorefrontAPIHandler) SubmitReview(c *gin.Context) {
	productID := c.Param("id")

	var req struct {
		CustomerID string `json:"customer_id" binding:"required"`
		Rating     int    `json:"rating" binding:"required,min=1,max=5"`
		Comment    string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := services.ReviewInput{
		CustomerID: req.CustomerID,
		Rating:     req.Rating,
		Comment:    req.Comment,
	}

	review, err := h.service(c).SubmitReview(productID, input)
	if err != nil {
		if err.Error() == "invalid product ID" || err.Error() == "invalid customer ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save review"})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// SubscribeNewsletter adds an email to the newsletter list.
func (h *StorefrontAPIHandler) SubscribeNewsletter(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service(c).SubscribeNewsletter(req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to subscribe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully subscribed to newsletter"})
}

// SubscribeBackInStock adds an email to the back in stock notification list for a product.
func (h *StorefrontAPIHandler) SubscribeBackInStock(c *gin.Context) {
	productID := c.Param("id")
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service(c).SubscribeBackInStock(productID, req.Email); err != nil {
		if err.Error() == "invalid product ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to subscribe for back in stock notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully subscribed to back in stock notification"})
}

// ApplyCoupon validates and applies a coupon code.
func (h *StorefrontAPIHandler) ApplyCoupon(c *gin.Context) {
	var req struct {
		Code      string  `json:"code" binding:"required"`
		CartTotal float64 `json:"cart_total" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service(c).ApplyCoupon(req.Code, req.CartTotal)
	if err != nil {
		if err.Error() == "invalid or expired coupon code" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "minimum purchase amount not met" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply coupon"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coupon":          result.Coupon,
		"discount_amount": result.DiscountAmount,
		"new_total":       result.NewTotal,
	})
}

// ListCoupons returns all coupons (admin).
func (h *StorefrontAPIHandler) ListCoupons(c *gin.Context) {
	coupons, err := h.service(c).ListCoupons()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coupons"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"coupons": coupons})
}

// CreateCoupon creates a new coupon (admin).
func (h *StorefrontAPIHandler) CreateCoupon(c *gin.Context) {
	var req models.Coupon
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service(c).CreateCoupon(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon"})
		return
	}
	c.JSON(http.StatusCreated, req)
}

// UpdateCoupon updates a coupon (admin).
func (h *StorefrontAPIHandler) UpdateCoupon(c *gin.Context) {
	id := c.Param("id")

	var req models.Coupon
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon, err := h.service(c).UpdateCoupon(id, req)
	if err != nil {
		if err.Error() == "invalid coupon ID" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "coupon not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update coupon"})
		return
	}

	c.JSON(http.StatusOK, coupon)
}

func (h *StorefrontAPIHandler) ListProducts(c *gin.Context) {
	// Alias to SearchProducts logic
	h.SearchProducts(c)
}

func (h *StorefrontAPIHandler) GetCart(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	var cart models.AbandonedCart
	if err := getDB(c, h.db).Where("email = ?", sessionID).First(&cart).Error; err != nil {
		c.JSON(200, gin.H{"cart": []interface{}{}})
		return
	}
	c.JSON(200, gin.H{"cart": cart.CartData})
}

type CartActionReq struct {
	ProductID string  `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required"`
}

func getOrInitCart(tx *gorm.DB, sessionID string) (*models.AbandonedCart, []CartActionReq, error) {
	var cart models.AbandonedCart
	if err := tx.Where("email = ?", sessionID).First(&cart).Error; err != nil {
		cart = models.AbandonedCart{Email: sessionID, CartData: datatypes.JSON("[]")}
		if err := tx.Create(&cart).Error; err != nil {
			return nil, nil, err
		}
	}

	var items []CartActionReq
	if len(cart.CartData) > 0 {
		if err := json.Unmarshal(cart.CartData, &items); err != nil {
			items = []CartActionReq{}
		}
	} else {
		items = []CartActionReq{}
	}
	return &cart, items, nil
}

// Gap #15: Real cart operations backed by AbandonedCart table
func (h *StorefrontAPIHandler) AddToCart(c *gin.Context) {
	var req CartActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Session-ID header required"})
		return
	}

	tx := getDB(c, h.db)
	cart, items, err := getOrInitCart(tx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access cart"})
		return
	}

	found := false
	for i, item := range items {
		if item.ProductID == req.ProductID {
			items[i].Quantity += req.Quantity
			found = true
			break
		}
	}
	if !found {
		items = append(items, req)
	}

	b, _ := json.Marshal(items)
	cart.CartData = datatypes.JSON(b)
	tx.Save(cart)

	c.JSON(http.StatusOK, gin.H{"status": "added", "item": req, "cart": items})
}

func (h *StorefrontAPIHandler) UpdateCart(c *gin.Context) {
	var req CartActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Session-ID header required"})
		return
	}

	tx := getDB(c, h.db)
	cart, items, err := getOrInitCart(tx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access cart"})
		return
	}

	for i, item := range items {
		if item.ProductID == req.ProductID {
			items[i].Quantity = req.Quantity
			break
		}
	}

	b, _ := json.Marshal(items)
	cart.CartData = datatypes.JSON(b)
	tx.Save(cart)

	c.JSON(http.StatusOK, gin.H{"status": "updated", "item": req, "cart": items})
}

func (h *StorefrontAPIHandler) RemoveFromCart(c *gin.Context) {
	id := c.Param("id")
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Session-ID header required"})
		return
	}

	tx := getDB(c, h.db)
	cart, items, err := getOrInitCart(tx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access cart"})
		return
	}

	var newItems []CartActionReq
	for _, item := range items {
		if item.ProductID != id {
			newItems = append(newItems, item)
		}
	}

	b, _ := json.Marshal(newItems)
	cart.CartData = datatypes.JSON(b)
	tx.Save(cart)

	tx.Save(cart)

	c.JSON(http.StatusOK, gin.H{"status": "removed", "id": id, "cart": newItems})
}

type CheckoutReq struct {
	Reference       string          `json:"reference"`
	PaymentMethod   string          `json:"payment_method"`
	CustomerID      string          `json:"customer_id"`
	BranchID        string          `json:"branch_id"`
	CustomerName    string          `json:"customer_name"`
	CustomerEmail   string          `json:"customer_email"`
	CustomerPhone   string          `json:"customer_phone"`
	Total           float64         `json:"total" binding:"required"`
	DeliveryMethod  string          `json:"delivery_method"`
	DeliveryAddress string          `json:"delivery_address"`
	OrderNotes      string          `json:"order_notes"`
	Items           []CartActionReq `json:"items" binding:"required"`
}

func (h *StorefrontAPIHandler) VerifyPaystackCheckout(c *gin.Context) {
	var req CheckoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var tenantID uuid.UUID
	if tid, exists := c.Get(middleware.ContextKeyTenantID); exists {
		if id, ok := tid.(uuid.UUID); ok {
			tenantID = id
		}
	}

	if strings.TrimSpace(req.CustomerPhone) == "" {
		c.JSON(400, gin.H{"error": "Customer phone number is required for order verification and updates"})
		return
	}

	sessionID := c.GetHeader("X-Session-ID")

	if req.PaymentMethod != "pickup" && req.PaymentMethod != "cash" {
		if req.Reference == "" {
			c.JSON(400, gin.H{"error": "Payment reference is required"})
			return
		}
		// 1. Verify with Paystack API if configured
		if h.paystackCfg != nil && h.paystackCfg.SecretKey != "" {
			url := fmt.Sprintf("https://api.paystack.co/transaction/verify/%s", req.Reference)
			reqHttp, err := http.NewRequest("GET", url, nil)
			if err == nil {
				reqHttp.Header.Set("Authorization", "Bearer "+h.paystackCfg.SecretKey)
				client := &http.Client{Timeout: 10 * time.Second}
				resp, err := client.Do(reqHttp)
				if err == nil {
					defer resp.Body.Close()
					body, _ := io.ReadAll(resp.Body)
					var paystackResp struct {
						Status bool `json:"status"`
						Data   struct {
							Status string  `json:"status"`
							Amount float64 `json:"amount"` // in kobo/lowest denomination
						} `json:"data"`
					}
					if err := json.Unmarshal(body, &paystackResp); err == nil {
						if !paystackResp.Status || paystackResp.Data.Status != "success" {
							c.JSON(400, gin.H{"error": "Payment verification failed"})
							return
						}
					}
				}
			}
		}
	}

	// 2. Process Order
	var orderNumber string
	err := getDB(c, h.db).Transaction(func(tx *gorm.DB) error {
		var customerID *uuid.UUID
		if req.CustomerID != "" {
			if parsed, err := uuid.Parse(req.CustomerID); err == nil {
				customerID = &parsed
			}
		}

		// Find or create customer if not provided directly
		if customerID == nil && (req.CustomerEmail != "" || req.CustomerPhone != "" || req.CustomerName != "") {
			var cust models.Customer
			findQuery := tx.Model(&models.Customer{})
			if req.CustomerEmail != "" {
				findQuery = findQuery.Where("email = ?", req.CustomerEmail)
			} else if req.CustomerPhone != "" {
				findQuery = findQuery.Where("phone = ?", req.CustomerPhone)
			}
			if err := findQuery.First(&cust).Error; err == nil {
				customerID = &cust.ID
			} else {
				name := req.CustomerName
				if name == "" {
					name = "Online Customer"
				}
				newCust := models.Customer{
					Name: name,
				}
				if req.CustomerEmail != "" {
					newCust.Email = &req.CustomerEmail
				}
				if req.CustomerPhone != "" {
					newCust.Phone = &req.CustomerPhone
				}
				if req.DeliveryAddress != "" {
					newCust.Address = &req.DeliveryAddress
				}
				if err := tx.Create(&newCust).Error; err == nil {
					customerID = &newCust.ID
				}
			}
		}

		// Resolve branch ID
		var branchID *uuid.UUID
		if req.BranchID != "" && req.BranchID != "default" {
			if parsed, err := uuid.Parse(req.BranchID); err == nil {
				branchID = &parsed
			}
		}
		if branchID == nil {
			if bHeader := c.GetHeader("X-Branch-ID"); bHeader != "" {
				if parsed, err := uuid.Parse(bHeader); err == nil {
					branchID = &parsed
				}
			}
		}
		if branchID == nil {
			if bQuery := c.Query("branch_id"); bQuery != "" {
				if parsed, err := uuid.Parse(bQuery); err == nil {
					branchID = &parsed
				}
			}
		}
		if branchID == nil {
			var branch models.Branch
			if err := tx.Where("is_active = ?", true).Order("created_at asc").First(&branch).Error; err == nil {
				branchID = &branch.ID
			}
		}

		var calculatedTotal float64
		var orderItems []models.OrderItem

		for _, item := range req.Items {
			pID, _ := uuid.Parse(item.ProductID)

			var product models.Product
			if err := tx.Where("id = ? AND is_active = ?", pID, true).First(&product).Error; err != nil {
				return fmt.Errorf("product %s not found or inactive", item.ProductID)
			}

			itemTotal := product.SellingPrice * float64(item.Quantity)
			calculatedTotal += itemTotal

			oItem := models.OrderItem{
				ProductID: pID,
				Quantity:  item.Quantity,
				UnitPrice: product.SellingPrice,
				CostPrice: product.CostPrice,
				Total:     itemTotal,
			}
			orderItems = append(orderItems, oItem)

			// Deduct tracked inventory
			if product.TrackInventory {
				tx.Model(&models.Product{}).Where("id = ?", pID).
					UpdateColumn("current_stock", gorm.Expr("current_stock - ?", item.Quantity))
			}
		}

		paymentStatus := "paid"
		paymentMethod := "paystack"
		amountPaid := calculatedTotal
		if req.PaymentMethod == "pickup" || req.PaymentMethod == "cash" {
			paymentStatus = "unpaid"
			paymentMethod = "cash"
			amountPaid = 0
		}

		notes := ""
		if req.Reference != "" {
			notes += "Paystack Reference: " + req.Reference + "\n"
		}
		if req.DeliveryMethod != "" {
			notes += "Fulfillment: " + req.DeliveryMethod + "\n"
		}
		if req.DeliveryAddress != "" {
			notes += "Delivery Address: " + req.DeliveryAddress + "\n"
		}
		if req.OrderNotes != "" {
			notes += "Notes: " + req.OrderNotes
		}

		var createdOrder models.Order
		createdOrder = models.Order{
			BranchScoped:    models.BranchScoped{BranchID: branchID},
			OrderNumber:     generateOrderNumber(),
			CustomerID:      customerID,
			CustomerName:    &req.CustomerName,
			CustomerPhone:   &req.CustomerPhone,
			DeliveryAddress: &req.DeliveryAddress,
			Subtotal:        calculatedTotal,
			Total:           calculatedTotal,
			AmountPaid:      amountPaid,
			Status:          "pending",
			PaymentStatus:   paymentStatus,
			PaymentMethod:   paymentMethod,
			OrderType:       "online",
			Notes:           &notes,
			ReceiptToken:    generateReceiptToken(),
			IsOTPVerified:   false,
			Items:           orderItems,
		}
		if err := tx.Create(&createdOrder).Error; err != nil {
			return err
		}

		// Dispatch SMS to customer with tracking code
		if h.smsService != nil {
			targetPhone := req.CustomerPhone
			targetName := req.CustomerName
			if targetPhone == "" && customerID != nil {
				var cust models.Customer
				if tx.First(&cust, "id = ?", customerID).Error == nil && cust.Phone != nil {
					targetPhone = *cust.Phone
					targetName = cust.Name
				}
			}
			if targetPhone != "" {
				scheme := "https"
				if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
					scheme = "http"
				}
				trackURL := fmt.Sprintf("%s://%s/store/track-order?code=%s", scheme, c.Request.Host, createdOrder.OrderNumber)
				var storeName string
				var tenant models.Tenant
				if tx.First(&tenant).Error == nil && tenant.Name != "" {
					storeName = tenant.Name
				} else {
					storeName = "Store"
				}
				h.smsService.SendStorefrontOrderSMS(tx, targetPhone, targetName, createdOrder.OrderNumber, createdOrder.Total, storeName, trackURL)
			}
		}

		if sessionID != "" {
			tx.Model(&models.AbandonedCart{}).Where("email = ? OR email = ?", sessionID, req.CustomerID).Update("is_recovered", true)
		}

		orderNumber = createdOrder.OrderNumber
		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Checkout failed: " + err.Error()})
		return
	}

	// Broadcast Real-time Push & Sound to Admins/Cashiers
	if h.pushService != nil && tenantID != uuid.Nil {
		custName := req.CustomerName
		if custName == "" {
			custName = "Online Customer"
		}
		h.pushService.SendToTenantAdminsWithSound(
			tenantID,
			fmt.Sprintf("New Online Order #%s", orderNumber),
			fmt.Sprintf("New online order from %s (%s). Total: GH₵%.2f", custName, req.DeliveryMethod, req.Total),
			"online_order",
			"/orders",
			"online_order",
			"online_order",
		)
	}

	c.JSON(201, gin.H{
		"status":        "checkout_successful",
		"order_number":  orderNumber,
		"tracking_code": orderNumber,
		"reference":     req.Reference,
	})
}

func (h *StorefrontAPIHandler) UpdateCartEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Session-ID header required"})
		return
	}

	tx := getDB(c, h.db)
	cart, _, err := getOrInitCart(tx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access cart"})
		return
	}

	cart.Email = req.Email
	if err := tx.Save(cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "email_captured"})
}

func generateOrderNumber() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 8)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			b[i] = charset[i%len(charset)]
		} else {
			b[i] = charset[num.Int64()]
		}
	}
	return string(b)
}

func generateReceiptToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *StorefrontAPIHandler) ToggleWishlist(c *gin.Context) {
	productIDStr := c.Param("id")
	customerIDStr := c.Query("customer_id")

	if customerIDStr == "" {
		c.JSON(400, gin.H{"error": "customer_id is required"})
		return
	}

	pID, _ := uuid.Parse(productIDStr)
	cID, _ := uuid.Parse(customerIDStr)

	var wl models.Wishlist
	err := getDB(c, h.db).Where("customer_id = ? AND product_id = ?", cID, pID).First(&wl).Error
	if err != nil {
		// Create
		wl = models.Wishlist{CustomerID: cID, ProductID: pID}
		getDB(c, h.db).Create(&wl)
		c.JSON(201, gin.H{"status": "added", "wishlist": wl})
	} else {
		// Remove
		getDB(c, h.db).Delete(&wl)
		c.JSON(200, gin.H{"status": "removed"})
	}
}

func (h *StorefrontAPIHandler) RemoveCoupon(c *gin.Context) {
	c.JSON(200, gin.H{"status": "coupon_removed"})
}

type ConvertGuestReq struct {
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

func (h *StorefrontAPIHandler) ConvertGuestToAccount(c *gin.Context) {
	var req ConvertGuestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Session-ID required"})
		return
	}

	tenantIDVal, exists := c.Get(middleware.ContextKeyTenantID)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No tenant found"})
		return
	}
	tenantID := tenantIDVal.(uuid.UUID)

	tx := getDB(c, h.db)
	var cart models.AbandonedCart
	if err := tx.Where("email = ?", sessionID).First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart/Session not found"})
		return
	}

	if cart.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email associated with this session"})
		return
	}

	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure password"})
		return
	}

	var customer models.Customer
	if err := tx.Where("email = ?", cart.Email).First(&customer).Error; err == nil {
		if customer.IsRegistered {
			c.JSON(http.StatusConflict, gin.H{"error": "Account already exists"})
			return
		}
		customer.Name = req.Name
		customer.PasswordHash = &hashedPassword
		customer.IsRegistered = true
		if err := tx.Save(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
			return
		}
	} else {
		customer = models.Customer{
			Name:         req.Name,
			Email:        &cart.Email,
			PasswordHash: &hashedPassword,
			IsRegistered: true,
			CustomerType: "retail",
		}
		if err := tx.Create(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
			return
		}
	}

	tokens, err := h.authService.GenerateTokenPair(customer.ID, tenantID, nil, "customer", nil, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":    tokens.AccessToken,
		"customer": customer,
	})
}

func (h *StorefrontAPIHandler) GetCustomerWishlist(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	customerID := userIDVal.(uuid.UUID)

	db := getDB(c, h.db)
	var wishlists []models.Wishlist
	if err := db.Where("customer_id = ?", customerID).Find(&wishlists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wishlist"})
		return
	}

	var productIDs []uuid.UUID
	for _, w := range wishlists {
		productIDs = append(productIDs, w.ProductID)
	}

	c.JSON(http.StatusOK, gin.H{"product_ids": productIDs})
}

func (h *StorefrontAPIHandler) ToggleCustomerWishlist(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	customerID := userIDVal.(uuid.UUID)

	productIDStr := c.Param("id")
	pID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	db := getDB(c, h.db)
	var wl models.Wishlist
	err = db.Where("customer_id = ? AND product_id = ?", customerID, pID).First(&wl).Error
	if err != nil {
		// Create
		wl = models.Wishlist{CustomerID: customerID, ProductID: pID}
		db.Create(&wl)
		c.JSON(http.StatusCreated, gin.H{"status": "added", "wishlist": wl})
	} else {
		// Remove
		db.Delete(&wl)
		c.JSON(http.StatusOK, gin.H{"status": "removed", "wishlist": wl})
	}
}

// GetGoogleMerchantFeed generates an RSS 2.0 XML product feed for Google Merchant Center
func (h *StorefrontAPIHandler) GetGoogleMerchantFeed(c *gin.Context) {
	db := getDB(c, h.db)
	var products []models.Product
	if err := db.Where("is_active = ?", true).Preload("Category").Find(&products).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to load products")
		return
	}

	scheme := "https"
	if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	host := c.Request.Host

	var xmlBuilder strings.Builder
	xmlBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	xmlBuilder.WriteString(`<rss version="2.0" xmlns:g="http://base.google.com/ns/1.0">` + "\n")
	xmlBuilder.WriteString("<channel>\n")
	xmlBuilder.WriteString("<title>Puxbay Store Products</title>\n")
	xmlBuilder.WriteString(fmt.Sprintf("<link>%s://%s/store</link>\n", scheme, host))
	xmlBuilder.WriteString("<description>Product Catalog Feed for Google Merchant Center</description>\n")

	for _, p := range products {
		avail := "in_stock"
		if p.CurrentStock <= 0 {
			avail = "out_of_stock"
		}
		categoryName := ""
		if p.Category != nil {
			categoryName = p.Category.Name
		}
		imgLink := ""
		if p.Image != nil && *p.Image != "" {
			imgLink = *p.Image
		} else {
			imgLink = fmt.Sprintf("%s://%s/assets/placeholder.png", scheme, host)
		}
		productLink := fmt.Sprintf("%s://%s/store/product/%s", scheme, host, p.ID)

		xmlBuilder.WriteString("<item>\n")
		xmlBuilder.WriteString(fmt.Sprintf("<g:id>%s</g:id>\n", p.ID))
		xmlBuilder.WriteString(fmt.Sprintf("<g:title><![CDATA[%s]]></g:title>\n", p.Name))
		xmlBuilder.WriteString(fmt.Sprintf("<g:description><![CDATA[%s]]></g:description>\n", p.Description))
		xmlBuilder.WriteString(fmt.Sprintf("<g:link>%s</g:link>\n", productLink))
		xmlBuilder.WriteString(fmt.Sprintf("<g:image_link>%s</g:image_link>\n", imgLink))
		xmlBuilder.WriteString("<g:condition>new</g:condition>\n")
		xmlBuilder.WriteString(fmt.Sprintf("<g:availability>%s</g:availability>\n", avail))
		xmlBuilder.WriteString(fmt.Sprintf("<g:price>%.2f GHS</g:price>\n", p.SellingPrice))
		if categoryName != "" {
			xmlBuilder.WriteString(fmt.Sprintf("<g:product_type><![CDATA[%s]]></g:product_type>\n", categoryName))
		}
		xmlBuilder.WriteString("<g:brand>Puxbay</g:brand>\n")
		xmlBuilder.WriteString("</item>\n")
	}

	xmlBuilder.WriteString("</channel>\n")
	xmlBuilder.WriteString("</rss>\n")

	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(xmlBuilder.String()))
}

// GetFacebookCatalogFeed generates a standard CSV product feed for Meta / Facebook Catalog
func (h *StorefrontAPIHandler) GetFacebookCatalogFeed(c *gin.Context) {
	db := getDB(c, h.db)
	var products []models.Product
	if err := db.Where("is_active = ?", true).Preload("Category").Find(&products).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to load products")
		return
	}

	scheme := "https"
	if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	host := c.Request.Host

	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	// Write CSV Header
	w.Write([]string{"id", "title", "description", "availability", "condition", "price", "link", "image_link", "brand", "product_type"})

	for _, p := range products {
		avail := "in stock"
		if p.CurrentStock <= 0 {
			avail = "out of stock"
		}
		categoryName := ""
		if p.Category != nil {
			categoryName = p.Category.Name
		}
		imgLink := ""
		if p.Image != nil && *p.Image != "" {
			imgLink = *p.Image
		} else {
			imgLink = fmt.Sprintf("%s://%s/assets/placeholder.png", scheme, host)
		}
		productLink := fmt.Sprintf("%s://%s/store/product/%s", scheme, host, p.ID)

		w.Write([]string{
			p.ID.String(),
			p.Name,
			p.Description,
			avail,
			"new",
			fmt.Sprintf("%.2f GHS", p.SellingPrice),
			productLink,
			imgLink,
			"Puxbay",
			categoryName,
		})
	}
	w.Flush()

	c.Data(http.StatusOK, "text/csv; charset=utf-8", b.Bytes())
}
