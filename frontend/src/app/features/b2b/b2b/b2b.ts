import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule, CurrencyPipe, DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { B2bService } from '../../../core/services/b2b.service';
import { CatalogService } from '../../../core/services/catalog.service';
import { CustomerService } from '../../../core/services/customer.service';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-b2b',
  standalone: true,
  imports: [CommonModule, FormsModule, DatePipe, AppCurrencyPipe],
  templateUrl: './b2b.html',
})
export class B2b implements OnInit {
  b2bService = inject(B2bService);
  catalogService = inject(CatalogService);
  customerService = inject(CustomerService);
  private toastr = inject(ToastrService);

  isQuoteModalOpen = signal(false);
  isViewModalOpen = signal(false);
  isConverting = signal(false);
  selectedQuote = signal<any>(null);
  filterStatus = signal<string>('all');

  newQuote = signal<{ customer_id: string; notes: string; items: { product_id: string; qty: number; unit_price: number }[] }>({
    customer_id: '',
    notes: '',
    items: [{ product_id: '', qty: 1, unit_price: 0 }]
  });

  quoteTotal = computed(() =>
    this.newQuote().items.reduce((s, i) => s + (i.qty * i.unit_price), 0)
  );

  // Computed stats
  totalQuotes = computed(() => this.b2bService.quotes().length);
  pendingQuotes = computed(() => this.b2bService.quotes().filter(q => q.status === 'pending').length);
  approvedQuotes = computed(() => this.b2bService.quotes().filter(q => q.status === 'approved').length);
  convertedQuotes = computed(() => this.b2bService.quotes().filter(q => q.status === 'converted').length);

  filteredQuotes = computed(() => {
    const status = this.filterStatus();
    if (status === 'all') return this.b2bService.quotes();
    return this.b2bService.quotes().filter(q => q.status === status);
  });

  ngOnInit() {
    this.b2bService.listQuotes().subscribe();
    this.b2bService.listClients().subscribe();
    this.customerService.getCustomers().subscribe();
    this.catalogService.getProducts().subscribe();
  }

  addItem() {
    this.newQuote.update(q => ({
      ...q,
      items: [...q.items, { product_id: '', qty: 1, unit_price: 0 }]
    }));
  }

  removeItem(i: number) {
    this.newQuote.update(q => ({ ...q, items: q.items.filter((_, idx) => idx !== i) }));
  }

  onProductSelect(index: number, productId: string) {
    const product = this.catalogService.products().find(p => p.id === productId);
    this.newQuote.update(q => ({
      ...q,
      items: q.items.map((item, i) => i === index
        ? { ...item, product_id: productId, unit_price: (product as any)?.wholesale_price || product?.selling_price || 0 }
        : item)
    }));
  }

  createQuote() {
    const q = this.newQuote();
    if (!q.customer_id || q.items.some(i => !i.product_id)) {
      this.toastr.error('Please fill all required fields.');
      return;
    }
    this.b2bService.createQuote({ customer_id: q.customer_id, items: q.items.map(i => ({ product_id: i.product_id, qty: i.qty })) })
      .subscribe({
        next: () => {
          this.toastr.success('Quote created!');
          this.isQuoteModalOpen.set(false);
          this.b2bService.listQuotes().subscribe();
          this.newQuote.set({ customer_id: '', notes: '', items: [{ product_id: '', qty: 1, unit_price: 0 }] });
        },
        error: () => this.toastr.error('Failed to create quote.')
      });
  }

  viewQuote(quote: any) {
    this.selectedQuote.set(quote);
    this.isViewModalOpen.set(true);
  }

  convertToOrder(quoteId: string) {
    this.isConverting.set(true);
    this.b2bService.convertQuoteToOrder(quoteId).subscribe({
      next: () => {
        this.toastr.success('Quote converted to order!');
        this.isViewModalOpen.set(false);
        this.isConverting.set(false);
        this.b2bService.listQuotes().subscribe();
      },
      error: () => {
        this.toastr.error('Conversion failed.');
        this.isConverting.set(false);
      }
    });
  }

  approveQuote(quoteId: string) {
    this.b2bService.updateQuote(quoteId, { action: 'approve' }).subscribe({
      next: () => {
        this.toastr.success('Quote approved!');
        this.b2bService.listQuotes().subscribe();
      },
      error: () => this.toastr.error('Action failed.')
    });
  }

  rejectQuote(quoteId: string) {
    this.b2bService.updateQuote(quoteId, { action: 'reject' }).subscribe({
      next: () => {
        this.toastr.success('Quote rejected.');
        this.b2bService.listQuotes().subscribe();
      },
      error: () => this.toastr.error('Action failed.')
    });
  }

  statusBadge(status: string): string {
    return {
      pending: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
      approved: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300',
      rejected: 'bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-300',
      converted: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
    }[status] ?? 'bg-zinc-100 text-zinc-600';
  }
}
