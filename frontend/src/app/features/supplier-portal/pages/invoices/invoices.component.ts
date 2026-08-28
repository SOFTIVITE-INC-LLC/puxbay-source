import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { SupplierPortalService, SupplierInvoice } from '../../services/supplier-portal.service';
import { AppCurrencyPipe } from '../../../../core/pipes/app-currency.pipe';
import { ToastService } from '../../../../core/services/toast';

@Component({
  selector: 'app-supplier-portal-invoices',
  standalone: true,
  imports: [CommonModule, DatePipe, AppCurrencyPipe],
  templateUrl: './invoices.component.html'
})
export class SupplierPortalInvoicesComponent implements OnInit {
  portalService = inject(SupplierPortalService);
  private toast = inject(ToastService);

  invoices = signal<SupplierInvoice[]>([]);
  loading = signal<boolean>(false);
  payoutLoading = signal<string | null>(null);

  ngOnInit() {
    this.loadInvoices();
  }

  loadInvoices() {
    this.loading.set(true);
    this.portalService.getInvoices().subscribe({
      next: (res) => {
        this.invoices.set(res || []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });
  }

  get totalOutstanding(): number {
    return this.invoices()
      .filter(i => i.status !== 'paid')
      .reduce((sum, i) => sum + (i.total - (i.amount_paid || 0)), 0);
  }

  get totalPaid(): number {
    return this.invoices()
      .reduce((sum, i) => sum + (i.amount_paid || 0), 0);
  }

  requestEarlyPayout(inv: SupplierInvoice) {
    if (!inv.id) return;
    this.payoutLoading.set(inv.id);
    this.portalService.initiateEarlyPayout(inv.id).subscribe({
      next: (res) => {
        this.toast.showSuccess(`Instant Settlement ${res.payout_ref} initiated for ${inv.invoice_number}!`);
        this.payoutLoading.set(null);
        this.loadInvoices();
      },
      error: (err) => {
        this.toast.showError(err.error?.error || 'Failed to process early settlement');
        this.payoutLoading.set(null);
      }
    });
  }

  statusClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s === 'paid') return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30';
    if (s === 'partially_paid' || s === 'pending') return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30';
    if (s === 'rejected') return 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30';
    return 'bg-slate-100 dark:bg-zinc-800 text-slate-700 dark:text-zinc-300 border-slate-200 dark:border-zinc-700';
  }
}
