export interface CustomerTier {
  name: string;
  min_spend: number;
  discount_percentage: number;
  color: string;
  icon: string;
}

export interface Customer {
  name: string;
  phone?: string | null;
  email?: string | null;
  address?: string | null;
  tier_id?: string;
  total_spend: number;
  order_count: number;
  loyalty_points: number;
  store_credit: number;
  debt_balance: number;
  accepts_marketing: boolean;
  last_visit?: string;
  date_of_birth?: string;
  notes?: string | null;
  tier?: CustomerTier | null;
}

export interface LoyaltyTransaction {
  tenant_id: string;
  customer_id: string;
  order_id?: string;
  points: number;
  transaction_type: string;
  description?: string | null;
}

export interface GiftCard {
  code: string;
  initial_balance: number;
  current_balance: number;
  purchaser_id?: string;
  recipient_email?: string | null;
  expires_at?: string;
  is_active: boolean;
}

export interface StoreCreditTransaction {
  customer_id: string;
  amount: number;
  transaction_type: string;
  reference?: string;
  notes?: string;
  created_by_id?: string;
}

