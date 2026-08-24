import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

export interface DiningTable {
  id: string;
  name: string;
  capacity: number;
  status: string;
  is_active: boolean;
}

export interface KDSTicket {
  id: string;
  order_id: string;
  table_id: string;
  status: string;
  started_at?: string;
  completed_at?: string;
}

@Injectable({
  providedIn: 'root'
})
export class FNBService {
  createTable(t: any): Observable<any> { return this.api.post('/fnb/tables', t); }

  private api = inject(ApiService);
  
  tables = signal<DiningTable[]>([]);
  tickets = signal<KDSTicket[]>([]);
  loading = signal<boolean>(false);

  getTables(): Observable<DiningTable[]> {
    this.loading.set(true);
    return this.api.get<DiningTable[]>('/fnb/tables').pipe(
      tap(tables => {
        this.tables.set(tables);
        this.loading.set(false);
      })
    );
  }

  updateTableStatus(id: string, status: string): Observable<{status: string}> {
    return this.api.put<{status: string}>(`/fnb/tables/${id}/status`, { status }).pipe(
      tap(res => {
        this.tables.update(list => list.map(t => t.id === id ? { ...t, status: res.status } : t));
      })
    );
  }

  getKDS(): Observable<KDSTicket[]> {
    return this.api.get<KDSTicket[]>('/fnb/kds').pipe(
      tap(tickets => this.tickets.set(tickets))
    );
  }

  advanceTicketStatus(id: string): Observable<{new_status: string}> {
    return this.api.put<{new_status: string}>(`/fnb/kds/${id}/advance`, {}).pipe(
      tap(res => {
        this.tickets.update(list => list.map(t => t.id === id ? { ...t, status: res.new_status } : t));
      })
    );
  }

  createSplitBill(data: any): Observable<any> { return this.api.post<any>('/fnb/split-bills', data); }
  getSplitBill(id: string): Observable<any> { return this.api.get<any>(`/fnb/split-bills/${id}`); }
}
