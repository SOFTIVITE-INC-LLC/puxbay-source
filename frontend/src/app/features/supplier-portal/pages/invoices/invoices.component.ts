import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { SupplierPortalService, SupplierInvoice } from '../../services/supplier-portal.service';
import { AppCurrencyPipe } from '../../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-supplier-portal-invoices',
  standalone: true,
  imports: [CommonModule, DatePipe, AppCurrencyPipe],
  templateUrl: './invoices.component.html'
})
export class SupplierPortalInvoicesComponent implements OnInit {
  portalService = inject(SupplierPortalService);

  invoices = signal<SupplierInvoice[]>([]);
  loading = signal<boolean>(false);

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
      .reduce((sum, i) => sum + (i.total - i.amount_paid), 0);
  }

  get totalPaid(): number {
    return this.invoices()
      .reduce((sum, i) => sum + i.amount_paid, 0);
  }

  statusClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s === 'paid') return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
    if (s === 'partially_paid' || s === 'pending') return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
    if (s === 'rejected') return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
    return 'bg-zinc-800 text-zinc-300 border-zinc-700';
  }
}
