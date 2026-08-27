package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// StorefrontSettings holds per-tenant online store configuration.
// TenantScoped = per-schema, no TenantID column needed.
type StorefrontSettings struct {
	TenantScoped
	DefaultBranchID        *uuid.UUID `gorm:"type:uuid" json:"default_branch_id,omitempty"`
	IsActive               bool       `gorm:"default:false" json:"is_active"`
	Slug                   string     `gorm:"size:100" json:"slug,omitempty"`
	StoreViewType          string     `gorm:"size:20;default:'branch'" json:"store_view_type"` // branch | global
	StoreName              string     `gorm:"size:100" json:"store_name,omitempty"`
	BannerImage            string     `gorm:"size:255" json:"banner_image,omitempty"`
	LogoImage              string     `gorm:"size:255" json:"logo_image,omitempty"`
	PrimaryColor           string     `gorm:"size:7;default:'#3b82f6'" json:"primary_color"`
	WelcomeMessage         string     `gorm:"type:text" json:"welcome_message,omitempty"`
	AboutText              string     `gorm:"type:text" json:"about_text,omitempty"`
	AllowPickup            bool       `gorm:"default:true" json:"allow_pickup"`
	AllowDelivery          bool       `gorm:"default:false" json:"allow_delivery"`
	DeliveryFee            float64    `gorm:"type:decimal(8,2);default:0" json:"delivery_fee"`
	MinOrderAmount         float64    `gorm:"type:decimal(10,2);default:0" json:"min_order_amount"`
	EnableStripe           bool       `gorm:"default:false" json:"enable_stripe"`
	EnablePaystack         bool       `gorm:"default:false" json:"enable_paystack"`
	PaystackPublicKey      string     `gorm:"size:255" json:"paystack_public_key,omitempty"`
	PaystackSubaccountCode string     `gorm:"size:100" json:"paystack_subaccount_code,omitempty"`
	EnableMobileMoney      bool       `gorm:"default:false" json:"enable_mobile_money"`
	Currency               string     `gorm:"size:10;default:'GHS'" json:"currency,omitempty"`
	CurrencySymbol         string     `gorm:"size:10;default:'GH₵'" json:"currency_symbol,omitempty"`
	FlashSaleEndTime       *time.Time `json:"flash_sale_end_time,omitempty"`
}

// ProductReview represents a customer rating and comment on a product.
type ProductReview struct {
	TenantScoped
	ProductID  uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`
	Rating     int       `gorm:"not null" json:"rating"` // 1-5
	Comment    string    `gorm:"type:text" json:"comment"`
	IsVisible  bool      `gorm:"default:true" json:"is_visible"`
}

// Wishlist represents a customer's saved product.
type Wishlist struct {
	Base
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
}

// Coupon represents a discount code for the storefront.
type Coupon struct {
	TenantScoped
	Code         string    `gorm:"size:50;not null;uniqueIndex" json:"code"`
	DiscountType string    `gorm:"size:20;not null" json:"discount_type"` // percentage | fixed
	Value        float64   `gorm:"type:decimal(10,2);not null" json:"value"`
	MinPurchase  float64   `gorm:"type:decimal(10,2);default:0" json:"min_purchase"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	ValidFrom    time.Time `json:"valid_from"`
	ValidTo      time.Time `json:"valid_to"`
	UsageLimit   int       `gorm:"default:100" json:"usage_limit"`
	UsedCount    int       `gorm:"default:0" json:"used_count"`
}

// NewsletterSubscription tracks email subscribers per tenant.
type NewsletterSubscription struct {
	TenantScoped
	Email    string `gorm:"size:255;not null" json:"email"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
}

// AbandonedCart stores session cart data for recovery emails.
type AbandonedCart struct {
	TenantScoped
	Email       string         `gorm:"size:255;not null" json:"email"`
	CartData    datatypes.JSON `gorm:"type:jsonb" json:"cart_data"`
	IsRecovered bool           `gorm:"default:false" json:"is_recovered"`
	EmailSent   bool           `gorm:"default:false" json:"email_sent"`
}

// ProductImageGallery holds extra images for a product.
type ProductImageGallery struct {
	Base
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	ImageURL  string    `gorm:"size:512;not null" json:"image_url"`
	AltText   *string   `gorm:"size:255" json:"alt_text,omitempty"`
	Order     uint      `gorm:"default:0" json:"order"`
}

// BackInStockSubscription tracks users who want to be notified when a product is restocked.
type BackInStockSubscription struct {
	TenantScoped
	ProductID  uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	Email      string    `gorm:"size:255;not null" json:"email"`
	IsNotified bool      `gorm:"default:false" json:"is_notified"`
}
