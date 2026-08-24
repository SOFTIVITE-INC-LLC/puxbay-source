import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface ReferralReward {
  id: string;
  referrer_id: string;
  referred_tenant_id: string;
  reward_amount: number;
  is_applied: boolean;
  applied_at?: string;
  created_at: string;
  referrer?: { name: string; subdomain: string; };
  referred_tenant?: { name: string; subdomain: string; };
}

export interface WebhookEvent {
  id: string;
  endpoint_id?: string;
  event_type: string;
  payload?: any;
  status: string;
  response_code?: number;
  attempts: number;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class GrowthService {
  private http = inject(HttpClient);
  private base = '/api/v1/admin';

  getReferrals(): Observable<{ data: ReferralReward[] }> {
    return this.http.get<{ data: ReferralReward[] }>(`${this.base}/referrals`);
  }

  getWebhookEvents(): Observable<{ data: WebhookEvent[] }> {
    return this.http.get<{ data: WebhookEvent[] }>(`${this.base}/webhook-events`);
  }

  getUpcomingRenewals(): Observable<{ data: any[] }> {
    return this.http.get<{ data: any[] }>(`${this.base}/subscriptions/upcoming-renewals`);
  }

  getFailedPayments(): Observable<{ data: any[] }> {
    return this.http.get<{ data: any[] }>(`${this.base}/payments/failed`);
  }
}
