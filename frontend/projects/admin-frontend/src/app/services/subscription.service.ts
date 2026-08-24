import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Subscription {
  id: string;
  tenant_id: string;
  status: string;
  current_period_end: string;
  api_requests_today: number;
  api_requests_this_month: number;
  paystack_subscription_code?: string;
  paystack_customer_code?: string;
  tenant?: {
    name: string;
    subdomain: string;
  };
  plan?: {
    name: string;
    price: number;
    api_daily_limit: number;
  };
}

export interface SubscriptionResponse {
  data: Subscription[];
  stats: {
    total: number;
    active: number;
    trialing: number;
    past_due: number;
    mrr: number;
  };
}

@Injectable({
  providedIn: 'root'
})
export class SubscriptionService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/subscriptions';

  getSubscriptions(): Observable<SubscriptionResponse> {
    return this.http.get<SubscriptionResponse>(this.apiUrl);
  }

  overrideSubscription(id: string, status: string): Observable<any> {
    return this.http.post(`${this.apiUrl}/${id}/override`, { status });
  }
}
