import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface PaymentLog {
  id: string;
  tenant_id?: string;
  tenant_name?: string;
  payment_type: 'store_order' | 'pos_order' | 'subscription' | 'manual_entry' | 'dispute_settlement' | 'refund' | string;
  reference: string;
  order_id?: string;
  order_number?: string;
  amount: number;
  currency: string;
  payment_method: string;
  gateway: string;
  subaccount_code?: string;
  is_subaccount_routed: boolean;
  subaccount_share?: number;
  platform_fee?: number;
  customer_name?: string;
  customer_email?: string;
  customer_phone?: string;
  status: 'successful' | 'pending' | 'failed' | 'refunded' | 'disputed' | string;
  dispute_status: 'none' | 'under_review' | 'resolved' | 'chargeback' | string;
  notes?: string;
  raw_metadata?: any;
  created_at: string;
  updated_at: string;
  tenant?: {
    id: string;
    name: string;
    subdomain?: string;
  };
}

export type Payment = PaymentLog;

export interface PaymentFilterParams {
  search?: string;
  status?: string;
  payment_type?: string;
  subaccount_routed?: string;
  dispute_status?: string;
  tenant_id?: string;
  from_date?: string;
  to_date?: string;
  page?: number;
  limit?: number;
}

export interface PaymentStats {
  total_volume: number;
  successful_count: number;
  failed_count: number;
  disputed_count: number;
  subaccount_routed_volume: number;
  subaccount_routed_count: number;
  platform_fee_total: number;
}

export interface PaymentResponse {
  data: PaymentLog[];
  stats: PaymentStats;
  total: number;
  page: number;
  limit: number;
}

@Injectable({
  providedIn: 'root'
})
export class PaymentService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/payments';

  getPayments(params?: PaymentFilterParams): Observable<PaymentResponse> {
    let httpParams = new HttpParams();
    if (params) {
      if (params.search) httpParams = httpParams.set('search', params.search);
      if (params.status && params.status !== 'all') httpParams = httpParams.set('status', params.status);
      if (params.payment_type && params.payment_type !== 'all') httpParams = httpParams.set('payment_type', params.payment_type);
      if (params.subaccount_routed && params.subaccount_routed !== 'all') httpParams = httpParams.set('subaccount_routed', params.subaccount_routed);
      if (params.dispute_status && params.dispute_status !== 'all') httpParams = httpParams.set('dispute_status', params.dispute_status);
      if (params.tenant_id) httpParams = httpParams.set('tenant_id', params.tenant_id);
      if (params.from_date) httpParams = httpParams.set('from_date', params.from_date);
      if (params.to_date) httpParams = httpParams.set('to_date', params.to_date);
      if (params.page) httpParams = httpParams.set('page', params.page.toString());
      if (params.limit) httpParams = httpParams.set('limit', params.limit.toString());
    }
    return this.http.get<PaymentResponse>(this.apiUrl, { params: httpParams });
  }

  createPayment(payment: Partial<PaymentLog>): Observable<{ message: string; data: PaymentLog }> {
    return this.http.post<{ message: string; data: PaymentLog }>(this.apiUrl, payment);
  }

  getPayment(id: string): Observable<{ data: PaymentLog }> {
    return this.http.get<{ data: PaymentLog }>(`${this.apiUrl}/${id}`);
  }

  updatePayment(id: string, update: { status?: string; dispute_status?: string; notes?: string }): Observable<{ message: string; data: PaymentLog }> {
    return this.http.put<{ message: string; data: PaymentLog }>(`${this.apiUrl}/${id}`, update);
  }

  deletePayment(id: string): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(`${this.apiUrl}/${id}`);
  }
}
