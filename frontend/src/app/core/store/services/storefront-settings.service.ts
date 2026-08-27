import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';

export interface StorefrontSettings {
  store_name?: string;
  banner_image?: string;
  logo_image?: string;
  primary_color?: string;
  welcome_message?: string;
  about_text?: string;
  allow_pickup?: boolean;
  allow_delivery?: boolean;
  delivery_fee?: number;
  min_order_amount?: number;
  enable_paystack?: boolean;
  paystack_public_key?: string;
  paystack_subaccount_code?: string;
  currency?: string;
  currency_symbol?: string;
  slug?: string;
  is_active?: boolean;
  flash_sale_end_time?: string;
}

@Injectable({
  providedIn: 'root'
})
export class StorefrontSettingsService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/storefront/config';

  settings = signal<StorefrontSettings | null>(null);

  loadSettings(): Observable<StorefrontSettings> {
    return this.http.get<StorefrontSettings>(this.apiUrl).pipe(
      tap(res => {
        this.settings.set(res);
        if (res.primary_color && typeof document !== 'undefined') {
          // Setting CSS variable on document root
          document.documentElement.style.setProperty('--primary-color', res.primary_color);
        }
      })
    );
  }

  updateSettings(settings: StorefrontSettings): Observable<StorefrontSettings> {
    return this.http.put<StorefrontSettings>(this.apiUrl, settings).pipe(
      tap(res => {
        this.settings.set(res);
      })
    );
  }
}
