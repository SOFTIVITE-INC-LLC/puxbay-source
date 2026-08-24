import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { BranchService } from './branch.service';
import { Observable, tap } from 'rxjs';

export interface DashboardTransaction {
  id: string;
  name: string;
  initials: string;
  amount: string;
  time: string;
  type: string;
}

export interface DashboardMetrics {
  total_sales: number;
  total_orders: number;
  active_customers: number;
  low_stock_items: number;
  sales_trend: number; // percentage
  revenue_chart: number[];
  recent_transactions: DashboardTransaction[];
}

export interface DashboardMetrics {
  today_revenue: number;
  today_orders: number;
  active_staff: number;
  low_stock_items: number;
  total_orders_alltime: number;
  total_revenue_alltime: number;
  revenue_chart: number[];
  recent_transactions: DashboardTransaction[];
}

@Injectable({
  providedIn: 'root'
})
export class DashboardService {
  private api = inject(ApiService);
  private branchService = inject(BranchService);

  metrics = signal<DashboardMetrics | null>(null);
  branchMetrics = signal<DashboardMetrics | null>(null);
  loading = signal<boolean>(false);

  getMetrics(): Observable<DashboardMetrics> {
    this.loading.set(true);
    return this.api.get<DashboardMetrics>('/analytics/dashboard').pipe(
      tap(res => {
        this.metrics.set(res);
        this.loading.set(false);
      })
    );
  }

  getBranchMetrics(branchId: string): Observable<DashboardMetrics> {
    this.loading.set(true);
    return this.api.get<DashboardMetrics>(`/branches/${branchId}/metrics`).pipe(
      tap(res => {
        this.branchMetrics.set(res);
        this.loading.set(false);
      })
    );
  }
}
