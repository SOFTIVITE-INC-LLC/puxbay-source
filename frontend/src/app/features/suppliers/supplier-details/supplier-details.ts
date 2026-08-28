import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { SupplierService } from '../../../core/services/supplier.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { ToastService } from '../../../core/services/toast';
import { AlertService } from '../../../core/services/alert.service';

@Component({
  selector: 'app-supplier-details',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule, AppCurrencyPipe, DatePipe],
  templateUrl: './supplier-details.html'
})
export class SupplierDetailsComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private supplierService = inject(SupplierService);
  private toast = inject(ToastService);
  private alertService = inject(AlertService);

  supplierId = '';
  details = signal<any | null>(null);
  loading = signal<boolean>(true);

  activeTab = signal<'overview' | 'ledger' | 'catalog' | 'orders' | 'invoices' | 'rmas' | 'documents'>('overview');

  // Ledger state
  newPayment = signal<{ entry_type: string; amount: number; reference_id: string; notes: string; transaction_date: string }>({
    entry_type: 'payment', amount: 0, reference_id: '', notes: '', transaction_date: ''
  });
  addingPayment = signal(false);

  // Edit Credit Limit modal state
  showCreditModal = signal(false);
  editCreditLimit = 0;
  savingCreditLimit = signal(false);

  // Print Statement Modal state
  showStatementModal = signal(false);

  // Today's date for Statement of Account header
  readonly today = new Date();

  ngOnInit() {
    this.supplierId = this.route.snapshot.paramMap.get('id') || '';
    if (this.supplierId) {
      this.loadDetails();
    }
  }

  loadDetails() {
    this.loading.set(true);
    this.supplierService.getSupplierDetails(this.supplierId).subscribe({
      next: (res) => {
        this.details.set(res);
        if (res?.supplier) {
          this.editCreditLimit = res.supplier.credit_limit || 0;
        }
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
        this.toast.showError('Failed to load supplier details');
      }
    });
  }

  addLedgerPayment() {
    const p = this.newPayment();
    if (p.amount <= 0) {
      this.toast.showError('Please enter a valid amount');
      return;
    }
    this.addingPayment.set(true);
    this.supplierService.addSupplierLedger(this.supplierId, {
      entry_type: p.entry_type,
      amount: Number(p.amount),
      reference_id: p.reference_id || undefined,
      notes: p.notes || undefined,
      transaction_date: p.transaction_date || undefined
    }).subscribe({
      next: () => {
        this.addingPayment.set(false);
        this.newPayment.set({ entry_type: 'payment', amount: 0, reference_id: '', notes: '', transaction_date: '' });
        this.toast.showSuccess('Transaction recorded successfully!');
        this.loadDetails();
      },
      error: () => {
        this.addingPayment.set(false);
        this.toast.showError('Failed to record transaction');
      }
    });
  }

  saveCreditLimit() {
    this.savingCreditLimit.set(true);
    this.supplierService.updateSupplier(this.supplierId, {
      ...this.details()?.supplier,
      credit_limit: Number(this.editCreditLimit)
    } as any).subscribe({
      next: () => {
        this.savingCreditLimit.set(false);
        this.showCreditModal.set(false);
        this.toast.showSuccess('Credit limit updated successfully!');
        this.loadDetails();
      },
      error: () => {
        this.savingCreditLimit.set(false);
        this.toast.showError('Failed to update credit limit');
      }
    });
  }

  disburseInvoice(invoiceId: string) {
    this.supplierService.disburseInvoicePayout(invoiceId).subscribe({
      next: () => {
        this.toast.showSuccess('Invoice payout disbursed successfully!');
        this.loadDetails();
      },
      error: () => this.toast.showError('Failed to disburse payout')
    });
  }

  printStatement() {
    window.print();
  }

  exportLedgerCSV() {
    const entries = this.details()?.ledger_entries || [];
    const name = this.details()?.supplier?.name || 'supplier';
    const rows = [
      ['Date', 'Type', 'Reference', 'Amount', 'Balance', 'Notes'],
      ...entries.map((e: any) => [
        e.transaction_date ? new Date(e.transaction_date).toLocaleDateString() : new Date(e.created_at).toLocaleDateString(),
        e.entry_type,
        e.reference_id || '',
        e.amount,
        e.balance,
        e.notes || ''
      ])
    ];
    const csv = rows.map(r => r.join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${name.replace(/\s+/g, '_')}_statement.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }
}
