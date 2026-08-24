export interface ServiceCategory {
  name: string;
  icon: string;
  color: string;
}

export interface Service {
  category_id?: string;
  name: string;
  description?: string | null;
  duration_minutes: number;
  price: number;
  default_staff_id?: string;
  image?: string | null;
  is_active: boolean;
}

export interface Appointment {
  customer_id?: string;
  customer_name?: string | null;
  customer_phone?: string | null;
  customer_email?: string | null;
  service_id: string;
  staff_member_id?: string;
  start_time: string;
  end_time: string;
  status: string;
  notes?: string | null;
  order_id?: string;
}

export interface ServiceCommissionRule {
  staff_member_id: string;
  commission_type: string;
  value: number;
  applies_to: string;
  is_active: boolean;
}

export interface ServiceCommissionRecord {
  staff_member_id: string;
  rule_id?: string;
  order_id: string;
  amount: number;
  is_paid: boolean;
  paid_at?: string;
  notes?: string | null;
}

