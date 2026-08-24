export interface TenantScoped {
  id: string;
  tenant_id: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface BranchScoped extends TenantScoped {
  branch_id?: string | null;
}

export interface Product extends BranchScoped {
  category_id?: string | null;
  name: string;
  slug: string;
  sku: string;
  barcode?: string | null;
  description?: string | null;
  cost_price: number;
  selling_price: number;
  wholesale_price?: number | null;
  tax_rate: number;
  track_inventory: boolean;
  current_stock: number;
  reorder_level: number;
  stock_unit: string;
  is_active: boolean;
  is_online: boolean;
  product_type: string;
  color?: string | null;
  brand?: string | null;
  weight?: number | null;
  image_url?: string | null;
  
  expiry_date?: string | null;
  manufacturing_date?: string | null;
  minimum_wholesale_quantity?: number;
  batch_number?: string | null;
  invoice_waybill_number?: string | null;
  country_of_origin?: string | null;
  manufacturer_name?: string | null;
  manufacturer_address?: string | null;

  category?: Category;
}

export interface Category extends TenantScoped {
  name: string;
  slug: string;
  description?: string | null;
  parent_id?: string | null;
  image_url?: string | null;
  is_active: boolean;
}

export interface Customer extends TenantScoped {
  last_visit?: string;
  name: string;
  email?: string | null;
  phone?: string | null;
  address?: string | null;
  loyalty_pts: number;
  store_credit: number;
  debt_balance: number;
  total_spend: number;
  order_count: number;
}

export interface CustomerTier extends TenantScoped {
  name: string;
  min_spend: number;
  point_multiplier: number;
  discount_percentage: number;
  color: string;
}

export interface Order extends BranchScoped {
  order_number: string;
  customer_id?: string | null;
  cashier_id?: string | null;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  amount_paid: number;
  status: string;
  payment_status: string;
  payment_method: string;
  order_type: string;
  notes?: string | null;
  receipt_token: string;
  item_count?: number;
  customer_name?: string;
  
  items?: OrderItem[];
}

export interface OrderItem {
  id: string;
  created_at: string;
  updated_at: string;
  order_id: string;
  product_id: string;
  variant_id?: string | null;
  quantity: number;
  unit_price: number;
  discount: number;
  total: number;
  cost_price: number;
  
  product?: Product;
}

export interface GiftCard extends TenantScoped {
  code: string;
  purchaser_id?: string | null;
  initial_balance: number;
  current_balance: number;
  is_active: boolean;
  expires_at?: string | null;
}

export interface StorefrontSettings extends TenantScoped {
  is_active: boolean;
  store_name: string;
  slug: string;
  primary_color: string;
  welcome_message?: string | null;
  about_text?: string | null;
  logo_url?: string | null;
  banner_url?: string | null;
  allow_pickup: boolean;
  allow_delivery: boolean;
  delivery_fee: number;
  min_order_amount: number;
  store_view_type: string;
  enable_paystack?: boolean;
  paystack_public_key?: string;
  paystack_subaccount_code?: string;
}
