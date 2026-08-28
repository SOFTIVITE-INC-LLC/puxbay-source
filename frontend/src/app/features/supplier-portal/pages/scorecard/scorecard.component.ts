import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SupplierPortalService, DashboardStats, SupplierScorecard, TierLevel, QuarterlyReview } from '../../services/supplier-portal.service';

@Component({
  selector: 'app-supplier-portal-scorecard',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './scorecard.component.html'
})
export class SupplierPortalScorecardComponent implements OnInit {
  portalService = inject(SupplierPortalService);

  activeTab = signal<'overview' | 'matrix' | 'history' | 'sla'>('overview');
  selectedPeriod = signal<'ytd' | '90d' | '30d' | 'all'>('ytd');

  stats = signal<DashboardStats>({
    total_pos: 0,
    pending_deliveries: 0,
    total_invoiced: 0,
    open_quotes: 0,
    otd_score: 98.5
  });

  scorecard = signal<SupplierScorecard | null>(null);
  loadingTier = signal<boolean>(false);

  ngOnInit() {
    this.portalService.getDashboard().subscribe({
      next: (res) => { if (res) this.stats.set(res); }
    });

    this.loadScorecard();
  }

  loadScorecard() {
    this.loadingTier.set(true);
    this.portalService.getScorecard().subscribe({
      next: (res) => {
        this.scorecard.set(res);
        this.loadingTier.set(false);
      },
      error: () => this.loadingTier.set(false)
    });
  }

  printScorecard() {
    window.print();
  }

  tierBadgeClass(tier: string = ''): string {
    const t = tier.toLowerCase();
    if (t === 'platinum') return 'bg-gradient-to-r from-violet-600 to-purple-700 text-white shadow-lg shadow-violet-500/30 border border-violet-400/40';
    if (t === 'gold') return 'bg-gradient-to-r from-amber-400 to-yellow-500 text-white shadow-lg shadow-amber-500/30 border border-amber-300/40';
    if (t === 'silver') return 'bg-gradient-to-r from-slate-400 to-slate-500 text-white shadow-md border border-slate-300/40';
    return 'bg-gradient-to-r from-amber-700 to-orange-700 text-white shadow-md border border-amber-600/40';
  }

  tierIcon(tier: string = ''): string {
    const t = tier.toLowerCase();
    if (t === 'platinum') return '🏆';
    if (t === 'gold') return '🥇';
    if (t === 'silver') return '🥈';
    return '🥉';
  }

  statusScoreClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s.includes('exceed')) return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30';
    if (s.includes('met')) return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/30';
    return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30';
  }
}
