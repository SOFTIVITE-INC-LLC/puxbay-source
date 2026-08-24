import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { Return } from '../models/order.models';

export interface ReturnItemInput {
  product_id?: string;
  quantity: number;
  reason?: string;
  restock: boolean;
}

export interface ReturnCreateInput {
  order_id: string;
  branch_id?: string;
  customer_id?: string;
  reason: string;
  reason_detail?: string;
  refund_method?: string;
  refund_amount?: number;
  items?: ReturnItemInput[];
}

@Injectable({
  providedIn: 'root'
})
export class ReturnService {
  private api = inject(ApiService);
  
  returns = signal<Return[]>([]);
  loading = signal<boolean>(false);

  getReturns(params?: any): Observable<{returns: Return[]}> {
    this.loading.set(true);
    return this.api.get<{returns: Return[]}>('/returns', { params }).pipe(
      tap(res => {
        this.returns.set(res.returns || []);
        this.loading.set(false);
      })
    );
  }

  getReturn(id: string): Observable<Return> {
    return this.api.get<Return>(`/returns/${id}`);
  }

  createReturn(input: ReturnCreateInput): Observable<Return> {
    return this.api.post<Return>('/returns', input).pipe(
      tap(ret => this.returns.update(list => [ret, ...list]))
    );
  }

  approveReturn(id: string): Observable<{message: string, return: Return}> {
    return this.api.post<{message: string, return: Return}>(`/returns/${id}/approve`, {}).pipe(
      tap(res => {
        this.returns.update(list => list.map(r => r.id === id ? res.return : r));
      })
    );
  }

  rejectReturn(id: string): Observable<{message: string, return: Return}> {
    return this.api.post<{message: string, return: Return}>(`/returns/${id}/reject`, {}).pipe(
      tap(res => {
        this.returns.update(list => list.map(r => r.id === id ? res.return : r));
      })
    );
  }

  processRefund(id: string): Observable<{message: string, net_refund: number, return: Return}> {
    return this.api.post<{message: string, net_refund: number, return: Return}>(`/returns/${id}/refund`, {}).pipe(
      tap(res => {
        this.returns.update(list => list.map(r => r.id === id ? res.return : r));
      })
    );
  }
}
