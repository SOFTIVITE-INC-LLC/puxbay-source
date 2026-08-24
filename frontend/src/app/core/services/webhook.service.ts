import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

export interface Webhook {
  id: string;
  url: string;
  event_types: string[];
  is_active: boolean;
  secret: string;
  created_at?: string;
  last_triggered_at?: string;
  failure_count?: number;
}

export interface WebhookDelivery {
  id: string;
  webhook_id: string;
  event_type: string;
  status: 'success' | 'failed' | 'pending';
  status_code?: number;
  payload?: string;
  response?: string;
  created_at: string;
}

export const WEBHOOK_EVENTS = [
  'order.created', 'order.updated', 'order.completed', 'order.cancelled',
  'customer.created', 'customer.updated',
  'product.created', 'product.updated', 'product.low_stock',
  'payment.completed', 'payment.refunded',
  'inventory.transfer.created', 'inventory.po.received'
];

@Injectable({
  providedIn: 'root'
})
export class WebhookService {
  private api = inject(ApiService);

  webhooks = signal<Webhook[]>([]);
  deliveries = signal<WebhookDelivery[]>([]);
  loading = signal<boolean>(false);

  getWebhooks(): Observable<{webhooks: Webhook[]}> {
    this.loading.set(true);
    return this.api.get<{webhooks: Webhook[]}>('/webhooks').pipe(
      tap(res => {
        this.webhooks.set(res.webhooks || []);
        this.loading.set(false);
      })
    );
  }

  createWebhook(webhook: Partial<Webhook>): Observable<Webhook> {
    return this.api.post<Webhook>('/webhooks', webhook).pipe(
      tap(w => this.webhooks.update(list => [...list, w]))
    );
  }

  updateWebhook(id: string, data: Partial<Webhook>): Observable<Webhook> {
    return this.api.put<Webhook>(`/webhooks/${id}`, data).pipe(
      tap(w => this.webhooks.update(list => list.map(wh => wh.id === id ? w : wh)))
    );
  }

  deleteWebhook(id: string): Observable<void> {
    return this.api.delete<void>(`/webhooks/${id}`).pipe(
      tap(() => this.webhooks.update(list => list.filter(w => w.id !== id)))
    );
  }

  getDeliveries(webhookId: string): Observable<{deliveries: WebhookDelivery[]}> {
    return this.api.get<{deliveries: WebhookDelivery[]}>(`/webhooks/${webhookId}/deliveries`).pipe(
      tap(res => this.deliveries.set(res.deliveries || []))
    );
  }

  retryDelivery(webhookId: string, deliveryId: string): Observable<any> {
    return this.api.post<any>(`/webhooks/${webhookId}/deliveries/${deliveryId}/retry`, {});
  }

  testWebhook(webhookId: string): Observable<any> {
    return this.api.post<any>(`/webhooks/${webhookId}/test`, {});
  }
}
