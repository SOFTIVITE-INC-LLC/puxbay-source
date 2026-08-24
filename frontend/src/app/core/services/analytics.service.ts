import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

export interface DailyData {
  date: string;
  revenue: number;
  orders: number;
}

export interface SalesTrendsResult {
  current_revenue: number;
  current_orders: number;
  previous_revenue: number;
  previous_orders: number;
  revenue_growth: number;
  order_growth: number;
  daily_data: DailyData[];
  period: string;
}

export interface PaymentData {
  method: string;
  revenue: number;
  count: number;
}

export interface CategoryData {
  name: string;
  revenue: number;
}

export interface RevenueBreakdownResult {
  by_category: CategoryData[];
  by_payment_method: PaymentData[];
}

export interface TopProductData {
  product_id: string;
  name: string;
  sku: string;
  quantity: number;
  revenue: number;
}

export interface TopProductsResult {
  by_quantity: TopProductData[];
  by_revenue: TopProductData[];
}

export interface CustomerMetricsResult {
  total_customers: number;
  active_customers: number;
  avg_order_value: number;
}

export interface RealTimeMetricsResult {
  today_revenue: number;
  today_orders: number;
  inventory_value: number;
  low_stock_count: number;
  out_of_stock_count: number;
  total_products: number;
}

export interface StaffPerformanceData {
  staff_id: string;
  staff_name: string;
  revenue: number;
  orders_count: number;
}

export interface SalesGoalProgress {
  goal: number;
  current: number;
  progress: number;
}

export interface CustomReportResult {
  headers: string[];
  rows: any[];
}

@Injectable({
  providedIn: 'root'
})
export class AnalyticsService {
  private api = inject(ApiService);
  
  loading = signal<boolean>(false);

  getSalesTrends(params?: any): Observable<SalesTrendsResult> {
    this.loading.set(true);
    return this.api.get<SalesTrendsResult>('/analytics/sales', { params }).pipe(
      tap(() => this.loading.set(false))
    );
  }

  getRevenueBreakdown(params?: any): Observable<RevenueBreakdownResult> {
    return this.api.get<RevenueBreakdownResult>('/analytics/revenue', { params });
  }

  getTopProducts(params?: any): Observable<TopProductsResult> {
    return this.api.get<TopProductsResult>('/analytics/top-products', { params });
  }

  getCustomerMetrics(): Observable<CustomerMetricsResult> {
    return this.api.get<CustomerMetricsResult>('/analytics/customers');
  }

  getRealTimeMetrics(params?: any): Observable<RealTimeMetricsResult> {
    return this.api.get<RealTimeMetricsResult>('/analytics/realtime', { params });
  }

  salesHeatmap(): Observable<any> { return this.api.get<any>('/analytics/heatmap'); }
  
  getStaffPerformance(params?: any): Observable<StaffPerformanceData[]> {
    return this.api.get<StaffPerformanceData[]>('/analytics/staff-performance', { params });
  }

  getSalesGoalProgress(): Observable<SalesGoalProgress> {
    return this.api.get<SalesGoalProgress>('/analytics/sales-goal');
  }

  reportBuilder(data: any): Observable<CustomReportResult> {
    return this.api.post<CustomReportResult>('/analytics/report-builder', data);
  }
}
