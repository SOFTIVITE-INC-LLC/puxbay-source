import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { IntelligenceService, PricingSuggestion } from '../../../core/services/intelligence.service';
import { ToastService } from '../../../core/services/toast';

@Component({
  selector: 'app-intelligence',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './intelligence.html',
  styles: `
    @keyframes shimmer {
      0% { background-position: -1000px 0; }
      100% { background-position: 1000px 0; }
    }
    @keyframes pulse-ring {
      0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.4); }
      70% { transform: scale(1); box-shadow: 0 0 0 10px rgba(16, 185, 129, 0); }
      100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
    }

    .shimmer-bar {
      background: #005b96;
      background-size: 1000px 100%;
      animation: shimmer 1.5s infinite;
    }
    .dark .shimmer-bar {
      background: #005b96;
      background-size: 1000px 100%;
    }
    .animated-gradient { background-color: #011f4b; }
    .glass {
      background: rgba(255,255,255,0.65);
      backdrop-filter: blur(18px);
      -webkit-backdrop-filter: blur(18px);
      border: 1px solid rgba(255,255,255,0.5);
    }
    :host-context(.dark) .glass {
      background: rgba(24,24,27,0.65);
      border: 1px solid rgba(255,255,255,0.07);
    }
    .velocity-bar { transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1); }
  `
})
export class Intelligence implements OnInit {
  svc = inject(IntelligenceService);
  private toast = inject(ToastService);
  Math = Math;

  activeTab = signal<'forecast' | 'leaderboard' | 'customers' | 'pricing' | 'anomalies'>('forecast');
  leaderboardDays = signal(30);

  // Dynamic Pricing state
  pricingStrategyFilter = signal<string>('all');
  pricingSearchQuery = signal<string>('');
  applyingProductId = signal<string | null>(null);
  bulkApplying = signal<boolean>(false);

  criticalCount = computed(() => this.svc.forecasts().filter(f => f.status === 'critical').length);
  warningCount = computed(() => this.svc.forecasts().filter(f => f.status === 'warning').length);
  healthyCount = computed(() => this.svc.forecasts().filter(f => f.status === 'healthy').length);

  // Anomaly Alerts computed stats
  anomalyCount = computed(() => this.svc.anomalies().length);
  criticalAnomalyCount = computed(() => this.svc.anomalies().filter(a => a.severity === 'critical').length);
  warningAnomalyCount = computed(() => this.svc.anomalies().filter(a => a.severity === 'warning').length);

  // Dynamic Pricing computed stats
  surgeCount = computed(() => this.svc.pricingSuggestions().filter(s => s.strategy === 'surge').length);
  clearanceCount = computed(() => this.svc.pricingSuggestions().filter(s => s.strategy === 'clearance').length);
  marginRecoveryCount = computed(() => this.svc.pricingSuggestions().filter(s => s.strategy === 'margin_recovery').length);
  overstockCount = computed(() => this.svc.pricingSuggestions().filter(s => s.strategy === 'overstock').length);

  filteredPricingSuggestions = computed(() => {
    let list = this.svc.pricingSuggestions();
    const strat = this.pricingStrategyFilter();
    if (strat !== 'all') {
      list = list.filter(s => s.strategy === strat);
    }
    const q = this.pricingSearchQuery().toLowerCase().trim();
    if (q) {
      list = list.filter(s =>
        (s.product_name && s.product_name.toLowerCase().includes(q)) ||
        (s.sku && s.sku.toLowerCase().includes(q)) ||
        (s.category_name && s.category_name.toLowerCase().includes(q))
      );
    }
    return list;
  });

  segmentEntries = computed(() => {
    const s = this.svc.segments();
    if (!s) return [];
    return [
      { label: 'VIP', count: s['VIP']?.count ?? 0, color: 'from-amber-500 ', icon: 'workspace_premium', desc: 'High spend, frequent buyers' },
      { label: 'Loyal', count: s['Loyal']?.count ?? 0, color: 'from-emerald-500 ', icon: 'favorite', desc: 'Regular, committed customers' },
      { label: 'Recent', count: s['Recent']?.count ?? 0, color: 'from-blue-500 ', icon: 'person_add', desc: 'New or recent buyers' },
      { label: 'At Risk', count: (s as any)['At Risk']?.count ?? 0, color: 'from-orange-500 ', icon: 'warning', desc: 'Haven\'t bought in 30–90 days' },
      { label: 'Lost', count: s['Lost']?.count ?? 0, color: 'from-rose-500 ', icon: 'person_off', desc: 'Inactive over 90 days' },
    ];
  });

  maxSales = computed(() => {
    const l = this.svc.leaderboard();
    return l.length ? Math.max(...l.map(e => e.total_sales)) : 1;
  });

  ngOnInit() {
    this.loadTab('forecast');
    this.svc.getCustomerSegmentation().subscribe();
    this.svc.getAnomalies().subscribe();
  }

  loadTab(tab: 'forecast' | 'leaderboard' | 'customers' | 'pricing' | 'anomalies') {
    this.activeTab.set(tab);
    switch (tab) {
      case 'forecast':
        this.svc.getInventoryForecast().subscribe();
        break;
      case 'leaderboard':
        this.svc.getStaffLeaderboard(this.leaderboardDays()).subscribe();
        break;
      case 'customers':
        this.svc.getCustomerSegmentation().subscribe();
        break;
      case 'pricing':
        this.svc.getDynamicPricing().subscribe();
        break;
      case 'anomalies':
        this.svc.getAnomalies().subscribe();
        break;
    }
  }

  changeLeaderboardDays(days: number) {
    this.leaderboardDays.set(days);
    this.svc.getStaffLeaderboard(days).subscribe();
  }

  refreshAll() {
    this.svc.getInventoryForecast().subscribe();
    this.svc.getStaffLeaderboard(this.leaderboardDays()).subscribe();
    this.svc.getCustomerSegmentation().subscribe();
    this.svc.getDynamicPricing().subscribe();
    this.svc.getAnomalies().subscribe();
  }

  applySinglePricing(suggestion: PricingSuggestion) {
    this.applyingProductId.set(suggestion.product_id);
    this.svc.applyPricingAction(suggestion.product_id, suggestion.suggested_price).subscribe({
      next: () => {
        this.toast.showSuccess(`Updated ${suggestion.product_name} price to ${suggestion.suggested_price}!`);
        this.applyingProductId.set(null);
        // Refresh pricing recommendations
        this.svc.getDynamicPricing().subscribe();
      },
      error: (err) => {
        this.toast.showError(err?.error?.error || 'Failed to update price');
        this.applyingProductId.set(null);
      }
    });
  }

  applyBulkPricing() {
    const items = this.filteredPricingSuggestions().map(s => ({
      product_id: s.product_id,
      new_price: s.suggested_price
    }));
    if (!items.length) return;

    this.bulkApplying.set(true);
    this.svc.bulkApplyPricing(items).subscribe({
      next: (res) => {
        this.toast.showSuccess(`Successfully updated ${res.updated_count || items.length} product prices!`);
        this.bulkApplying.set(false);
        this.svc.getDynamicPricing().subscribe();
      },
      error: (err) => {
        this.toast.showError(err?.error?.error || 'Failed to bulk apply prices');
        this.bulkApplying.set(false);
      }
    });
  }

  strategyBadge(strategy: string): { label: string; class: string; icon: string } {
    switch (strategy) {
      case 'surge':
        return { label: 'High Demand Surge', class: 'bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 border border-indigo-500/20', icon: 'trending_up' };
      case 'clearance':
        return { label: 'Clearance Discount', class: 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20', icon: 'local_offer' };
      case 'margin_recovery':
        return { label: 'Margin Recovery', class: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20', icon: 'price_check' };
      case 'overstock':
        return { label: 'Overstock Markdown', class: 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/20', icon: 'inventory_2' };
      default:
        return { label: 'Competitive Tuning', class: 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/20', icon: 'tune' };
    }
  }

  statusLabel(status: string): string {
    return { critical: 'Critical', warning: 'Low Stock', healthy: 'Healthy' }[status] ?? status;
  }

  daysLabel(days: number): string {
    if (days >= 999) return '999+ days';
    return `${Math.round(days)} days`;
  }

  priceDelta(current: number, suggested: number): string {
    const pct = ((suggested - current) / current) * 100;
    return (pct >= 0 ? '+' : '') + pct.toFixed(1) + '%';
  }
}
