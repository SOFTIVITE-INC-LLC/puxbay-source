import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { Customer } from '../../../core/models/models';
import { CustomerService } from '../../../core/services/customer.service';
import { CrmService } from '../../../core/services/crm.service';
import { ToastService } from '../../../core/services/toast';

@Component({
  selector: 'app-customers',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe],
  templateUrl: './customers.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.6);
      backdrop-filter: blur(16px);
      border: 1px solid rgba(255, 255, 255, 0.3);
    }
    .dark .glass-panel {
      background: rgba(24, 24, 27, 0.6);
      border: 1px solid rgba(255, 255, 255, 0.05);
    }
  `,
})
export class Customers implements OnInit {
  crmService = inject(CustomerService);
  private crm = inject(CrmService);
  private toastService = inject(ToastService);

  viewMode = signal<'grid' | 'table'>('grid');
  isModalOpen = signal(false);
  modalTitle = signal('Add Customer');

  // Payment modal
  showPaymentModal = signal(false);
  paymentCustomer = signal<Customer | null>(null);
  paymentAmount = signal(0);
  paymentMethod = signal('cash');
  paymentNotes = signal('');
  payingDown = signal(false);
  
  currentCustomer = signal<Partial<Customer>>({
    loyalty_pts: 0,
    store_credit: 0,
    debt_balance: 0,
    total_spend: 0,
    order_count: 0
  });

  searchQuery = signal('');

  // Overview stats
  totalCustomers = computed(() => this.crmService.customers().length);
  totalSpend = computed(() => this.crmService.customers().reduce((sum: number, c: any) => sum + (Number(c.total_spend) || 0), 0));
  totalCredit = computed(() => this.crmService.customers().reduce((sum: number, c: any) => sum + (Number(c.store_credit) || 0), 0));
  totalDebt = computed(() => this.crmService.customers().reduce((sum: number, c: any) => sum + (Number(c.debt_balance) || 0), 0));

  ngOnInit() {
    this.crmService.getCustomers().subscribe();
  }

  get filteredCustomers() {
    const q = this.searchQuery().toLowerCase();
    return this.crmService.customers().filter((c: any) => 
      (c.name || '').toLowerCase().includes(q) || 
      (c.phone || '').toLowerCase().includes(q) ||
      (c.email || '').toLowerCase().includes(q)
    );
  }

  openAddModal() {
    this.modalTitle.set('Add Customer');
    this.currentCustomer.set({
      loyalty_pts: 0,
      store_credit: 0,
      debt_balance: 0,
      total_spend: 0,
      order_count: 0
    });
    this.isModalOpen.set(true);
  }

  openEditModal(customer: Customer) {
    this.modalTitle.set('Edit Customer');
    this.currentCustomer.set({ ...customer });
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  saveCustomer() {
    const c = this.currentCustomer();
    if (c.id) {
      this.crmService.updateCustomer(c.id, c).subscribe(() => {
        this.closeModal();
        this.toastService.showSuccess('Customer updated!');
        this.crmService.getCustomers().subscribe();
      });
    } else {
      this.crmService.createCustomer(c).subscribe(() => {
        this.closeModal();
        this.toastService.showSuccess('Customer added!');
        this.crmService.getCustomers().subscribe();
      });
    }
  }

  openPaymentModal(customer: Customer, event: Event) {
    event.stopPropagation();
    this.paymentCustomer.set(customer);
    this.paymentAmount.set(customer.debt_balance || 0);
    this.paymentMethod.set('cash');
    this.paymentNotes.set('');
    this.showPaymentModal.set(true);
  }

  submitPayment() {
    const c = this.paymentCustomer();
    if (!c) return;
    if (this.paymentAmount() <= 0) { this.toastService.showError('Amount must be positive'); return; }
    this.payingDown.set(true);
    this.crm.recordCustomerPayment(c.id, this.paymentAmount(), this.paymentMethod(), this.paymentNotes()).subscribe({
      next: () => {
        this.toastService.showSuccess('Payment recorded successfully!');
        this.showPaymentModal.set(false);
        this.payingDown.set(false);
        this.crmService.getCustomers().subscribe();
      },
      error: (err: any) => {
        this.toastService.showError(err?.error?.error || 'Failed to record payment.');
        this.payingDown.set(false);
      }
    });
  }

  getLoyaltyTier(pts: number): { label: string; color: string; icon: string } {
    if (pts >= 10000) return { label: 'Platinum', color: 'text-slate-400', icon: 'diamond' };
    if (pts >= 5000) return { label: 'Gold', color: 'text-amber-500', icon: 'military_tech' };
    if (pts >= 1000) return { label: 'Silver', color: 'text-slate-400', icon: 'workspace_premium' };
    return { label: 'Bronze', color: 'text-orange-400', icon: 'star' };
  }

  exportCSV() {
    const headers = ['Name', 'Email', 'Phone', 'Total Spend', 'Loyalty Points', 'Store Credit', 'Debt Balance', 'Orders'];
    const rows = this.filteredCustomers.map((c: any) => [
      c.name, c.email || '', c.phone || '',
      c.total_spend || 0, c.loyalty_pts || 0,
      c.store_credit || 0, c.debt_balance || 0, c.order_count || 0
    ]);
    const csv = [headers, ...rows].map(r => r.map((v: any) => `"${v}"`).join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = 'customers.csv';
    a.click();
    this.toastService.showSuccess('Customers exported successfully!');
  }

  importCSV(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (e) => {
      const text = e.target?.result as string;
      const lines = text.split('\n').slice(1).filter(l => l.trim());
      let count = 0;
      const requests = lines.map(line => {
        const [name, email, phone] = line.split(',').map(v => v.replace(/"/g, '').trim());
        if (!name) return null;
        count++;
        return this.crmService.createCustomer({ name, email, phone });
      }).filter(Boolean);

      Promise.all(requests.map(r => r!.toPromise())).then(() => {
        this.toastService.showSuccess(`Imported ${count} customers!`);
        this.crmService.getCustomers().subscribe();
      }).catch(() => this.toastService.showError('Some customers failed to import.'));
    };
    reader.readAsText(file);
  }
}
