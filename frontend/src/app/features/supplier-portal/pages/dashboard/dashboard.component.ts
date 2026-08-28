import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { SupplierPortalService, DashboardStats, PurchaseOrder, DemandForecast, QRScanResult, SupplierAnnouncement } from '../../services/supplier-portal.service';
import { AppCurrencyPipe } from '../../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-supplier-portal-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule, AppCurrencyPipe, FormsModule],
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
  forecasts = signal<DemandForecast[]>([]);
  loading = signal<boolean>(true);

  // Announcements
  announcements = signal<SupplierAnnouncement[]>([]);
  dismissedAnnouncements = new Set<string>();

  // QR Scanner
  showScanner = signal<boolean>(false);
  qrInput = signal<string>('');
  scanResult = signal<QRScanResult | null>(null);
  scanLoading = signal<boolean>(false);

  ngOnInit() {
    this.loadData();
  }

  dismissAnnouncement(id: string) {
    this.dismissedAnnouncements.add(id);
    this.announcements.update(anns => anns.filter(a => a.id !== id));
  }

  loadData() {
    this.loading.set(true);
    this.portalService.getDashboard().subscribe({
      next: (res) => {
        if (res) this.stats.set(res);
      }
    });

    this.portalService.getAnnouncements().subscribe({
      next: (anns) => this.announcements.set(anns || [])
    });

    this.portalService.getPurchaseOrders().subscribe({
      next: (orders) => {
        this.recentOrders.set((orders || []).slice(0, 5));
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });

    this.portalService.getForecasts().subscribe({
      next: (data) => this.forecasts.set(data || []),
      error: () => {}
    });
  }

  submitQRScan() {
    if (!this.qrInput()) return;
    this.scanLoading.set(true);
    this.portalService.submitQRScan(this.qrInput()).subscribe({
      next: (res) => {
        this.scanResult.set(res);
        this.scanLoading.set(false);
      },
      error: () => this.scanLoading.set(false)
    });
  }

  clearScan() {
    this.qrInput.set('');
    this.scanResult.set(null);
  }

  urgencyClass(urgency: string): string {
    if (urgency === 'high') return 'text-rose-400 bg-rose-500/10 border-rose-500/30';
    if (urgency === 'medium') return 'text-amber-400 bg-amber-500/10 border-amber-500/30';
    return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/30';
  }

  statusClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s === 'received' || s === 'confirmed') return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
    if (s === 'partially_received' || s === 'issued') return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
    if (s === 'cancelled' || s === 'rejected') return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
    return 'bg-zinc-800 text-zinc-300 border-zinc-700';
  }
}
