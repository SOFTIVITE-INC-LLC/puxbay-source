import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

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

@Injectable({
  providedIn: 'root'
})
export class SettingsService {
  private api = inject(ApiService);
  
  settings = signal<GlobalSettings | null>(null);
  loading = signal<boolean>(false);

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
