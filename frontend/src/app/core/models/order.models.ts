export interface Order { 
  id: string; 
  created_at?: string; 
  order_number?: string;
  customer_id?: string;
  customer_name?: string;
  item_count?: number;
  total?: number;
  status?: string;
  customer?: any;
  notes?: string;
  items?: any[];
  order_type?: string;
  payment_method?: string;
  subtotal?: number;
  tax?: number;
  discount?: number;
  amount_paid?: number;
}
export interface Return { 
  id: string; 
  order_id?: string;
  amount?: number;
  status: string;
  total_refund?: number;
  reason?: string;
  reason_detail?: string;
  created_at?: string;
}
