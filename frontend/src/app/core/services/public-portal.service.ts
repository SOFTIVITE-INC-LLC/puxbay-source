import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { Product } from '../models/models';

export interface TenantPublicInfo {
  name: string;
  description: string;
  logo_url: string;
  contact_email: string;
}

export interface TrackOrderResult {
  order_id: string;
  status: string;
  created_at: string;
  total: number;
}

@Injectable({
  providedIn: 'root'
})
export class PublicPortalService {
  private api = inject(ApiService);
  
  loading = signal<boolean>(false);

  getTenantInfo(domain: string): Observable<TenantPublicInfo> {
    this.loading.set(true);
    return this.api.get<TenantPublicInfo>('/public/tenant-info', { params: { domain } }).pipe(
      tap(() => this.loading.set(false))
    );
  }

  listProducts(tenantId: string): Observable<{products: Product[]}> {
    this.loading.set(true);
    return this.api.get<{products: Product[]}>('/public/products', { params: { tenant_id: tenantId } }).pipe(
      tap(() => this.loading.set(false))
    );
  }

  trackOrder(tenantId: string, orderId: string): Observable<TrackOrderResult> {
    this.loading.set(true);
    return this.api.get<TrackOrderResult>('/public/track-order', { params: { tenant_id: tenantId, order_id: orderId } }).pipe(
      tap(() => this.loading.set(false))
    );
  }

  submitFeedback(tenantId: string, data: { name: string, email: string, rating: number, comment: string }): Observable<any> {
    this.loading.set(true);
    return this.api.post('/public/feedback', data, { params: { tenant_id: tenantId } }).pipe(
      tap(() => this.loading.set(false))
    );
  }
}
