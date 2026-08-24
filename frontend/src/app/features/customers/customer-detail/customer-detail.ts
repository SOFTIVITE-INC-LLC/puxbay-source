import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule, ActivatedRoute, Router } from '@angular/router';
import { CrmService } from '../../../core/services/crm.service';
import { OrderService } from '../../../core/services/order.service';
import { ToastService } from '../../../core/services/toast';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { Customer } from '../../../core/models/models';
import { Order } from '../../../core/models/order.models';

@Component({
  selector: 'app-customer-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe, DatePipe],
  templateUrl: './customer-detail.html'
})
export class CustomerDetail implements OnInit {
  protected readonly Math = Math;
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private crmService = inject(CrmService);
  private orderService = inject(OrderService);
  private toastService = inject(ToastService);

  customer = signal<Customer | null>(null);
  orders = signal<Order[]>([]);
  loadingCustomer = signal(true);
  loadingOrders = signal(true);
  activeTab = signal<'overview' | 'orders'>('overview');

  // Payment modal
  showPaymentModal = signal(false);
  paymentAmount = signal(0);
  paymentMethod = signal('cash');
  paymentNotes = signal('');
  payingDown = signal(false);

  // Edit modal
  showEditModal = signal(false);
  editName = signal('');
  editEmail = signal('');
  editPhone = signal('');
  editAddress = signal('');
  saving = signal(false);

  loyaltyTier = computed(() => {
    const pts = this.customer()?.loyalty_pts || 0;
    if (pts >= 10000) return { label: 'Platinum', color: 'text-slate-300', bg: 'bg-slate-600/30 dark:bg-slate-400/20', icon: 'diamond' };
    if (pts >= 5000) return { label: 'Gold', color: 'text-amber-500', bg: 'bg-amber-50 dark:bg-amber-500/20', icon: 'military_tech' };
    if (pts >= 1000) return { label: 'Silver', color: 'text-slate-400', bg: 'bg-slate-100 dark:bg-slate-500/20', icon: 'workspace_premium' };
    return { label: 'Bronze', color: 'text-orange-400', bg: 'bg-orange-50 dark:bg-orange-500/20', icon: 'star' };
  });

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id')!;
    this.loadCustomer(id);
    this.loadOrders(id);
  }

  loadCustomer(id: string) {
    this.loadingCustomer.set(true);
    this.crmService['api'].get<Customer>(`/customers/${id}`).subscribe({
      next: (c) => { this.customer.set(c); this.loadingCustomer.set(false); },
      error: () => { this.loadingCustomer.set(false); this.router.navigate(['/customers']); }
    });
  }

  loadOrders(id: string) {
    this.loadingOrders.set(true);
    this.orderService.getOrders({ customer_id: id, limit: 50 }).subscribe({
      next: (res) => { this.orders.set(res.data || []); this.loadingOrders.set(false); },
      error: () => this.loadingOrders.set(false)
    });
  }

  openPaymentModal() {
    this.paymentAmount.set(this.customer()?.debt_balance || 0);
    this.paymentMethod.set('cash');
    this.paymentNotes.set('');
    this.showPaymentModal.set(true);
  }

  submitPayment() {
    const c = this.customer();
    if (!c) return;
    if (this.paymentAmount() <= 0) { this.toastService.showError('Amount must be greater than 0'); return; }
    this.payingDown.set(true);
    this.crmService.recordCustomerPayment(c.id, this.paymentAmount(), this.paymentMethod(), this.paymentNotes()).subscribe({
      next: () => {
        this.toastService.showSuccess('Payment recorded successfully!');
        this.showPaymentModal.set(false);
        this.loadCustomer(c.id);
        this.payingDown.set(false);
      },
      error: (err) => {
        this.toastService.showError(err?.error?.error || 'Failed to record payment.');
        this.payingDown.set(false);
      }
    });
  }

  openEditModal() {
    const c = this.customer();
    if (!c) return;
    this.editName.set(c.name);
    this.editEmail.set(c.email || '');
    this.editPhone.set((c as any).phone || '');
    this.editAddress.set((c as any).address || '');
    this.showEditModal.set(true);
  }

  saveEdit() {
    const c = this.customer();
    if (!c || !this.editName()) return;
    this.saving.set(true);
    this.crmService.updateCustomer(c.id, {
      name: this.editName(),
      email: this.editEmail(),
      phone: this.editPhone(),
      address: this.editAddress()
    }).subscribe({
      next: () => {
        this.toastService.showSuccess('Customer updated!');
        this.showEditModal.set(false);
        this.loadCustomer(c.id);
        this.saving.set(false);
      },
      error: () => { this.toastService.showError('Failed to update customer.'); this.saving.set(false); }
    });
  }

  getOrderStatusClass(status: string): string {
    const map: Record<string, string> = {
      completed: 'bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400',
      pending: 'bg-amber-50 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400',
      cancelled: 'bg-rose-50 dark:bg-rose-900/30 text-rose-700 dark:text-rose-400',
      refunded: 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400',
    };
    return map[status] || 'bg-zinc-100 dark:bg-zinc-800 text-zinc-500';
  }

  getInitial(name: string): string {
    return (name || 'C').charAt(0).toUpperCase();
  }
}
