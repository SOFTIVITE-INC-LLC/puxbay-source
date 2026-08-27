package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
}

func NewStorefrontAPIHandler(db *gorm.DB, authService *services.AuthService, paystackCfg *config.PaystackConfig, rdb *redis.Client) *StorefrontAPIHandler {
	return &StorefrontAPIHandler{db: db, authService: authService, paystackCfg: paystackCfg, redis: rdb}
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

// TrackOrder allows public order tracking.
func (h *StorefrontAPIHandler) TrackOrder(c *gin.Context) {
	orderNumber := c.Query("order_number")

	order, err := h.service(c).TrackOrder(orderNumber)
	if err != nil {
		if err.Error() == "order_number is required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_number": order.OrderNumber,
		"status":       order.Status,
		"created_at":   order.CreatedAt,
		"total_amount": order.Total,
		"items":        order.Items,
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
	Total           float64         `json:"total" binding:"required"`
	DeliveryMethod  string          `json:"delivery_method"`
	DeliveryAddress string          `json:"delivery_address"`
	Items           []CartActionReq `json:"items" binding:"required"`
}

func (h *StorefrontAPIHandler) VerifyPaystackCheckout(c *gin.Context) {
	var req CheckoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	sessionID := c.GetHeader("X-Session-ID")

	if req.PaymentMethod != "pickup" {
		if req.Reference == "" {
			c.JSON(400, gin.H{"error": "Payment reference is required"})
			return
		}
		// 1. Verify with Paystack API
		url := fmt.Sprintf("https://api.paystack.co/transaction/verify/%s", req.Reference)
		reqHttp, err := http.NewRequest("GET", url, nil)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create request"})
			return
		}
		reqHttp.Header.Set("Authorization", "Bearer "+h.paystackCfg.SecretKey)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(reqHttp)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to connect to payment gateway"})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var paystackResp struct {
			Status bool `json:"status"`
			Data   struct {
				Status string  `json:"status"`
				Amount float64 `json:"amount"` // in kobo/lowest denomination
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &paystackResp); err != nil {
			c.JSON(500, gin.H{"error": "Invalid response from payment gateway"})
			return
		}

		if !paystackResp.Status || paystackResp.Data.Status != "success" {
			c.JSON(400, gin.H{"error": "Payment verification failed"})
			return
		}

		// Ensure amount matches (ignoring currency conversion complexity for this mock, just checking total * 100)
		expectedAmount := req.Total * 100
		if paystackResp.Data.Amount < expectedAmount-1 {
			// -1 to account for floating point rounding issues
			c.JSON(400, gin.H{"error": "Payment amount mismatch"})
			return
		}
	}

	// 2. Process Order
	err := getDB(c, h.db).Transaction(func(tx *gorm.DB) error {
		var customerID *uuid.UUID
		if req.CustomerID != "" {
			parsed, _ := uuid.Parse(req.CustomerID)
			customerID = &parsed
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
		}

		orderType := "online"
		if req.DeliveryMethod != "" {
			orderType = req.DeliveryMethod
		}

		notes := "Paystack Reference: " + req.Reference
		if req.DeliveryMethod == "delivery" {
			notes += "\nDelivery Address: " + req.DeliveryAddress
		}

		order := models.Order{
			OrderNumber:   generateOrderNumber(),
			CustomerID:    customerID,
			Subtotal:      calculatedTotal,
			Total:         calculatedTotal, // Add tax/delivery later if needed
			Status:        "pending",
			PaymentStatus: "paid", // Mark as paid!
			PaymentMethod: "paystack",
			OrderType:     orderType,
			Notes:         &notes,
			ReceiptToken:  generateReceiptToken(),
			Items:         orderItems,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		if sessionID != "" {
			tx.Model(&models.AbandonedCart{}).Where("email = ? OR email = ?", sessionID, req.CustomerID).Update("is_recovered", true)
		}

		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Checkout failed: " + err.Error()})
		return
	}

	c.JSON(201, gin.H{"status": "checkout_successful", "reference": req.Reference})
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
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("ORD-%d-%s", time.Now().Unix(), hex.EncodeToString(b))
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
