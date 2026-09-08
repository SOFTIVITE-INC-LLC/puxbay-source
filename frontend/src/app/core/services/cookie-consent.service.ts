import { Injectable, PLATFORM_ID, inject, signal } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';

export interface CookiePreferences {
  necessary: boolean;
  analytics: boolean;
  functional: boolean;
  marketing: boolean;
}

export interface CookieConsentRecord {
  accepted: boolean;
  timestamp: string;
  version: string;
  preferences: CookiePreferences;
}

const STORAGE_KEY = 'pux_cookie_consent_v1';
const CURRENT_VERSION = '1.0';

@Injectable({
  providedIn: 'root'
})
export class CookieConsentService {
  private readonly platformId = inject(PLATFORM_ID);
  private readonly isBrowser = isPlatformBrowser(this.platformId);

  readonly showBanner = signal<boolean>(false);
  readonly showModal = signal<boolean>(false);
  readonly hasConsented = signal<boolean>(false);
  readonly consentDate = signal<string | null>(null);

  readonly preferences = signal<CookiePreferences>({
    necessary: true,
    analytics: false,
    functional: false,
    marketing: false
  });

  constructor() {
    this.initConsent();
  }

  private initConsent(): void {
    if (!this.isBrowser) return;

    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const parsed: CookieConsentRecord = JSON.parse(stored);
        if (parsed && parsed.version === CURRENT_VERSION && parsed.preferences) {
          this.preferences.set({
            ...parsed.preferences,
            necessary: true // strictly necessary is always true
          });
          this.hasConsented.set(true);
          this.consentDate.set(parsed.timestamp);
          this.showBanner.set(false);
          return;
        }
      }
    } catch {
      // Fallback if localStorage access fails
    }

    // No consent given yet: display the banner
    this.hasConsented.set(false);
    this.showBanner.set(true);
  }

  acceptAll(): void {
    const allAccepted: CookiePreferences = {
      necessary: true,
      analytics: true,
      functional: true,
      marketing: true
    };
    this.saveConsent(allAccepted);
  }

  acceptNecessaryOnly(): void {
    const necessaryOnly: CookiePreferences = {
      necessary: true,
      analytics: false,
      functional: false,
      marketing: false
    };
    this.saveConsent(necessaryOnly);
  }

  saveCustomPreferences(prefs: Partial<CookiePreferences>): void {
    const updated: CookiePreferences = {
      necessary: true, // Always locked
      analytics: !!prefs.analytics,
      functional: !!prefs.functional,
      marketing: !!prefs.marketing
    };
    this.saveConsent(updated);
  }

  openPreferences(): void {
    this.showModal.set(true);
  }

  closePreferences(): void {
    this.showModal.set(false);
  }

  openBanner(): void {
    this.showBanner.set(true);
  }

  closeBanner(): void {
    this.showBanner.set(false);
  }

  resetConsent(): void {
    if (this.isBrowser) {
      try {
        localStorage.removeItem(STORAGE_KEY);
      } catch {
        // Ignore
      }
    }
    this.hasConsented.set(false);
    this.consentDate.set(null);
    this.preferences.set({
      necessary: true,
      analytics: false,
      functional: false,
      marketing: false
    });
    this.showBanner.set(true);
  }

  isCategoryAllowed(category: keyof CookiePreferences): boolean {
    return !!this.preferences()[category];
  }

  private saveConsent(prefs: CookiePreferences): void {
    const record: CookieConsentRecord = {
      accepted: true,
      timestamp: new Date().toISOString(),
      version: CURRENT_VERSION,
      preferences: prefs
    };

    if (this.isBrowser) {
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(record));
      } catch {
        // Ignore storage errors
      }
    }

    this.preferences.set(prefs);
    this.hasConsented.set(true);
    this.consentDate.set(record.timestamp);
    this.showBanner.set(false);
    this.showModal.set(false);

    if (this.isBrowser) {
      window.dispatchEvent(
        new CustomEvent('pux:cookie-consent-updated', {
          detail: record
        })
      );
    }
  }
}
