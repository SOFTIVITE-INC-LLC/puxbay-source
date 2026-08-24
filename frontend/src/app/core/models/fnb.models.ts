export interface DiningTable {
  name: string;
  capacity: number;
  status: string;
  qr_code_url?: string | null;
  position_x: number;
  position_y: number;
  is_active: boolean;
}

export interface KDSTicket {
  order_id: string;
  table_id?: string;
  status: string;
  kitchen_notes?: string | null;
  is_rush: boolean;
  started_at?: string;
  completed_at?: string;
}

export interface SplitBillGroup {
  table_id?: string;
  original_order_id?: string;
  notes?: string | null;
}

