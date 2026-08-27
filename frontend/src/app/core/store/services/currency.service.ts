import { Injectable, signal, inject } from '@angular/core';
import { StorefrontSettingsService } from './storefront-settings.service';

export interface Currency {
  code: string;
  symbol: string;
  rate: number; // Rate relative to USD (1 USD = rate in currency)
}

@Injectable({
  providedIn: 'root'
})
export class CurrencyService {
  private storefrontSettings = inject(StorefrontSettingsService, { optional: true });

  // Standard conversion rates relative to USD (1 USD = X in currency)
  private readonly currencies: Currency[] = [
    { code: 'GHS', symbol: 'GH₵', rate: 14.5 },
    { code: 'USD', symbol: '$', rate: 1.0 },
    { code: 'EUR', symbol: '€', rate: 0.92 },
    { code: 'GBP', symbol: '£', rate: 0.79 },
    { code: 'NGN', symbol: '₦', rate: 1200 },
    { code: 'KES', symbol: 'KSh', rate: 130 },
    { code: 'ZAR', symbol: 'R', rate: 18.5 },
    { code: 'CAD', symbol: 'CA$', rate: 1.36 }
  ];

  baseCurrencyCode = signal<string>('GHS');
  activeCurrency = signal<Currency>(this.currencies[0]);

  getCurrencies(): Currency[] {
    return this.currencies;
  }

  setBaseCurrency(code: string) {
    if (!code) return;
    this.baseCurrencyCode.set(code.toUpperCase());
    
    // If user hasn't explicitly selected another currency in session, default active currency to base
    if (typeof localStorage !== 'undefined') {
      const saved = localStorage.getItem('store_currency');
      if (!saved) {
        this.setCurrency(code);
      }
    } else {
      this.setCurrency(code);
    }
  }

  setCurrency(code: string) {
    const currency = this.currencies.find(c => c.code.toUpperCase() === code.toUpperCase());
    if (currency) {
      this.activeCurrency.set(currency);
      // Persist user preference
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('store_currency', currency.code);
      }
    }
  }

  init(tenantDefaultCurrency?: string) {
    const base = tenantDefaultCurrency || this.storefrontSettings?.settings()?.currency || 'GHS';
    this.baseCurrencyCode.set(base.toUpperCase());

    if (typeof localStorage !== 'undefined') {
      const saved = localStorage.getItem('store_currency');
      if (saved) {
        this.setCurrency(saved);
        return;
      }
    }
    // Default to the tenant's base currency
    this.setCurrency(base);
  }

  /**
   * Converts an amount from tenant base currency to the user's selected active currency
   */
  convert(amountInTenantBase: number): number {
    const baseCode = this.baseCurrencyCode().toUpperCase();
    const targetCode = this.activeCurrency().code.toUpperCase();

    if (baseCode === targetCode) {
      return amountInTenantBase;
    }

    const baseCurrency = this.currencies.find(c => c.code === baseCode);
    const targetCurrency = this.currencies.find(c => c.code === targetCode);

    const baseRate = baseCurrency ? baseCurrency.rate : 1.0;
    const targetRate = targetCurrency ? targetCurrency.rate : 1.0;

    // Convert from base currency -> USD -> target currency
    // amountInUSD = amountInTenantBase / baseRate
    // amountInTarget = amountInUSD * targetRate
    return amountInTenantBase * (targetRate / baseRate);
  }
}
