export interface Plan {
  name: string;
  description: string;
  price: number;
  price_ghs: number;
  interval: string;
  trial_days: number;
  stripe_price_id?: string | null;
  max_branches: number;
  max_users: number;
  api_access: boolean;
  api_daily_limit: number;
  is_custom: boolean;
  price_per_branch: number;
  price_per_user: number;
  price_per_branch_ghs: number;
  price_per_user_ghs: number;
}

export interface Subscription {
  tenant_id: string;
  plan_id?: string;
  stripe_subscription_id?: string | null;
  stripe_customer_id?: string | null;
  status: string;
  current_period_end?: string;
  cancel_at_period_end: boolean;
  api_requests_today: number;
  api_requests_this_month: number;
  api_last_reset_date: string;
  api_month_reset_date: string;
  custom_branches_count?: number;
  custom_users_count?: number;
  plan?: Plan | null;
}

export interface BillingPayment {
  subscription_id: string;
  amount: number;
  receipt_url?: string | null;
  status: string;
  date: string;
}

export interface PromoCode {
  code: string;
  discount_type: string;
  discount_value: number;
  max_uses: number;
  current_uses: number;
  is_active: boolean;
  valid_from: string;
  valid_until?: string;
}

export interface ReferralReward {
  referrer_id: string;
  referred_tenant_id: string;
  reward_amount: number;
  is_applied: boolean;
  applied_at?: string;
}

export interface BillingSettings {
  id: number;
  referral_reward_ghs: number;
}

export interface PricingPlan {
  name: string;
  slug: string;
  price_monthly: number;
  price_yearly: number;
  currency: string;
  description: string;
  is_popular: boolean;
  button_text: string;
  order_index: number;
}

export interface PlanFeature {
  plan_id: string;
  text: string;
  is_available: boolean;
  order_index: number;
}

export interface PaymentGatewayConfig {
  name: string;
  slug: string;
  is_active: boolean;
  description: string;
}

