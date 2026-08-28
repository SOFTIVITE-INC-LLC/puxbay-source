import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SupplierPortalService, SupplierQuote } from '../../services/supplier-portal.service';
import { AppCurrencyPipe } from '../../../../core/pipes/app-currency.pipe';
import { ToastService } from '../../../../core/services/toast';

@Component({
  selector: 'app-supplier-portal-rfqs',
  standalone: true,
  imports: [CommonModule, DatePipe, AppCurrencyPipe, FormsModule],
  templateUrl: './rfqs.component.html'
})
export class SupplierPortalRfqsComponent implements OnInit {
  portalService = inject(SupplierPortalService);
  private toast = inject(ToastService);

  quotes = signal<SupplierQuote[]>([]);
  loading = signal<boolean>(false);

  showCreateModal = signal<boolean>(false);
  quoteTitle = '';
  totalAmount = 0;
  currency = 'USD';
  validUntil = '';
  leadTimeDays = 7;
  paymentTerms = 'Net 30';
  notes = '';

  ngOnInit() {
    this.loadQuotes();
  }

  loadQuotes() {
    this.loading.set(true);
    this.portalService.getQuotes().subscribe({
      next: (res) => {
        this.quotes.set(res || []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });
  }

  openCreateModal() {
    this.quoteTitle = '';
    this.totalAmount = 0;
    this.currency = 'USD';
    const date = new Date();
    date.setDate(date.getDate() + 30);
    this.validUntil = date.toISOString().split('T')[0];
    this.leadTimeDays = 7;
    this.paymentTerms = 'Net 30';
    this.notes = '';
    this.showCreateModal.set(true);
  }

  submitQuote() {
    if (!this.quoteTitle || this.totalAmount <= 0) {
      this.toast.showError('Please provide a title and total quote amount');
      return;
    }

    const payload: Partial<SupplierQuote> = {
      title: this.quoteTitle,
      total_amount: Number(this.totalAmount),
      currency: this.currency,
      valid_until: new Date(this.validUntil).toISOString(),
      lead_time_days: Number(this.leadTimeDays) || 7,
      payment_terms: this.paymentTerms,
      notes: this.notes
    };

    this.portalService.createQuote(payload).subscribe({
      next: () => {
        this.toast.showSuccess('Quotation submitted successfully!');
        this.showCreateModal.set(false);
        this.loadQuotes();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to submit quote')
    });
  }

  statusClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s === 'accepted') return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
    if (s === 'submitted') return 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20';
    if (s === 'rejected' || s === 'expired') return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
    return 'bg-zinc-800 text-zinc-300 border-zinc-700';
  }
}
