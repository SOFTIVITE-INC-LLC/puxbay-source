import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

export interface CashDrawer {
  id: string;
  status: string;
  opening_balance: number;
  closing_balance?: number;
  expected_balance: number;
  actual_balance?: number;
  float_amount?: number;
  opened_at: string;
  closed_at?: string;
  opened_by?: string;
  notes?: string;
}

export interface ShiftSummary {
  total_sales: number;
  total_cash: number;
  total_card: number;
  total_refunds: number;
  net_cash: number;
  expected_in_drawer: number;
}

@Injectable({
  providedIn: 'root'
})
export class CashDrawerService {
  private api = inject(ApiService);
  
  drawers = signal<CashDrawer[]>([]);
  activeDrawer = signal<CashDrawer | null>(null);
  shiftSummary = signal<ShiftSummary | null>(null);
  loading = signal<boolean>(false);

  getDrawers(): Observable<{cash_drawers: CashDrawer[]}> {
    this.loading.set(true);
    return this.api.get<{cash_drawers: CashDrawer[]}>('/cash-drawers').pipe(
      tap(res => {
        this.drawers.set(res.cash_drawers || []);
        const open = (res.cash_drawers || []).find(d => d.status === 'open');
        this.activeDrawer.set(open || null);
        this.loading.set(false);
      })
    );
  }

  openShift(openingBalance: number, notes?: string): Observable<CashDrawer> {
    return this.api.post<CashDrawer>('/cash-drawers/open', { opening_balance: openingBalance, notes }).pipe(
      tap(drawer => {
        this.drawers.update(d => [drawer, ...d]);
        this.activeDrawer.set(drawer);
      })
    );
  }

  closeShift(drawerId: string, actualBalance: number, notes?: string): Observable<{drawer: CashDrawer, summary: ShiftSummary}> {
    return this.api.post<{drawer: CashDrawer, summary: ShiftSummary}>(`/cash-drawers/${drawerId}/close`, { 
      actual_balance: actualBalance, notes 
    }).pipe(
      tap(res => {
        this.drawers.update(d => d.map(dr => dr.id === drawerId ? res.drawer : dr));
        this.activeDrawer.set(null);
        this.shiftSummary.set(res.summary);
      })
    );
  }

  addFloat(drawerId: string, amount: number, notes?: string): Observable<CashDrawer> {
    return this.api.post<CashDrawer>(`/cash-drawers/${drawerId}/float`, { amount, type: 'add', notes });
  }

  removeFloat(drawerId: string, amount: number, notes?: string): Observable<CashDrawer> {
    return this.api.post<CashDrawer>(`/cash-drawers/${drawerId}/float`, { amount, type: 'remove', notes });
  }
}
