export interface Order { 
  id: string; 
  created_at?: string; 
  order_number?: string;
  customer_id?: string;
  customer_name?: string;
  customer_phone?: string;
  delivery_address?: string;
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
  payment_status?: string;
  is_otp_verified?: boolean;
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
