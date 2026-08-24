package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type StorefrontService struct {
	db       *gorm.DB
	redis    *redis.Client
	tenantID uuid.UUID
}

func NewStorefrontService(db *gorm.DB, rdb *redis.Client, tenantID uuid.UUID) *StorefrontService {
	return &StorefrontService{db: db, redis: rdb, tenantID: tenantID}
}

func (s *StorefrontService) GetSettings() (*models.StorefrontSettings, error) {
	cacheKey := fmt.Sprintf("storefront:settings:%s", s.tenantID.String())
	if s.redis != nil {
		if val, err := s.redis.Get(context.Background(), cacheKey).Result(); err == nil {
			var settings models.StorefrontSettings
			if err := json.Unmarshal([]byte(val), &settings); err == nil {
				return &settings, nil
			}
		}
	}

	var settings models.StorefrontSettings
	if err := s.db.First(&settings).Error; err != nil {
		return &models.StorefrontSettings{
			IsActive:     false,
			StoreName:    "",
			PrimaryColor: "#3b82f6",
			AllowPickup:  true,
		}, nil
	}

	if s.redis != nil {
		if data, err := json.Marshal(settings); err == nil {
			s.redis.Set(context.Background(), cacheKey, data, 5*time.Minute)
		}
	}

	return &settings, nil
}

func (s *StorefrontService) UpdateSettings(req *models.StorefrontSettings) (*models.StorefrontSettings, error) {
	var settings models.StorefrontSettings
	if err := s.db.First(&settings).Error; err != nil {
		if err := s.db.Create(req).Error; err != nil {
			return nil, err
		}
		return req, nil
	}

	settings.IsActive = req.IsActive
	settings.StoreName = req.StoreName
	settings.PrimaryColor = req.PrimaryColor
	settings.WelcomeMessage = req.WelcomeMessage
	settings.AboutText = req.AboutText
	settings.AllowPickup = req.AllowPickup
	settings.AllowDelivery = req.AllowDelivery
	settings.DeliveryFee = req.DeliveryFee
	settings.MinOrderAmount = req.MinOrderAmount
	settings.StoreViewType = req.StoreViewType
	settings.Slug = req.Slug
	settings.FlashSaleEndTime = req.FlashSaleEndTime

	if err := s.db.Save(&settings).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

type ProductSearchParams struct {
	BranchID   string
	Search     string
	CategoryID string
	MinPrice   string
	MaxPrice   string
	SortBy     string
	Page       int
	PageSize   int
	InStock    string
}

type ProductSearchResult struct {
	Products   []models.Product
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

func (s *StorefrontService) SearchProducts(params ProductSearchParams) (*ProductSearchResult, error) {
	cacheKey := fmt.Sprintf("storefront:products:%s:%+v", s.tenantID.String(), params)
	if s.redis != nil {
		if val, err := s.redis.Get(context.Background(), cacheKey).Result(); err == nil {
			var result ProductSearchResult
			if err := json.Unmarshal([]byte(val), &result); err == nil {
				return &result, nil
			}
		}
	}

	offset := (params.Page - 1) * params.PageSize

	query := s.db.Model(&models.Product{}).Where("is_active = ? AND is_online = ?", true, true)

	if params.BranchID != "" {
		query = query.Where("branch_id = ?", params.BranchID)
	}
	if params.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}
	if params.CategoryID != "" {
		query = query.Where("category_id = ?", params.CategoryID)
	}
	if params.MinPrice != "" {
		query = query.Where("selling_price >= ?", params.MinPrice)
	}
	if params.MaxPrice != "" {
		query = query.Where("selling_price <= ?", params.MaxPrice)
	}
	if params.InStock == "true" {
		query = query.Where("current_stock > 0")
	}

	switch params.SortBy {
	case "price_low":
		query = query.Order("selling_price ASC")
	case "price_high":
		query = query.Order("selling_price DESC")
	default:
		query = query.Order("created_at DESC")
	}

	var total int64
	query.Count(&total)

	var products []models.Product
	if err := query.Preload("Category").Limit(params.PageSize).Offset(offset).Find(&products).Error; err != nil {
		return nil, err
	}

	result := &ProductSearchResult{
		Products:   products,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: (int(total) + params.PageSize - 1) / params.PageSize,
	}

	if s.redis != nil {
		if data, err := json.Marshal(result); err == nil {
			s.redis.Set(context.Background(), cacheKey, data, 5*time.Minute)
		}
	}

	return result, nil
}

func (s *StorefrontService) ListCategories(branchID string) ([]models.Category, error) {
	var categories []models.Category

	query := s.db.Order("name ASC")

	if branchID != "" {
		// Only fetch categories that have active online products in this branch
		query = query.Where("id IN (SELECT category_id FROM products WHERE branch_id = ? AND is_active = true AND is_online = true AND category_id IS NOT NULL)", branchID)
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

type ProductDetail struct {
	Product         models.Product
	Images          []models.ProductImageGallery
	Reviews         []models.ProductReview
	AvgRating       float64
	RelatedProducts []models.Product
}

func (s *StorefrontService) GetProduct(id string) (*ProductDetail, error) {
	productID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid product ID")
	}

	var product models.Product
	if err := s.db.Where("id = ? AND is_active = ? AND is_online = ?", productID, true, true).
		Preload("Category").First(&product).Error; err != nil {
		return nil, errors.New("product not found")
	}

	var reviews []models.ProductReview
	s.db.Where("product_id = ? AND is_visible = ?", productID, true).
		Order("created_at DESC").Limit(10).Find(&reviews)

	type AvgResult struct {
		AvgRating float64
	}
	var avgResult AvgResult
	s.db.Model(&models.ProductReview{}).
		Select("AVG(rating) as avg_rating").
		Where("product_id = ? AND is_visible = ?", productID, true).
		Scan(&avgResult)

	var images []models.ProductImageGallery
	s.db.Where("product_id = ?", productID).Order("\"order\" ASC").Find(&images)

	var relatedProducts []models.Product
	if product.CategoryID != nil {
		s.db.Where("category_id = ? AND id != ? AND is_active = ? AND is_online = ?",
			product.CategoryID, product.ID, true, true).
			Limit(4).Find(&relatedProducts)
	}

	return &ProductDetail{
		Product:         product,
		Images:          images,
		Reviews:         reviews,
		AvgRating:       avgResult.AvgRating,
		RelatedProducts: relatedProducts,
	}, nil
}

func (s *StorefrontService) TrackOrder(orderNumber string) (*models.Order, error) {
	if orderNumber == "" {
		return nil, errors.New("order_number is required")
	}

	var order models.Order
	if err := s.db.Where("order_number = ?", orderNumber).
		Preload("Items").First(&order).Error; err != nil {
		return nil, errors.New("order not found")
	}

	return &order, nil
}

type ReviewInput struct {
	CustomerID string
	Rating     int
	Comment    string
}

func (s *StorefrontService) SubmitReview(productIDStr string, input ReviewInput) (*models.ProductReview, error) {
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return nil, errors.New("invalid product ID")
	}

	customerID, err := uuid.Parse(input.CustomerID)
	if err != nil {
		return nil, errors.New("invalid customer ID")
	}

	review := models.ProductReview{
		ProductID:  productID,
		CustomerID: customerID,
		Rating:     input.Rating,
		Comment:    input.Comment,
		IsVisible:  true,
	}

	if err := s.db.Create(&review).Error; err != nil {
		return nil, err
	}

	return &review, nil
}

func (s *StorefrontService) SubscribeNewsletter(email string) error {
	var sub models.NewsletterSubscription
	if err := s.db.Where("email = ?", email).Assign(models.NewsletterSubscription{IsActive: true}).FirstOrCreate(&sub).Error; err != nil {
		return err
	}
	return nil
}

func (s *StorefrontService) SubscribeBackInStock(productIDStr string, email string) error {
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return errors.New("invalid product ID")
	}
	var sub models.BackInStockSubscription
	if err := s.db.Where("email = ? AND product_id = ?", email, productID).FirstOrCreate(&sub, models.BackInStockSubscription{
		ProductID: productID,
		Email:     email,
	}).Error; err != nil {
		return err
	}
	return nil
}

type CouponValidationResult struct {
	Coupon         models.Coupon
	DiscountAmount float64
	NewTotal       float64
}

func (s *StorefrontService) ApplyCoupon(code string, cartTotal float64) (*CouponValidationResult, error) {
	var coupon models.Coupon
	if err := s.db.Where("code = ? AND is_active = ?", code, true).First(&coupon).Error; err != nil {
		return nil, errors.New("invalid or expired coupon code")
	}

	if cartTotal < coupon.MinPurchase {
		return nil, errors.New("minimum purchase amount not met")
	}

	var discountAmount float64
	if coupon.DiscountType == "percentage" {
		discountAmount = cartTotal * (coupon.Value / 100)
	} else {
		discountAmount = coupon.Value
	}
	if discountAmount > cartTotal {
		discountAmount = cartTotal
	}

	return &CouponValidationResult{
		Coupon:         coupon,
		DiscountAmount: discountAmount,
		NewTotal:       cartTotal - discountAmount,
	}, nil
}

func (s *StorefrontService) ListCoupons() ([]models.Coupon, error) {
	var coupons []models.Coupon
	if err := s.db.Order("created_at DESC").Find(&coupons).Error; err != nil {
		return nil, err
	}
	return coupons, nil
}

func (s *StorefrontService) CreateCoupon(coupon *models.Coupon) error {
	return s.db.Create(coupon).Error
}

func (s *StorefrontService) UpdateCoupon(id string, input models.Coupon) (*models.Coupon, error) {
	couponID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid coupon ID")
	}

	var coupon models.Coupon
	if err := s.db.Where("id = ?", couponID).First(&coupon).Error; err != nil {
		return nil, errors.New("coupon not found")
	}

	coupon.Code = input.Code
	coupon.DiscountType = input.DiscountType
	coupon.Value = input.Value
	coupon.MinPurchase = input.MinPurchase
	coupon.ValidFrom = input.ValidFrom
	coupon.ValidTo = input.ValidTo
	coupon.IsActive = input.IsActive

	if err := s.db.Save(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}
