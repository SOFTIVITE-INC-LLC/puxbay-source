export interface PayrollPeriod {
  id: string;
  name: string;
  start_date: string;
  end_date: string;
  is_processed: boolean;
  processed_at?: string;
}

export interface PayrollRecord {
  id: string;
  period_id: string;
  staff_id: string;
  base_salary_snapshot: number;
  total_commission: number;
  bonus: number;
  deductions: number;
  net_pay: number;
  is_paid: boolean;
  paid_at?: string;
  payment_reference?: string | null;
}

export interface LeaveRequest {
  id: string;
  staff_id: string;
  leave_type: string;
  start_date: string;
  end_date: string;
  reason?: string | null;
  status: string;
  reviewed_by_id?: string;
  reviewed_at?: string;
}

export interface Attendance { 
  id: string;
  employee_id?: string;
  staff_id: string;
  clock_in: string;
  clock_out?: string;
  metadata?: any;
  status: string;
}


export interface Staff { id: string; user_id: string; branch_id: string; role: string; first_name?: string; last_name?: string; pin_code?: string; is_active?: boolean; phone?: string; [key: string]: any; }
