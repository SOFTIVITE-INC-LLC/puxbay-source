import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

export interface PaystackSubaccount {
  id: number;
  subaccount_code: string;
  business_name: string;
  settlement_bank: string;
  account_number: string;
  percentage_charge: number;
  currency: string;
  active: boolean;
  is_verified: boolean;
}

export interface PaystackCountry {
  id: number;
  name: string;
  iso_code: string;
  default_currency_code: string;
  integration_defaults: Record<string, unknown>;
  calling_code: string;
  pilot_mode: boolean;
  relationships: {
    currency: { type: string; data: string[] };
    integration_feature: { type: string; data: string[] };
    integration_type: { type: string; data: string[] };
    payment_method: { type: string; data: string[] };
  };
}

export interface PaystackBank {
  id: number;
  name: string;
  slug: string;
  code: string;
  longcode: string;
  gateway: string | null;
  pay_with_bank: boolean;
  active: boolean;
  country: string;
  currency: string;
  type: string;
  is_deleted: boolean;
}

export interface PaymentMethod {
  id: string;
  name: string;
  provider: 'cash' | 'mobile' | 'card' | 'bank_transfer' | 'stripe' | 'paystack' | 'paystack_subaccount' | 'crypto' | 'custom' | string;
  is_active: boolean;
  api_key_hint?: string;
  paystack_subaccount_code?: string;
  created_at?: string;
  updated_at?: string;
}

@Injectable({
  providedIn: 'root'
})
export class PaymentMethodService {
  private api = inject(ApiService);
  
  methods = signal<PaymentMethod[]>([]);
  loading = signal<boolean>(false);

  getMethods(): Observable<any> {
    this.loading.set(true);
    return this.api.get<any>('/payment-methods').pipe(
      tap(res => {
        let list: PaymentMethod[] = [];
        if (Array.isArray(res)) {
          list = res;
        } else if (res && Array.isArray(res.payment_methods)) {
          list = res.payment_methods;
        } else if (res && Array.isArray(res.data)) {
          list = res.data;
        }
        this.methods.set(list);
        this.loading.set(false);
      })
    );
  }

  createMethod(method: Partial<PaymentMethod>): Observable<PaymentMethod> {
    return this.api.post<PaymentMethod>('/payment-methods', method).pipe(
      tap(newMethod => {
        if (newMethod && newMethod.id) {
          this.methods.update(current => [...current, newMethod]);
        } else {
          this.getMethods().subscribe();
        }
      })
    );
  }

  updateMethod(id: string, updates: Partial<PaymentMethod>): Observable<PaymentMethod> {
    return this.api.put<PaymentMethod>(`/payment-methods/${id}`, updates).pipe(
      tap(updated => {
        this.methods.update(current =>
          current.map(m => (m.id === id ? { ...m, ...updates, ...updated } : m))
        );
      })
    );
  }

  toggleMethod(id: string, isActive: boolean): Observable<PaymentMethod> {
    this.methods.update(current =>
      current.map(m => (m.id === id ? { ...m, is_active: isActive } : m))
    );
    return this.api.put<PaymentMethod>(`/payment-methods/${id}`, { is_active: isActive });
  }

  deleteMethod(id: string): Observable<any> {
    return this.api.delete<any>(`/payment-methods/${id}`).pipe(
      tap(() => {
        this.methods.update(current => current.filter(m => m.id !== id));
      })
    );
  }

  // ── Guided Paystack Subaccount Creation Flow ──────────────────────────────

  /** List countries where Paystack is available */
  getPaystackCountries(): Observable<{ status: boolean; data: PaystackCountry[] }> {
    return this.api.get<{ status: boolean; data: PaystackCountry[] }>(
      '/payment-methods/paystack/countries', undefined, true
    );
  }

  /** List banks for the given country (e.g. "ghana", "nigeria") */
  getPaystackBanks(country: string, currency?: string): Observable<{ status: boolean; data: PaystackBank[] }> {
    const params: Record<string, string> = { country };
    if (currency) params['currency'] = currency;
    return this.api.get<{ status: boolean; data: PaystackBank[] }>(
      '/payment-methods/paystack/banks',
      { params },
      true
    );
  }

  /** Resolve account number to account name via Paystack */
  resolvePaystackAccount(accountNumber: string, bankCode: string): Observable<{
    account_number: string;
    account_name: string;
    bank_id: number;
  }> {
    return this.api.get<any>(
      '/payment-methods/paystack/resolve-account',
      { params: { account_number: accountNumber, bank_code: bankCode } },
      true
    );
  }

  /** Create a Paystack subaccount and save it locally */
  createPaystackSubaccount(payload: {
    business_name: string;
    settlement_bank: string;
    account_number: string;
    percentage_charge?: number;
    description?: string;
    primary_contact_email?: string;
    local_name?: string;
    is_active?: boolean;
  }): Observable<{ subaccount: PaystackSubaccount; payment_method?: PaymentMethod; warning?: string }> {
    return this.api.post<{ subaccount: PaystackSubaccount; payment_method?: PaymentMethod; warning?: string }>(
      '/payment-methods/paystack/create-subaccount',
      payload
    ).pipe(
      tap(res => {
        if (res && res.payment_method) {
          const pm = res.payment_method;
          this.methods.update(current => [...current, pm]);
        } else {
          this.getMethods().subscribe();
        }
      })
    );
  }

  // ── Legacy helpers (kept for edit/list subaccounts flow) ─────────────────

  getPaystackSubaccounts(): Observable<{ subaccounts: PaystackSubaccount[] }> {
    return this.api.get<{ subaccounts: PaystackSubaccount[] }>('/payment-methods/paystack/subaccounts', undefined, true);
  }

  verifyPaystackSubaccount(code: string): Observable<{ subaccount: PaystackSubaccount }> {
    return this.api.get<{ subaccount: PaystackSubaccount }>(
      `/payment-methods/paystack/subaccounts/verify/${encodeURIComponent(code)}`, undefined, true
    );
  }
}
