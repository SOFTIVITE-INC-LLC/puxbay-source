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
  creditData = signal<{
    account: any;
    available_credit: number;
    transactions: any[];
    instalments: any[];
  } | null>(null);

  loadingCustomer = signal(true);
  loadingOrders = signal(true);
  loadingCredit = signal(true);
  activeTab = signal<'overview' | 'orders' | 'credit'>('overview');

  // Payment modal
  showPaymentModal = signal(false);
  paymentAmount = signal(0);
  paymentMethod = signal('cash');
  paymentReference = signal('');
  paymentNotes = signal('');
  payingDown = signal(false);

  // Credit limit modal
  showCreditLimitModal = signal(false);
  creditLimitValue = signal(0);
  creditDaysToRepay = signal(30);
  creditLimitNotes = signal('');
  savingCreditLimit = signal(false);
  sendingReminder = signal(false);

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

  // Outstanding debt computed from customer or credit account
  outstandingDebt = computed(() => {
    const accBal = this.creditData()?.account?.balance;
    if (accBal !== undefined && accBal !== null) return Number(accBal);
    return Number(this.customer()?.debt_balance || 0);
  });

  // Credit limit computed
  creditLimit = computed(() => {
    const lim = this.creditData()?.account?.credit_limit;
    if (lim !== undefined && lim !== null) return Number(lim);
    return Number((this.customer() as any)?.credit_limit || 0);
  });

  // Available credit computed
  availableCredit = computed(() => {
    const avail = this.creditData()?.available_credit;
    if (avail !== undefined && avail !== null) return Number(avail);
    return Math.max(0, this.creditLimit() - this.outstandingDebt());
  });

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id')!;
    this.loadCustomer(id);
    this.loadOrders(id);
    this.loadCreditAccount(id);
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

  loadCreditAccount(id: string) {
    this.loadingCredit.set(true);
    this.crmService.getCreditAccount(id).subscribe({
      next: (data) => {
        this.creditData.set(data);
        this.loadingCredit.set(false);
      },
      error: () => {
        this.loadingCredit.set(false);
      }
    });
  }

  openPaymentModal() {
    const currentDebt = this.outstandingDebt();
    this.paymentAmount.set(currentDebt > 0 ? currentDebt : 0);
    this.paymentMethod.set('cash');
    this.paymentReference.set('');
    this.paymentNotes.set('');
    this.showPaymentModal.set(true);
  }

  submitPayment() {
    const c = this.customer();
    if (!c) return;
    if (this.paymentAmount() <= 0) {
      this.toastService.showError('Repayment amount must be greater than 0');
      return;
    }
    this.payingDown.set(true);
    this.crmService.recordCustomerPayment(
      c.id,
      this.paymentAmount(),
      this.paymentMethod(),
      this.paymentNotes(),
      this.paymentReference()
    ).subscribe({
      next: (res) => {
        this.toastService.showSuccess(res?.message || 'Debt repayment recorded successfully! SMS receipt sent.');
        this.showPaymentModal.set(false);
        this.loadCustomer(c.id);
        this.loadCreditAccount(c.id);
        this.payingDown.set(false);
      },
      error: (err) => {
        this.toastService.showError(err?.error?.error || 'Failed to record payment.');
        this.payingDown.set(false);
      }
    });
  }

  openCreditLimitModal() {
    this.creditLimitValue.set(this.creditLimit());
    this.creditDaysToRepay.set(this.creditData()?.account?.days_to_repay || 30);
    this.creditLimitNotes.set(this.creditData()?.account?.notes || '');
    this.showCreditLimitModal.set(true);
  }

  saveCreditLimit() {
    const c = this.customer();
    if (!c) return;
    if (this.creditLimitValue() < 0) {
      this.toastService.showError('Credit limit cannot be negative');
      return;
    }
    this.savingCreditLimit.set(true);
    this.crmService.setCreditLimit(
      c.id,
      this.creditLimitValue(),
      this.creditDaysToRepay(),
      this.creditLimitNotes()
    ).subscribe({
      next: () => {
        this.toastService.showSuccess('Customer credit limit updated successfully!');
        this.showCreditLimitModal.set(false);
        this.loadCreditAccount(c.id);
        this.loadCustomer(c.id);
        this.savingCreditLimit.set(false);
      },
      error: (err) => {
        this.toastService.showError(err?.error?.error || 'Failed to update credit limit.');
        this.savingCreditLimit.set(false);
      }
    });
  }

  sendReminder() {
    const c = this.customer();
    if (!c) return;
    this.sendingReminder.set(true);
    this.crmService.sendCreditReminder(c.id).subscribe({
      next: () => {
        this.toastService.showSuccess(`Repayment reminder sent to ${c.phone || c.name} via SMS/WhatsApp!`);
        this.sendingReminder.set(false);
      },
      error: (err) => {
        this.toastService.showError(err?.error?.error || 'Failed to send reminder.');
        this.sendingReminder.set(false);
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
        this.toastService.showSuccess('Customer profile updated!');
        this.showEditModal.set(false);
        this.loadCustomer(c.id);
        this.saving.set(false);
      },
      error: () => {
        this.toastService.showError('Failed to update customer.');
        this.saving.set(false);
      }
    });
  }

  getOrderStatusClass(status?: string): string {
    const map: Record<string, string> = {
      completed: 'bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400',
      pending: 'bg-amber-50 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400',
      cancelled: 'bg-rose-50 dark:bg-rose-900/30 text-rose-700 dark:text-rose-400',
      refunded: 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400',
    };
    return (status && map[status]) || 'bg-zinc-100 dark:bg-zinc-800 text-zinc-500';
  }

  getInitial(name: string): string {
    return (name || 'C').charAt(0).toUpperCase();
  }
}
