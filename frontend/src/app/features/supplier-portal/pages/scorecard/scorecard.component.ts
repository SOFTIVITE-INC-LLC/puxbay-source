import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SupplierPortalService, DashboardStats, SupplierScorecard } from '../../services/supplier-portal.service';

@Component({
  selector: 'app-supplier-portal-scorecard',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './scorecard.component.html'
})
export class SupplierPortalScorecardComponent implements OnInit {
  portalService = inject(SupplierPortalService);

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

    this.loadingTier.set(true);
    this.portalService.getScorecard().subscribe({
      next: (res) => {
        this.scorecard.set(res);
        this.loadingTier.set(false);
      },
      error: () => this.loadingTier.set(false)
    });
  }

  tierBadgeClass(tier: string = ''): string {
    const t = tier.toLowerCase();
    if (t === 'platinum') return 'bg-gradient-to-r from-violet-600 to-purple-700 text-white shadow-lg shadow-violet-500/30';
    if (t === 'gold') return 'bg-gradient-to-r from-amber-400 to-yellow-500 text-white shadow-lg shadow-amber-500/30';
    if (t === 'silver') return 'bg-gradient-to-r from-slate-400 to-slate-500 text-white shadow-md';
    return 'bg-gradient-to-r from-amber-700 to-orange-700 text-white shadow-md';
  }

  tierIcon(tier: string = ''): string {
    const t = tier.toLowerCase();
    if (t === 'platinum') return '🏆';
    if (t === 'gold') return '🥇';
    if (t === 'silver') return '🥈';
    return '🥉';
  }
}
