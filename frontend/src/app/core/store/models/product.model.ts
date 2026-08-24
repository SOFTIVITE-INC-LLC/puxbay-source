export interface Category {
  id: string;
  name: string;
  image?: string;
  color?: string;
}

export interface Product {
  id: string;
  name: string;
  description?: string;
  sku: string;
  barcode?: string;
  cost_price: number;
  selling_price: number;
  stock_quantity: number;
  low_stock_threshold: number;
  category_id?: string;
  supplier_id?: string;
  branch_id?: string;
  tax_rate: number;
  is_active: boolean;
  image_url?: string;
  images?: string[];
  category?: { id: string; name: string };
  created_at?: string;
  updated_at?: string;
}

export interface ProductReview {
  id: string;
  product_id: string;
  customer_id: string;
  rating: number;
  comment: string;
  is_visible: boolean;
  created_at?: string;
}

export interface ProductDetail {
  product: Product;
  reviews: ProductReview[];
  avg_rating: number;
}

export interface ProductsResponse {
  products: Product[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ProductResponse {
  product: Product;
  images?: { image_url: string }[];
  reviews: ProductReview[];
  avg_rating: number;
  related_products?: Product[];
}
