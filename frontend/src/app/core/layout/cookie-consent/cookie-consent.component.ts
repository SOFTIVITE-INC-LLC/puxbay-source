import { Component, inject, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { CookieConsentService, CookiePreferences } from '../../services/cookie-consent.service';

@Component({
  selector: 'app-cookie-consent',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './cookie-consent.component.html',
  styleUrls: ['./cookie-consent.component.css']
})
export class CookieConsentComponent {
  readonly consentService = inject(CookieConsentService);

  // Local draft state for preferences modal
  readonly draftPreferences = signal<CookiePreferences>({
    necessary: true,
    analytics: false,
    functional: false,
    marketing: false
  });

  constructor() {
    // Sync draft with active preferences whenever modal opens
    effect(() => {
      if (this.consentService.showModal()) {
        const current = this.consentService.preferences();
        this.draftPreferences.set({ ...current });
      }
    });
  }

  acceptAll(): void {
    this.consentService.acceptAll();
  }

  acceptNecessary(): void {
    this.consentService.acceptNecessaryOnly();
  }

  openPreferences(): void {
    this.draftPreferences.set({ ...this.consentService.preferences() });
    this.consentService.openPreferences();
  }

  closePreferences(): void {
    this.consentService.closePreferences();
  }

  toggleCategory(category: 'analytics' | 'functional' | 'marketing'): void {
    const current = this.draftPreferences();
    this.draftPreferences.set({
      ...current,
      [category]: !current[category]
    });
  }

  saveDraftPreferences(): void {
    this.consentService.saveCustomPreferences(this.draftPreferences());
  }

  dismissBanner(): void {
    // When dismissed without choosing, keep necessary only or hide banner
    this.consentService.closeBanner();
  }
}
