import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { Campaign } from '../models/crm.models';
export type { Campaign };

export interface CustomerSegment {
  id: string;
  name: string;
  description?: string;
  criteria_json?: string;
  created_at?: string;
}

export interface CampaignAnalytics {
  open_count: number;
  click_count: number;
  conversion_count: number;
  revenue_generated: number;
}

@Injectable({
  providedIn: 'root'
})
export class MarketingService {
  private api = inject(ApiService);

  campaigns = signal<Campaign[]>([]);
  promotions = signal<any[]>([]);
  discounts = signal<any[]>([]);
  segments = signal<CustomerSegment[]>([]);
  loading = signal<boolean>(false);

  // ── Campaigns ──────────────────────────────────────────────────────────

  getCampaigns(): Observable<Campaign[]> {
    this.loading.set(true);
    return this.api.get<Campaign[]>('/marketing/campaigns').pipe(
      tap(res => {
        this.campaigns.set(res || []);
        this.loading.set(false);
      })
    );
  }

  createCampaign(c: any): Observable<Campaign> {
    return this.api.post<Campaign>('/marketing/campaigns', c).pipe(
      tap(res => this.campaigns.update(list => [res, ...list]))
    );
  }

  updateCampaign(id: string, c: any): Observable<Campaign> {
    return this.api.put<Campaign>('/marketing/campaigns/' + id, c).pipe(
      tap(updated => this.campaigns.update(list => list.map(x => x.id === id ? updated : x)))
    );
  }

  deleteCampaign(id: string): Observable<any> {
    return this.api.delete<any>('/marketing/campaigns/' + id).pipe(
      tap(() => this.campaigns.update(list => list.filter(x => x.id !== id)))
    );
  }

  sendCampaign(id: string): Observable<any> {
    return this.api.post<any>(`/marketing/campaigns/${id}/send`, {}).pipe(
      tap(() => this.campaigns.update(list => list.map(x => x.id === id ? { ...x, status: 'sent' } : x)))
    );
  }

  recordOpen(id: string): Observable<any> {
    return this.api.post<any>(`/marketing/campaigns/${id}/open`, {});
  }

  recordConversion(id: string, revenue: number): Observable<any> {
    return this.api.post<any>(`/marketing/campaigns/${id}/convert`, { revenue });
  }

  triggerEventCampaigns(eventType: string): Observable<any> {
    return this.api.post<any>('/marketing/trigger', { event_type: eventType });
  }

  // ── Segments ───────────────────────────────────────────────────────────

  getSegments(): Observable<CustomerSegment[]> {
    return this.api.get<CustomerSegment[]>('/marketing/segments').pipe(
      tap(res => this.segments.set(res || []))
    );
  }

  createSegment(s: Partial<CustomerSegment>): Observable<CustomerSegment> {
    return this.api.post<CustomerSegment>('/marketing/segments', s).pipe(
      tap(res => this.segments.update(list => [res, ...list]))
    );
  }

  updateSegment(id: string, s: Partial<CustomerSegment>): Observable<CustomerSegment> {
    return this.api.put<CustomerSegment>(`/marketing/segments/${id}`, s).pipe(
      tap(updated => this.segments.update(list => list.map(x => x.id === id ? updated : x)))
    );
  }

  deleteSegment(id: string): Observable<any> {
    return this.api.delete<any>(`/marketing/segments/${id}`).pipe(
      tap(() => this.segments.update(list => list.filter(x => x.id !== id)))
    );
  }

  // ── Promotions ─────────────────────────────────────────────────────────

  getPromotions(): Observable<any[]> {
    return this.api.get<any[]>('/marketing/promotions').pipe(
      tap(res => this.promotions.set(res || []))
    );
  }

  createPromotion(p: any): Observable<any> {
    return this.api.post<any>('/marketing/promotions', p).pipe(
      tap(res => this.promotions.update(list => [res, ...list]))
    );
  }

  updatePromotion(id: string, p: any): Observable<any> {
    return this.api.put<any>(`/marketing/promotions/${id}`, p).pipe(
      tap(updated => this.promotions.update(list => list.map(x => x.id === id ? updated : x)))
    );
  }

  deletePromotion(id: string): Observable<any> {
    return this.api.delete<any>(`/marketing/promotions/${id}`).pipe(
      tap(() => this.promotions.update(list => list.filter(x => x.id !== id)))
    );
  }

  // ── Discounts ──────────────────────────────────────────────────────────

  getDiscounts(): Observable<any[]> {
    return this.api.get<any[]>('/marketing/discounts').pipe(
      tap(res => this.discounts.set(res || []))
    );
  }

  createDiscount(d: any): Observable<any> {
    return this.api.post<any>('/marketing/discounts', d).pipe(
      tap(res => this.discounts.update(list => [res, ...list]))
    );
  }

  deleteDiscount(id: string): Observable<any> {
    return this.api.delete<any>(`/marketing/discounts/${id}`).pipe(
      tap(() => this.discounts.update(list => list.filter(x => x.id !== id)))
    );
  }

  // ── Loyalty Redemption ─────────────────────────────────────────────────

  redeemPointsForDiscount(customerId: string, points: number, discountValue: number): Observable<any> {
    return this.api.post<any>('/marketing/redeem-points', {
      customer_id: customerId,
      points,
      discount_value: discountValue
    });
  }
}
