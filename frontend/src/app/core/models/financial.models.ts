export interface Supplier {
  id: string;
  name: string;
  email?: string;
  contact_email?: string;
  contact_person?: string;
  phone?: string;
  address?: string;
  website?: string;
  tax_number?: string;
  status?: string;
  is_active?: boolean;
  credit_balance?: number;
  payment_terms?: string;
  notes?: string;
  portal_email?: string;
}
export interface ExpenseCategory { id: string; name: string; type: string; description?: string; monthly_budget?: number; }
export interface Expense { id: string; amount?: number; description?: string; date?: string; category_id?: string; is_recurring?: boolean; recurrence_interval?: string; receipt_url?: string; category?: ExpenseCategory; }
export interface TaxConfiguration { id: string; tax_type?: string; tax_rate?: number; is_active?: boolean; }
export interface GiftCard { id: string; code?: string; balance?: number; is_active?: boolean; }
export interface LoyaltyTransaction { id: string; customer_id?: string; points?: number; type?: string; }
export interface BillingPayment { id: string; amount?: number; date?: string; status?: string; receipt_url?: string; }
