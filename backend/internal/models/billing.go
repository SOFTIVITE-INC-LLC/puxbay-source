package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Plan represents a subscription plan.
// Maps from Django: billing.models.Plan
type Plan struct {
	Base
	Name              string         `gorm:"size:50;not null" json:"name"`
	Description       string         `gorm:"type:text" json:"description"`
	Price             float64        `gorm:"type:decimal(10,2);default:0" json:"price"`
	PriceGHS          float64        `gorm:"type:decimal(10,2);default:0" json:"price_ghs"`
	Interval          string         `gorm:"size:20;default:'monthly'" json:"interval"` // monthly, 6-month, yearly
	TrialDays         uint           `gorm:"default:0" json:"trial_days"`
	PaystackPlanCode  *string        `gorm:"size:100" json:"paystack_plan_code,omitempty"`
	MaxBranches       uint           `gorm:"default:1" json:"max_branches"`
	MaxUsers          uint           `gorm:"default:1" json:"max_users"`
	APIAccess         bool           `gorm:"default:false" json:"api_access"`
	APIDailyLimit     uint           `gorm:"default:0" json:"api_daily_limit"`
	IsCustom          bool           `gorm:"default:false" json:"is_custom"`
	PricePerBranch    float64        `gorm:"type:decimal(10,2);default:0" json:"price_per_branch"`
	PricePerUser      float64        `gorm:"type:decimal(10,2);default:0" json:"price_per_user"`
	PricePerBranchGHS float64        `gorm:"type:decimal(10,2);default:0" json:"price_per_branch_ghs"`
	PricePerUserGHS   float64        `gorm:"type:decimal(10,2);default:0" json:"price_per_user_ghs"`
	Features          datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"features"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
}

func (Plan) TableName() string {
	return "public.plans"
}

// Subscription links a tenant to a plan.
// Maps from Django: billing.models.Subscription
type Subscription struct {
	Base
	TenantID                 uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"tenant_id"`
	PlanID                   *uuid.UUID `gorm:"type:uuid" json:"plan_id,omitempty"`
	PaystackSubscriptionCode *string    `gorm:"size:100" json:"paystack_subscription_code,omitempty"`
	PaystackCustomerCode     *string    `gorm:"size:100" json:"paystack_customer_code,omitempty"`
	Status                   string     `gorm:"size:20;default:'trialing'" json:"status"` // active, past_due, trialing, canceled, incomplete
	CurrentPeriodEnd         *time.Time `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd        bool       `gorm:"default:false" json:"cancel_at_period_end"`
	APIRequestsToday         uint       `gorm:"default:0" json:"api_requests_today"`
	APIRequestsMonth         uint       `gorm:"default:0" json:"api_requests_this_month"`
	APILastResetDate         time.Time  `gorm:"autoCreateTime" json:"api_last_reset_date"`
	APIMonthResetDate        time.Time  `gorm:"autoCreateTime" json:"api_month_reset_date"`
	CustomBranchesCount      *uint      `json:"custom_branches_count,omitempty"`
	CustomUsersCount         *uint      `json:"custom_users_count,omitempty"`
	LastBillingEmailAt       *time.Time `json:"last_billing_email_at,omitempty"`

	// Relations
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Plan   *Plan  `gorm:"-" json:"plan,omitempty"`
}

func (Subscription) TableName() string {
	return "public.subscriptions"
}

// BillingPayment records subscription payments.
// Maps from Django: billing.models.Payment
type BillingPayment struct {
	Base
	SubscriptionID    uuid.UUID `gorm:"type:uuid;not null;index" json:"subscription_id"`
	Amount            float64   `gorm:"type:decimal(10,2)" json:"amount"`
	PaystackReference *string   `gorm:"size:100" json:"paystack_reference,omitempty"`
	Status            string    `gorm:"size:20;default:'succeeded'" json:"status"`
	Date              time.Time `gorm:"autoCreateTime" json:"date"`

	Subscription Subscription `gorm:"foreignKey:SubscriptionID" json:"subscription,omitempty"`
}

func (BillingPayment) TableName() string {
	return "public.billing_payments"
}

// PromoCode for subscription discounts.
// Maps from Django: billing.models.PromoCode
type PromoCode struct {
	Base
	Code          string     `gorm:"size:50;uniqueIndex;not null" json:"code"`
	DiscountType  string     `gorm:"size:20;default:'percentage'" json:"discount_type"` // percentage, fixed
	DiscountValue float64    `gorm:"type:decimal(10,2)" json:"discount_value"`
	MaxUses       uint       `gorm:"default:0" json:"max_uses"` // 0 = unlimited
	CurrentUses   uint       `gorm:"default:0" json:"current_uses"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	ValidFrom     time.Time  `gorm:"autoCreateTime" json:"valid_from"`
	ValidUntil    *time.Time `json:"valid_until,omitempty"`
}

func (PromoCode) TableName() string {
	return "public.promo_codes"
}

// ReferralReward tracks rewards from tenant referrals.
type ReferralReward struct {
	Base
	ReferrerID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"referrer_id"`
	ReferredTenantID uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"referred_tenant_id"`
	RewardAmount     float64    `gorm:"type:decimal(10,2)" json:"reward_amount"`
	IsApplied        bool       `gorm:"default:false" json:"is_applied"`
	AppliedAt        *time.Time `json:"applied_at,omitempty"`

	Referrer       Tenant `gorm:"foreignKey:ReferrerID" json:"-"`
	ReferredTenant Tenant `gorm:"foreignKey:ReferredTenantID" json:"-"`
}

func (ReferralReward) TableName() string {
	return "public.referral_rewards"
}

// BillingSettings — global billing settings.
type BillingSettings struct {
	ID                uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	ReferralRewardGHS float64 `gorm:"type:decimal(10,2);default:10.00" json:"referral_reward_ghs"`
}

func (BillingSettings) TableName() string {
	return "public.billing_settings"
}

func (b *BillingSettings) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// PricingPlan is for the public marketing site.
type PricingPlan struct {
	Base
	Name         string        `gorm:"size:100" json:"name"`
	Slug         string        `gorm:"size:50;uniqueIndex" json:"slug"`
	PriceMonthly float64       `gorm:"type:decimal(10,2)" json:"price_monthly"`
	PriceYearly  float64       `gorm:"type:decimal(10,2)" json:"price_yearly"`
	Currency     string        `gorm:"size:3;default:'USD'" json:"currency"`
	Description  string        `gorm:"type:text" json:"description"`
	IsPopular    bool          `gorm:"default:false" json:"is_popular"`
	ButtonText   string        `gorm:"size:50;default:'Get Started'" json:"button_text"`
	OrderIndex   uint          `gorm:"default:0" json:"order_index"`
	MaxBranches  int           `gorm:"default:1" json:"max_branches"`
	MaxStaff     int           `gorm:"default:1" json:"max_staff"`
	Features     []PlanFeature `gorm:"foreignKey:PlanID" json:"features"`
}

func (PricingPlan) TableName() string {
	return "public.pricing_plans"
}

// PlanFeature maps features to PricingPlans.
type PlanFeature struct {
	Base
	PlanID      uuid.UUID `gorm:"type:uuid;not null;index" json:"plan_id"`
	Text        string    `gorm:"size:255" json:"text"`
	IsAvailable bool      `gorm:"default:true" json:"is_available"`
	OrderIndex  uint      `gorm:"default:0" json:"order_index"`
}

func (PlanFeature) TableName() string {
	return "public.plan_features"
}

// PaymentGatewayConfig stores global config for Stripe, Paystack etc.
type PaymentGatewayConfig struct {
	Base
	Name        string `gorm:"size:50" json:"name"`
	Slug        string `gorm:"size:50;uniqueIndex" json:"slug"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	Description string `gorm:"size:255" json:"description"`
}

func (PaymentGatewayConfig) TableName() string {
	return "public.payment_gateway_configs"
}
