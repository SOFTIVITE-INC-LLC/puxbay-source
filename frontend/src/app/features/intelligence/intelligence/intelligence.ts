import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { IntelligenceService } from '../../../core/services/intelligence.service';

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
    @keyframes gradientShift {
      0% { background-position: 0% 50%; }
      50% { background-position: 100% 50%; }
      100% { background-position: 0% 50%; }
    }
    .shimmer-bar {
      background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
      background-size: 1000px 100%;
      animation: shimmer 1.5s infinite;
    }
    .dark .shimmer-bar {
      background: linear-gradient(90deg, #27272a 25%, #3f3f46 50%, #27272a 75%);
      background-size: 1000px 100%;
    }
    .animated-gradient {
      background-size: 300% 300%;
      animation: gradientShift 8s ease infinite;
    }
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
  Math = Math;

  activeTab = signal<'forecast' | 'leaderboard' | 'customers' | 'pricing'>('forecast');
  leaderboardDays = signal(30);

  criticalCount = computed(() => this.svc.forecasts().filter(f => f.status === 'critical').length);
  warningCount = computed(() => this.svc.forecasts().filter(f => f.status === 'warning').length);
  healthyCount = computed(() => this.svc.forecasts().filter(f => f.status === 'healthy').length);

  segmentEntries = computed(() => {
    const s = this.svc.segments();
    if (!s) return [];
    return [
      { label: 'VIP', count: s['VIP']?.count ?? 0, color: 'from-amber-500 to-yellow-400', icon: 'workspace_premium', desc: 'High spend, frequent buyers' },
      { label: 'Loyal', count: s['Loyal']?.count ?? 0, color: 'from-emerald-500 to-teal-400', icon: 'favorite', desc: 'Regular, committed customers' },
      { label: 'Recent', count: s['Recent']?.count ?? 0, color: 'from-blue-500 to-indigo-400', icon: 'person_add', desc: 'New or recent buyers' },
      { label: 'At Risk', count: (s as any)['At Risk']?.count ?? 0, color: 'from-orange-500 to-amber-400', icon: 'warning', desc: 'Haven\'t bought in 30–90 days' },
      { label: 'Lost', count: s['Lost']?.count ?? 0, color: 'from-rose-500 to-red-400', icon: 'person_off', desc: 'Inactive over 90 days' },
    ];
  });

  maxSales = computed(() => {
    const l = this.svc.leaderboard();
    return l.length ? Math.max(...l.map(e => e.total_sales)) : 1;
  });

  ngOnInit() {
    this.loadTab('forecast');
    this.svc.getCustomerSegmentation().subscribe();
  }

  loadTab(tab: 'forecast' | 'leaderboard' | 'customers' | 'pricing') {
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
