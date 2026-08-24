import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { Quotation } from '../models/b2b.models';
import { Customer } from '../models/crm.models';

export interface QuoteItemInput {
  product_id: string;
  qty: number;
}

export interface QuoteCreateInput {
  customer_id: string;
  items: QuoteItemInput[];
}

export interface QuoteUpdateInput {
  action: string;
  internal_notes?: string;
}

@Injectable({
  providedIn: 'root'
})
export class B2bService {
  private api = inject(ApiService);
  
  quotes = signal<Quotation[]>([]);
  clients = signal<Customer[]>([]);
  loading = signal<boolean>(false);

  listQuotes(): Observable<Quotation[]> {
    this.loading.set(true);
    return this.api.get<Quotation[]>('/b2b/quotes').pipe(
      tap(res => {
        this.quotes.set(res || []);
        this.loading.set(false);
      })
    );
  }

  createQuote(input: QuoteCreateInput): Observable<{success: boolean, quote_id: string}> {
    return this.api.post<{success: boolean, quote_id: string}>('/b2b/quotes', input);
  }

  updateQuote(id: string, input: QuoteUpdateInput): Observable<{success: boolean, status: string}> {
    return this.api.post<{success: boolean, status: string}>(`/b2b/quotes/${id}`, input).pipe(
      tap(res => {
        if (res.success) {
          this.quotes.update(list => list.map(q => q.id === id ? { ...q, status: res.status } : q));
        }
      })
    );
  }

  listClients(): Observable<Customer[]> {
    return this.api.get<Customer[]>('/b2b/clients').pipe(
      tap(res => this.clients.set(res || []))
    );
  }

  getQuote(id: string): Observable<any> { return this.api.get<any>(`/b2b/quotes/${id}`); }
  convertQuoteToOrder(id: string): Observable<any> { return this.api.post<any>(`/b2b/quotes/${id}/convert`, {}); }
  bulkOrder(data: any): Observable<any> { return this.api.post<any>('/b2b/bulk-order', data); }
}
