export interface StorefrontSettings {
  default_branch_id?: string;
  is_active: boolean;
  slug?: string;
  store_view_type: string;
  store_name?: string;
  banner_image?: string;
  logo_image?: string;
  primary_color: string;
  welcome_message?: string;
  about_text?: string;
  allow_pickup: boolean;
  allow_delivery: boolean;
  delivery_fee: number;
  min_order_amount: number;
  enable_stripe: boolean;
  enable_paystack: boolean;
  enable_mobile_money: boolean;
}

export interface ProductReview {
  product_id: string;
  customer_id: string;
  rating: number;
  comment: string;
  is_visible: boolean;
}

export interface Wishlist {
  customer_id: string;
  product_id: string;
}

export interface Coupon {
  code: string;
  discount_type: string;
  value: number;
  min_purchase: number;
  is_active: boolean;
  valid_from: string;
  valid_to: string;
  usage_limit: number;
  used_count: number;
}

export interface NewsletterSubscription {
  email: string;
  is_active: boolean;
}

export interface AbandonedCart {
  email: string;
  cart_data: any;
  is_recovered: boolean;
  email_sent: boolean;
}

export interface ProductImageGallery {
  product_id: string;
  image_url: string;
  alt_text?: string | null;
  order: number;
}

