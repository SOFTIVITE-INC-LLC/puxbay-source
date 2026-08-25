import { Injectable, inject, signal, computed } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { BranchService } from './branch.service';

export interface GlobalSettings {
  currency: string;
  timezone: string;
  date_format: string;
  enable_email_receipts: boolean;
  hardware_proxy_url: string;
  enable_hardware_proxy: boolean;
  auto_print_receipts: boolean;
  enable_sms_notifications: boolean;
  enable_push_notifications: boolean;
  admin_notification_email: string;
  promo_threshold: number;
  promo_discount_percent: number;
}

/** ISO 4217 currency code → display symbol map */
export const CURRENCY_SYMBOL_MAP: Record<string, string> = {
  USD: '$', EUR: '€', GBP: '£', JPY: '¥', CNY: '¥',
  GHS: 'GH₵', NGN: '₦', KES: 'KSh', ZAR: 'R',
  CAD: 'CA$', AUD: 'A$', INR: '₹', BRL: 'R$',
  MXN: 'MX$', CHF: 'CHF', SEK: 'kr', NOK: 'kr',
  DKK: 'kr', SGD: 'S$', HKD: 'HK$', NZD: 'NZ$',
  EGP: 'E£', TZS: 'TSh', UGX: 'USh', RWF: 'RF',
  XOF: 'CFA', XAF: 'FCFA',
};

@Injectable({
  providedIn: 'root'
})
export class SettingsService {
  private api = inject(ApiService);
  private branchService = inject(BranchService);
  
  settings = signal<GlobalSettings | null>(null);
  loading = signal<boolean>(false);

  /**
   * Reactive currency symbol for the current context.
   * Priority: active branch currency → global settings currency → localStorage → USD
   */
  currencySymbol = computed(() => {
    const branchCode = this.branchService.activeBranch()?.currency_code;
    const globalCode = this.settings()?.currency;
    const storedCode = typeof window !== 'undefined' ? localStorage.getItem('currency_code') : null;
    const code = branchCode || globalCode || storedCode || 'USD';
    return CURRENCY_SYMBOL_MAP[code] ?? code;
  });

  getSettings(): Observable<GlobalSettings> {
    this.loading.set(true);
    return this.api.get<GlobalSettings>('/settings').pipe(
      tap(res => {
        this.settings.set(res);
        if (res.currency) {
          localStorage.setItem('currency_code', res.currency);
        }
        this.loading.set(false);
      })
    );
  }

  updateSettings(settings: GlobalSettings): Observable<GlobalSettings> {
    return this.api.put<GlobalSettings>('/settings', settings).pipe(
      tap(res => {
        this.settings.set(res);
        if (res.currency && localStorage.getItem('currency_code') !== res.currency) {
          localStorage.setItem('currency_code', res.currency);
          // appCurrency pipe will automatically react to the Signal change!
        }
      })
    );
  }
}
