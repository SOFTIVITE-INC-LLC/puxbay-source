import { Pipe, PipeTransform, inject } from '@angular/core';
import { CurrencyPipe } from '@angular/common';
import { Router } from '@angular/router';
import { SettingsService } from '../services/settings.service';
import { CurrencyService } from '../store/services/currency.service';
import { BranchService } from '../services/branch.service';

@Pipe({
  name: 'appCurrency',
  standalone: true,
  pure: false // Impure so it updates immediately when the signal changes without reference changes to the input
})
export class AppCurrencyPipe implements PipeTransform {
  private currencyPipe = new CurrencyPipe('en-US');
  private settingsService = inject(SettingsService);
  private storefrontCurrency = inject(CurrencyService, { optional: true });
  private branchService = inject(BranchService);
  private router = inject(Router);

  transform(
    value: number | string | null | undefined,
    display: 'code' | 'symbol' | 'symbol-narrow' | string | boolean = 'symbol',
    digitsInfo?: string,
    locale?: string
  ): string | null {
    if (value == null) return null;
    
    let currentCode = 'USD';
    let valToFormat = Number(value);

    // Check if we are in the storefront based on URL
    const isStorefront = this.router.url.startsWith('/store') || this.router.url.startsWith('/shop') || (this.router.url.includes('/checkout') && !this.router.url.includes('/billing'));

    // If we're in the storefront and have a CurrencyService, use it to convert!
    if (isStorefront && this.storefrontCurrency) {
      const active = this.storefrontCurrency.activeCurrency();
      currentCode = active.code;
      valToFormat = this.storefrontCurrency.convert(valToFormat);
    } else {
      // Fallback priority:
      // 1. Current Branch Currency
      // 2. Global Company Currency
      // 3. LocalStorage
      // 4. USD
      currentCode = this.branchService.activeBranch()?.currency_code || 
                    this.settingsService.settings()?.currency || 
                    (typeof window !== 'undefined' ? localStorage.getItem('currency_code') : null) || 
                    'USD';
    }
    
    return this.currencyPipe.transform(valToFormat, currentCode, display, digitsInfo, locale);
  }
}
