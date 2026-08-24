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
  current_price: number;
  suggested_price: number;
  reason: string;
}

@Injectable({ providedIn: 'root' })
export class IntelligenceService {
  private api = inject(ApiService);

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

  getDynamicPricing() {
    this.pricingLoading.set(true);
    return this.api.get<{ suggestions: PricingSuggestion[] }>('/intelligence/dynamic-pricing').pipe(
      tap(res => {
        this.pricingSuggestions.set(res?.suggestions || []);
        this.pricingLoading.set(false);
      })
    );
  }
}
