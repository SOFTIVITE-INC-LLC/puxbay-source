import { Injectable, signal } from '@angular/core';

export interface Currency {
  code: string;
  symbol: string;
  rate: number; // Conversion rate relative to base (USD)
}

@Injectable({
  providedIn: 'root'
})
export class CurrencyService {
  // Hardcoded rates for demo purposes
  private readonly currencies: Currency[] = [
    { code: 'USD', symbol: '$', rate: 1 },
    { code: 'EUR', symbol: '€', rate: 0.92 },
    { code: 'GBP', symbol: '£', rate: 0.79 },
    { code: 'NGN', symbol: '₦', rate: 1200 },
    { code: 'GHS', symbol: 'GH₵', rate: 14.5 }
  ];

  activeCurrency = signal<Currency>(this.currencies[0]);

  getCurrencies(): Currency[] {
    return this.currencies;
  }

  setCurrency(code: string) {
    const currency = this.currencies.find(c => c.code === code);
    if (currency) {
      this.activeCurrency.set(currency);
      // Persist user preference
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('store_currency', code);
      }
    }
  }

  init() {
    if (typeof localStorage !== 'undefined') {
      const saved = localStorage.getItem('store_currency');
      if (saved) {
        this.setCurrency(saved);
      }
    }
  }

  convert(amountInUSD: number): number {
    return amountInUSD * this.activeCurrency().rate;
  }
}
