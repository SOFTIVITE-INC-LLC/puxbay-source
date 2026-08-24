package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Tenant represents a merchant/company using the platform.
// Maps from Django: accounts.models.Tenant
type Tenant struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	Subdomain     string         `gorm:"size:100;uniqueIndex;not null" json:"subdomain"`
	SchemaName    string         `gorm:"size:100;uniqueIndex" json:"schema_name"`
	TenantType    string         `gorm:"size:20;default:'standard'" json:"tenant_type"`
	Logo          *string        `gorm:"size:512" json:"logo,omitempty"`
	Address       string         `gorm:"type:text" json:"address,omitempty"` // Encrypted at application layer
	PosAPIKey     string         `gorm:"type:text" json:"-"`                 // Encrypted, never exposed in API
	IsSandbox     bool           `gorm:"default:false" json:"is_sandbox"`
	SandboxWipeAt *time.Time     `json:"sandbox_wipe_at,omitempty"`
	HasUsedTrial  bool           `gorm:"default:false" json:"has_used_trial"`
	ReferralCode  *string        `gorm:"size:20;uniqueIndex" json:"referral_code,omitempty"`
	ReferredByID  *uuid.UUID     `gorm:"type:uuid" json:"referred_by_id,omitempty"`
	Metadata      datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedOn     time.Time      `gorm:"autoCreateTime" json:"created_on"`
	Status        string         `gorm:"size:20;default:'active'" json:"status"`

	// Relations
	// Relations
	ReferredBy   *Tenant       `gorm:"foreignKey:ReferredByID" json:"-"`
	Domains      []Domain      `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE" json:"domains,omitempty"`
	Subscription *Subscription `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE" json:"subscription,omitempty"`
}

func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.SchemaName == "" && t.Subdomain != "" {
		t.SchemaName = t.Subdomain
	}
	if t.ReferralCode == nil {
		code := generateReferralCode(8)
		t.ReferralCode = &code
	}
	return nil
}

// TableName explicitly sets the table name to public.tenants to allow cross-schema foreign keys.
func (Tenant) TableName() string {
	return "public.tenants"
}

// Domain represents a custom domain mapped to a tenant.
// Maps from Django: accounts.models.Domain
type Domain struct {
	ID                uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Domain            string     `gorm:"size:253;uniqueIndex;not null" json:"domain"`
	IsPrimary         bool       `gorm:"default:false" json:"is_primary"`
	IsVerified        bool       `gorm:"default:false" json:"is_verified"`
	VerificationToken uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid()" json:"-"`
	DNSCheckedAt      *time.Time `json:"dns_checked_at,omitempty"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

// SEOSettings for a tenant's public presence.
// Maps from Django: accounts.models.SEOSettings
type SEOSettings struct {
	Base
	TenantID          uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"tenant_id"`
	MetaTitle         *string   `gorm:"size:150" json:"meta_title,omitempty"`
	MetaDescription   *string   `gorm:"type:text" json:"meta_description,omitempty"`
	Keywords          *string   `gorm:"size:255" json:"keywords,omitempty"`
	OGTitle           *string   `gorm:"size:150" json:"og_title,omitempty"`
	OGDescription     *string   `gorm:"type:text" json:"og_description,omitempty"`
	OGImage           *string   `gorm:"size:512" json:"og_image,omitempty"`
	GoogleAnalyticsID *string   `gorm:"size:50" json:"google_analytics_id,omitempty"`
	FacebookPixelID   *string   `gorm:"size:50" json:"facebook_pixel_id,omitempty"`
	HomepageVideoID   *string   `gorm:"size:50;default:'dQw4w9WgXcQ'" json:"homepage_video_id,omitempty"`
	ContactEmail      *string   `gorm:"size:254" json:"contact_email,omitempty"`
	SupportEmail      *string   `gorm:"size:254" json:"support_email,omitempty"`
	ContactPhone      *string   `gorm:"size:50" json:"contact_phone,omitempty"`
	ContactAddress    *string   `gorm:"type:text" json:"contact_address,omitempty"`
	OfficeHours       *string   `gorm:"size:100" json:"office_hours,omitempty"`

	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

// TenantMetrics caches dashboard metrics for fast loading.
// Maps from Django: accounts.models.TenantMetrics
type TenantMetrics struct {
	Base
	TenantID          uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"tenant_id"`
	TotalProducts     uint      `gorm:"default:0" json:"total_products"`
	TotalOrders       uint      `gorm:"default:0" json:"total_orders"`
	TotalCustomers    uint      `gorm:"default:0" json:"total_customers"`
	TotalBranches     uint      `gorm:"default:0" json:"total_branches"`
	TotalRevenue      float64   `gorm:"type:decimal(15,2);default:0" json:"total_revenue"`
	TotalCustomerDebt float64   `gorm:"type:decimal(15,2);default:0" json:"total_customer_debt"`
	LastUpdated       time.Time `gorm:"autoUpdateTime" json:"last_updated"`

	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

// generateReferralCode creates a random alphanumeric code.
func generateReferralCode(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// Use uuid as entropy source for simplicity
		b[i] = chars[uuid.New()[0]%byte(len(chars))]
	}
	return string(b)
}
