export interface Branch {
  status?: string;
  id: string;
  tenant_id: string;
  name: string;
  unique_id?: string | null;
  address?: string | null;
  latitude?: number;
  longitude?: number;
  phone?: string | null;
  primary_color: string;
  logo?: string | null;
  low_stock_threshold: number;
  currency_symbol: string;
  currency_code: string;
  receipt_header?: string | null;
  receipt_footer?: string | null;
  branch_type: string;
  last_sync_at?: string;
  sync_status: string;
  pending_sync_count: number;
  sync_error_message?: string | null;
  created_at: string;
  updated_at: string;
  users?: any[];
}

export interface CashDrawerSession {
  branch_id: string;
  user_id: string;
  opening_balance: number;
  closing_balance: number;
  opened_at: string;
  closed_at?: string;
  notes?: string | null;
}

export interface Shift {
  branch_id: string;
  user_id: string;
  start_time: string;
  end_time?: string;
}

export interface PrintJob {
  branch_id: string;
  document_type: string;
  content: string;
  status: string;
  printed_at?: string;
}

