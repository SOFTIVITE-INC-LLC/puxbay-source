import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { BillingPayment } from '../models/financial.models';

export interface Plan {
  id: string;
  name: string;
  price: number;
}

@Injectable({
  providedIn: 'root'
})
export class BillingService {
  private api = inject(ApiService);
  
  invoices = signal<BillingPayment[]>([]);
  subscription = signal<any>(null);
  getSubscription() { return this.api.get<any>('/billing/subscription').pipe(tap(res => this.subscription.set(res))); }
  getInvoices() { return this.api.get<any[]>('/billing/invoices').pipe(tap(res => this.invoices.set(res || []))); }
  plans = signal<Plan[]>([]);
  loading = signal<boolean>(false);

  listInvoices(): Observable<BillingPayment[]> {
    this.loading.set(true);
    return this.api.get<BillingPayment[]>('/billing/invoices').pipe(
      tap(res => {
        this.invoices.set(res || []);
        this.loading.set(false);
      })
    );
  }

  validatePromo(code: string): Observable<{success: boolean, discount_type: string, discount: number, message: string}> {
    return this.api.post<{success: boolean, discount_type: string, discount: number, message: string}>('/billing/validate-promo', { code });
  }

  listPlans(): Observable<{plans: Plan[]}> {
    return this.api.get<{plans: Plan[]}>('/billing/plans').pipe(
      tap(res => {
        if (res && res.plans) {
          this.plans.set(res.plans);
        }
      })
    );
  }

  processPayment(planId: string, billingCycle: string = 'monthly', promoCode?: string): Observable<{url: string}> {
    return this.api.post<{url: string}>(`/billing/checkout/${planId}`, { billing_cycle: billingCycle, promo_code: promoCode });
  }
}
