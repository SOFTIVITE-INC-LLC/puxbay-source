export interface Supplier {
  name: string;
  contact_person?: string | null;
  email?: string | null;
  phone?: string | null;
  address?: string | null;
  tax_number?: string | null;
  payment_terms?: string | null;
  notes?: string | null;
  credit_balance: number;
  is_active: boolean;
}

export interface SupplierProfile {
  user_id: string;
  supplier_id: string;
  role: string;
  is_active: boolean;
  user?: any;
  supplier?: Supplier;
}

