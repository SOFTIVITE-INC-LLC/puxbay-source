export interface Category { id: string; [key: string]: any; }
export interface Product { 
  id: string; 
  name: string; 
  description?: string; 
  sku?: string; 
  image_url?: string; 
  slug?: string;
  cost_price?: number;
  selling_price?: number;
  tax_rate?: number;
  product_type?: string;
  tenant_id?: string;
  branch_id?: string;
  is_active?: boolean;
  track_inventory?: boolean;
  current_stock?: number;
  reorder_level?: number;
  stock_unit?: string;
  is_online?: boolean;
  created_at?: string;
  updated_at?: string;
  [key: string]: any; 
}
