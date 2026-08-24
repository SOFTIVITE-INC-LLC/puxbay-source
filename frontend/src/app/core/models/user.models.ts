export interface User {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  is_active: boolean;
  is_superuser: boolean;
  is_staff: boolean;
  last_login?: string;
  date_joined: string;
  profiles?: UserProfile[];
}

export interface UserProfile {
  id: string;
  user_id: string;
  tenant_id: string;
  branch_id?: string;
  role: string;
  can_perform_credit_sales: boolean;
  base_salary: number;
  hourly_rate: number;
  payment_method: string;
  user?: User;
}

