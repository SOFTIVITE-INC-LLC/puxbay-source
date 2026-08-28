import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { SupplierPortalService, DashboardStats, PurchaseOrder } from '../../services/supplier-portal.service';
import { AppCurrencyPipe } from '../../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-supplier-portal-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule, AppCurrencyPipe],
  templateUrl: './dashboard.component.html'
})
export class SupplierPortalDashboardComponent implements OnInit {
  portalService = inject(SupplierPortalService);

  stats = signal<DashboardStats>({
    total_pos: 0,
    pending_deliveries: 0,
    total_invoiced: 0,
    open_quotes: 0,
    otd_score: 98.5
  });

  recentOrders = signal<PurchaseOrder[]>([]);
  loading = signal<boolean>(true);

  ngOnInit() {
    this.loadData();
  }

  loadData() {
    this.loading.set(true);
    this.portalService.getDashboard().subscribe({
      next: (res) => {
        if (res) this.stats.set(res);
      }
    });

    this.portalService.getPurchaseOrders().subscribe({
      next: (orders) => {
        this.recentOrders.set((orders || []).slice(0, 5));
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });
  }

  statusClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s === 'received' || s === 'confirmed') return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
    if (s === 'partially_received' || s === 'issued') return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
    if (s === 'cancelled' || s === 'rejected') return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
    return 'bg-zinc-800 text-zinc-300 border-zinc-700';
  }
}
