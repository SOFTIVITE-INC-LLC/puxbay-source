import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { tap } from 'rxjs';

export interface InventoryForecast {
  product: {
    id: string;
    name: string;
    sku: string;
    current_stock: number;
    reorder_level: number;
    category?: { name: string };
  };
  days_left: number;
  velocity: number;
  status: 'healthy' | 'warning' | 'critical';
}

export interface LeaderboardEntry {
  cashier_id: string;
  first_name: string;
  last_name: string;
  total_sales: number;
  order_count: number;
}

export interface CustomerSegments {
  VIP: { count: number };
  Loyal: { count: number };
  Recent: { count: number };
  'At Risk': { count: number };
  Lost: { count: number };
}

export interface PricingSuggestion {
  product_id: string;
  product_name: string;
  sku: string;
  category_name: string;
  current_price: number;
  cost_price: number;
  suggested_price: number;
  change_percent: number;
  strategy: string;
  reason: string;
  velocity: number;
  current_stock: number;
}

export interface AnomalyAlert {
  id: string;
  type: string;
  severity: 'warning' | 'critical';
  title: string;
  description: string;
  metric: string;
  baseline: number;
  actual: number;
  deviation: number;
  detected_at: string;
}

@Injectable({ providedIn: 'root' })
export class IntelligenceService {
  private api = inject(ApiService);

  // Anomaly Alerts
  anomalies = signal<AnomalyAlert[]>([]);
  anomaliesLoading = signal(false);

  // Inventory Forecast
  forecasts = signal<InventoryForecast[]>([]);
  forecastsLoading = signal(false);

  // Staff Leaderboard
  leaderboard = signal<LeaderboardEntry[]>([]);
  leaderboardLoading = signal(false);

  // Customer Segmentation
  segments = signal<CustomerSegments | null>(null);
  totalCustomers = signal(0);
  segmentsLoading = signal(false);

  // Dynamic Pricing
  pricingSuggestions = signal<PricingSuggestion[]>([]);
  pricingLoading = signal(false);

  // Legacy
  insights = signal<any[]>([]);
  loading = signal(false);

  getInsights() {
    return this.api.get<any[]>('/intelligence/insights').pipe(tap(res => this.insights.set(res || [])));
  }

  getAnomalies() {
    this.anomaliesLoading.set(true);
    return this.api.get<{ anomalies: AnomalyAlert[] }>('/intelligence/anomalies').pipe(
      tap(res => {
        this.anomalies.set(res?.anomalies || []);
        this.anomaliesLoading.set(false);
      })
    );
  }

  getInventoryForecast(branchId?: string) {
    this.forecastsLoading.set(true);
    const params = branchId ? `?branch_id=${branchId}` : '';
    return this.api.get<{ forecasts: InventoryForecast[] }>(`/intelligence/inventory-forecast${params}`).pipe(
      tap(res => {
        this.forecasts.set(res?.forecasts || []);
        this.forecastsLoading.set(false);
      })
    );
  }

  getStaffLeaderboard(days = 30, branchId?: string) {
    this.leaderboardLoading.set(true);
    const params = new URLSearchParams({ days: String(days) });
    if (branchId) params.set('branch_id', branchId);
    return this.api.get<{ leaderboard: LeaderboardEntry[] }>(`/intelligence/staff-leaderboard?${params}`).pipe(
      tap(res => {
        this.leaderboard.set(res?.leaderboard || []);
        this.leaderboardLoading.set(false);
      })
    );
  }

  getCustomerSegmentation() {
    this.segmentsLoading.set(true);
    return this.api.get<{ segments: CustomerSegments; total_customers: number }>('/intelligence/customer-segmentation').pipe(
      tap(res => {
        this.segments.set(res?.segments || null);
        this.totalCustomers.set(res?.total_customers || 0);
        this.segmentsLoading.set(false);
      })
    );
  }

  getDynamicPricing(branchId?: string) {
    this.pricingLoading.set(true);
    const params = branchId ? `?branch_id=${branchId}` : '';
    return this.api.get<{ suggestions: PricingSuggestion[] }>(`/intelligence/dynamic-pricing${params}`).pipe(
      tap(res => {
        this.pricingSuggestions.set(res?.suggestions || []);
        this.pricingLoading.set(false);
      })
    );
  }

  applyPricingAction(productId: string, newPrice: number) {
    return this.api.post<{ status: string; message: string }>('/intelligence/dynamic-pricing/apply', {
      product_id: productId,
      new_price: newPrice
    });
  }

  bulkApplyPricing(items: { product_id: string; new_price: number }[]) {
    return this.api.post<{ status: string; updated_count: number; message: string }>('/intelligence/dynamic-pricing/apply-bulk', {
      items
    });
  }
}
