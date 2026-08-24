import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { tap } from 'rxjs';

export interface WalletCustomer {
  id: string;
  name: string;
  phone?: string;
  email?: string;
  loyalty_pts: number;
  store_credit: number;
  debt_balance: number;
  total_spend: number;
  total_orders: number;
}

export interface LoyaltyTransaction {
  id: string;
  points: number;
  type: 'earn' | 'redeem' | 'adjust' | 'expire';
  note?: string;
  created_at: string;
}

export interface GiftCard {
  id: string;
  code: string;
  balance: number;
  original_amount: number;
  is_active: boolean;
  expires_at?: string;
  created_at: string;
}

export interface WalletOrder {
  id: string;
  order_number: string;
  total: number;
  status: string;
  payment_method: string;
  created_at: string;
}

export interface WalletDashboard {
  customer: WalletCustomer;
  recent_orders: WalletOrder[];
  gift_cards: GiftCard[];
  loyalty_history: LoyaltyTransaction[];
}

@Injectable({ providedIn: 'root' })
export class WalletService {
  private api = inject(ApiService);

  dashboard = signal<WalletDashboard | null>(null);
  loading = signal(false);
  searchQuery = signal('');
  lookupResult = signal<{ customer_id: string; name: string; loyalty_points: number } | null>(null);
  lookupError = signal('');
  lookupLoading = signal(false);

  loadDashboard(customerId: string) {
    this.loading.set(true);
    return this.api.get<WalletDashboard>(`/wallet/dashboard?customer_id=${customerId}`).pipe(
      tap(res => {
        this.dashboard.set(res);
        this.loading.set(false);
      })
    );
  }

  lookupByPhone(phone: string) {
    this.lookupLoading.set(true);
    this.lookupError.set('');
    return this.api.post<{ customer_id: string; name: string; loyalty_points: number }>(
      '/wallet/lookup', { phone }
    ).pipe(
      tap({
        next: res => {
          this.lookupResult.set(res);
          this.lookupLoading.set(false);
        },
        error: () => {
          this.lookupError.set('No customer found with this phone number.');
          this.lookupLoading.set(false);
        }
      })
    );
  }

  adjustLoyaltyPoints(customerId: string, points: number, note: string) {
    return this.api.post(`/wallet/customers/${customerId}/loyalty/adjust`, { points, note });
  }

  adjustStoreCredit(customerId: string, amount: number, note: string) {
    return this.api.post(`/wallet/customers/${customerId}/store-credit/adjust`, { amount, note });
  }

  issueGiftCard(customerId: string, amount: number) {
    return this.api.post('/wallet/gift-cards', { purchaser_id: customerId, original_amount: amount, balance: amount, is_active: true });
  }
}
