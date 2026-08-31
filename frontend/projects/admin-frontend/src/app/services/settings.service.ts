import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import { delay } from 'rxjs/operators';

export interface FeatureFlags {
  enable_new_dashboard: boolean;
  enable_pos_beta: boolean;
  maintenance_mode: boolean;
  enable_api_keys: boolean;
}

export interface PlatformContactSettings {
  company_name: string;
  headquarters_address: string;
  contact_phone: string;
  support_phone: string;
  contact_email: string;
  sales_email: string;
  support_email: string;
  working_hours: string;
}

@Injectable({
  providedIn: 'root'
})
export class SettingsService {

  private flags: FeatureFlags = {
    enable_new_dashboard: true,
    enable_pos_beta: false,
    maintenance_mode: false,
    enable_api_keys: true
  };

  private defaultContact: PlatformContactSettings = {
    company_name: 'Puxbay / Softivite',
    headquarters_address: 'No. 12 Independence Avenue, Ridge, Accra, Ghana',
    contact_phone: '+233 (0) 246136978',
    support_phone: '+233 (0) 598001682',
    contact_email: 'support@puxbay.com',
    sales_email: 'sales@puxbay.com',
    support_email: 'support@puxbay.com',
    working_hours: 'Mon - Fri, 8:00 AM - 6:00 PM GMT'
  };

  getFeatureFlags(): Observable<FeatureFlags> {
    return of({ ...this.flags }).pipe(delay(500));
  }

  updateFeatureFlags(flags: FeatureFlags): Observable<any> {
    this.flags = { ...flags };
    return of({ status: 'updated' }).pipe(delay(800));
  }

  getContactSettings(): Observable<PlatformContactSettings> {
    try {
      const saved = localStorage.getItem('puxbay_platform_contact');
      if (saved) {
        return of({ ...this.defaultContact, ...JSON.parse(saved) }).pipe(delay(300));
      }
    } catch (_) { }
    return of({ ...this.defaultContact }).pipe(delay(300));
  }

  updateContactSettings(contact: PlatformContactSettings): Observable<any> {
    try {
      localStorage.setItem('puxbay_platform_contact', JSON.stringify(contact));
    } catch (_) { }
    return of({ status: 'updated', data: contact }).pipe(delay(500));
  }
}
