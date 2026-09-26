export interface PayrollRecord {
  id: string;
  period_id: string;
  staff_id: string;
  staff_name?: string;
  staff_role?: string;
  hours_worked?: number;
  sales_amount?: number;
  base_salary_snapshot: number;
  total_commission: number;
  bonus: number;
  deductions: number;
  net_pay: number;
  is_paid: boolean;
  paid_at?: string;
  payment_method?: string;
  payment_reference?: string | null;
  notes?: string;
}

export interface PayrollPeriod {
  id: string;
  name: string;
  start_date: string;
  end_date: string;
  is_processed: boolean;
  processed_at?: string;
  total_gross?: number;
  total_net?: number;
  staff_count?: number;
  records?: PayrollRecord[];
}

export interface LeaveRequest {
  id: string;
  staff_id: string;
  leave_type: string;
  start_date: string;
  end_date: string;
  days_count?: number;
  reason?: string | null;
  status: 'pending' | 'approved' | 'rejected';
  reviewed_by_id?: string;
  reviewed_at?: string;
  manager_notes?: string;
}

export interface LeaveSummary {
  total_quota: number;
  used_days: number;
  pending_days: number;
  remaining_days: number;
}

export interface Attendance { 
  id: string;
  employee_id?: string;
  staff_id: string;
  clock_in: string;
  clock_out?: string;
  hours_worked?: number;
  metadata?: any;
  status: 'present' | 'completed' | 'on_leave' | 'absent' | string;
  notes?: string;
}

export interface CommissionRule {
  id?: string;
  branch_id?: string | null;
  name: string;
  rule_type?: 'percentage' | 'flat' | 'tiered';
  min_sales_amount: number;
  commission_percentage: number;
  flat_bonus: number;
  is_active: boolean;
}

export interface StaffAchievement {
  id?: string;
  staff_id: string;
  badge_name: string;
  badge_icon: string;
  description?: string;
  awarded_at?: string;
}

export interface ShiftSwapRequest {
  id?: string;
  requesting_staff_id: string;
  target_staff_id?: string;
  original_shift_id: string;
  status: 'pending' | 'approved' | 'rejected' | 'completed';
  notes?: string;
}

export interface StaffShift {
  id?: string;
  staff_id: string;
  start_time: string;
  end_time: string;
  role?: string;
  status?: 'scheduled' | 'in_progress' | 'completed' | 'missed';
  notes?: string;
}

export interface Staff {
  id: string;
  user_id: string;
  branch_id?: string;
  role: string;
  first_name?: string;
  last_name?: string;
  email?: string;
  phone?: string;
  department?: string;
  job_title?: string;
  employment_type?: 'full_time' | 'part_time' | 'contract';
  base_salary?: number;
  hourly_rate?: number;
  payment_method?: string;
  bank_details?: any;
  pin_code?: string;
  is_active?: boolean;
  [key: string]: any;
}
