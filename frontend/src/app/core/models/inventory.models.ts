export interface StockTransfer {
  reference_no: string;
  from_branch_id: string;
  to_branch_id: string;
  status: string;
  notes?: string | null;
  created_by_id?: string;
  shipped_at?: string;
  received_at?: string;
  items?: StockTransferItem[];
}

export interface StockTransferItem {
  transfer_id: string;
  product_id: string;
  variant_id?: string;
  quantity: number;
}

export interface PurchaseOrder {
  id: string;
  po_number: string;
  supplier_id: string;
  status: string;
  total_amount: number;
  expected_date?: string;
  notes?: string | null;
  items?: PurchaseOrderItem[];
}

export interface PurchaseOrderItem {
  id: string;
  po_id: string;
  product_id: string;
  variant_id?: string;
  quantity_ordered: number;
  quantity_received: number;
  unit_cost: number;
}

export interface StocktakeSession {
  name: string;
  status: string;
  notes?: string | null;
  created_by_id?: string;
  completed_at?: string;
}

export interface StockMovement {
  tenant_id: string;
  branch_id: string;
  product_id: string;
  variant_id?: string;
  quantity: number;
  previous_stock: number;
  new_stock: number;
  reason: string;
  reference_id?: string | null;
  user_id?: string;
}

export interface StockBatch {
  id?: string;
  branch_id: string;
  product_id: string;
  product?: { id: string; name: string; sku: string; };
  batch_number: string;
  quantity: number;
  expiry_date?: string;
  manufacture_date?: string;
  created_at?: string;
}

export interface StocktakeEntry {
  session_id: string;
  product_id: string;
  variant_id?: string;
  expected_stock: number;
  actual_stock: number;
  difference: number;
}

export interface InventoryRecommendation {
  branch_id: string;
  product_id: string;
  recommended_stock: number;
  reason: string;
  is_applied: boolean;
}

export interface StockAlert {
  branch_id: string;
  product_id: string;
  message: string;
  is_resolved: boolean;
}

