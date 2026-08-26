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
    // Optimistic UI update
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

  // Paystack Subaccount API integration
  getPaystackSubaccounts(): Observable<{ subaccounts: PaystackSubaccount[] }> {
    return this.api.get<{ subaccounts: PaystackSubaccount[] }>('/payment-methods/paystack/subaccounts', undefined, true);
  }

  verifyPaystackSubaccount(code: string): Observable<{ subaccount: PaystackSubaccount }> {
    return this.api.get<{ subaccount: PaystackSubaccount }>(`/payment-methods/paystack/subaccounts/verify/${encodeURIComponent(code)}`, undefined, true);
  }
}
