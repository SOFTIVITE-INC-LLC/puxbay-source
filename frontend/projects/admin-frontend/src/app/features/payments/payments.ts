import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { PaymentService, PaymentLog, PaymentStats, PaymentFilterParams } from '../../services/payment.service';
import { TenantService, Tenant } from '../../services/tenant.service';
import { AlertService } from '../../services/alert.service';

@Component({
  selector: 'app-payments',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './payments.html',
  styleUrls: ['./payments.css']
})
export class PaymentsComponent implements OnInit {
  private paymentService = inject(PaymentService);
  private tenantService = inject(TenantService);
  private alertService = inject(AlertService);

  payments = signal<PaymentLog[]>([]);
  stats = signal<PaymentStats | null>(null);
  total = signal<number>(0);
  isLoading = signal(true);
  isSaving = signal(false);
  isUpdatingDispute = signal(false);

  // Filter signals
  searchQuery = signal<string>('');
  statusFilter = signal<string>('all');
  typeFilter = signal<string>('all');
  routingFilter = signal<string>('all');
  disputeFilter = signal<string>('all');
  currentPage = signal<number>(1);
  pageSize = signal<number>(25);

  // Available tenants for manual logging
  tenants = signal<Tenant[]>([]);

  // Modals & Drawers
  isLogModalOpen = signal(false);
  isDetailDrawerOpen = signal(false);
  selectedPayment = signal<PaymentLog | null>(null);

  // Active Dropdown
  openDropdownId = signal<string | null>(null);
  copiedRef = signal<string | null>(null);

  // Manual payment form model
  newPayment = signal<Partial<PaymentLog>>({
    tenant_id: '',
    tenant_name: '',
    payment_type: 'store_order',
    reference: '',
    order_number: '',
    amount: 0,
    currency: 'GHS',
    payment_method: 'paystack',
    gateway: 'paystack',
    subaccount_code: '',
    is_subaccount_routed: false,
    subaccount_share: 0,
    platform_fee: 0,
    customer_name: '',
    customer_email: '',
    customer_phone: '',
    status: 'successful',
    dispute_status: 'none',
    notes: '',
  });

  // Edit dispute model
  editDisputeStatus = signal<string>('none');
  editDisputeNotes = signal<string>('');
  editPaymentStatus = signal<string>('successful');

  ngOnInit() {
    this.loadPayments();
    this.loadTenants();

    document.addEventListener('click', (e) => {
      if (!(e.target as Element).closest('.action-menu-container')) {
        this.openDropdownId.set(null);
      }
    });
  }

  loadTenants() {
    this.tenantService.getTenants().subscribe({
      next: (res) => this.tenants.set(res.data || []),
      error: () => {}
    });
  }

  loadPayments() {
    this.isLoading.set(true);
    const params: PaymentFilterParams = {
      search: this.searchQuery().trim(),
      status: this.statusFilter(),
      payment_type: this.typeFilter(),
      subaccount_routed: this.routingFilter(),
      dispute_status: this.disputeFilter(),
      page: this.currentPage(),
      limit: this.pageSize(),
    };

    this.paymentService.getPayments(params).subscribe({
      next: (res) => {
        this.payments.set(res.data || []);
        this.stats.set(res.stats || null);
        this.total.set(res.total || (res.data ? res.data.length : 0));
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load payments', err);
        this.alertService.error('Failed to load payment records');
        this.isLoading.set(false);
      }
    });
  }

  onSearch(query: string) {
    this.searchQuery.set(query);
    this.currentPage.set(1);
    this.loadPayments();
  }

  setStatusFilter(status: string) {
    this.statusFilter.set(status);
    this.currentPage.set(1);
    this.loadPayments();
  }

  setTypeFilter(type: string) {
    this.typeFilter.set(type);
    this.currentPage.set(1);
    this.loadPayments();
  }

  setRoutingFilter(routing: string) {
    this.routingFilter.set(routing);
    this.currentPage.set(1);
    this.loadPayments();
  }

  setDisputeFilter(dispute: string) {
    this.disputeFilter.set(dispute);
    this.currentPage.set(1);
    this.loadPayments();
  }

  toggleDropdown(id: string, event: Event) {
    event.stopPropagation();
    if (this.openDropdownId() === id) {
      this.openDropdownId.set(null);
    } else {
      this.openDropdownId.set(id);
    }
  }

  copyReference(ref: string, event?: Event) {
    if (event) event.stopPropagation();
    if (!ref) return;
    navigator.clipboard.writeText(ref);
    this.copiedRef.set(ref);
    setTimeout(() => this.copiedRef.set(null), 2500);
  }

  // Drawer details
  openDetail(payment: PaymentLog) {
    this.selectedPayment.set(payment);
    this.editDisputeStatus.set(payment.dispute_status || 'none');
    this.editDisputeNotes.set(payment.notes || '');
    this.editPaymentStatus.set(payment.status || 'successful');
    this.isDetailDrawerOpen.set(true);
    this.openDropdownId.set(null);
  }

  closeDetail() {
    this.isDetailDrawerOpen.set(false);
    this.selectedPayment.set(null);
  }

  saveDisputeUpdates() {
    const payment = this.selectedPayment();
    if (!payment) return;

    this.isUpdatingDispute.set(true);
    this.paymentService.updatePayment(payment.id, {
      status: this.editPaymentStatus(),
      dispute_status: this.editDisputeStatus(),
      notes: this.editDisputeNotes()
    }).subscribe({
      next: (res) => {
        this.isUpdatingDispute.set(false);
        this.selectedPayment.set(res.data);
        this.alertService.success('Payment details & dispute status updated');
        this.loadPayments();
      },
      error: (err) => {
        this.isUpdatingDispute.set(false);
        this.alertService.error(err.error?.error || 'Failed to update payment');
      }
    });
  }

  // Modal Log Payment
  openLogModal() {
    const autoRef = `MAN-${Math.floor(100000 + Math.random() * 900000)}`;
    this.newPayment.set({
      tenant_id: '',
      tenant_name: '',
      payment_type: 'store_order',
      reference: autoRef,
      order_number: '',
      amount: 0,
      currency: 'GHS',
      payment_method: 'paystack',
      gateway: 'paystack',
      subaccount_code: '',
      is_subaccount_routed: false,
      subaccount_share: 0,
      platform_fee: 0,
      customer_name: '',
      customer_email: '',
      customer_phone: '',
      status: 'successful',
      dispute_status: 'none',
      notes: '',
    });
    this.isLogModalOpen.set(true);
  }

  closeLogModal() {
    this.isLogModalOpen.set(false);
  }

  onTenantSelected(tenantId: string) {
    const t = this.tenants().find(item => item.id === tenantId);
    if (t) {
      this.newPayment.update(p => ({
        ...p,
        tenant_id: t.id,
        tenant_name: t.name
      }));
    }
  }

  submitNewPayment() {
    const payload = this.newPayment();
    if (!payload.amount || payload.amount <= 0) {
      this.alertService.error('Please enter a valid amount');
      return;
    }

    if (payload.subaccount_code && payload.subaccount_code.trim()) {
      payload.is_subaccount_routed = true;
    }

    this.isSaving.set(true);
    this.paymentService.createPayment(payload).subscribe({
      next: () => {
        this.isSaving.set(false);
        this.isLogModalOpen.set(false);
        this.alertService.success('Payment logged successfully');
        this.loadPayments();
      },
      error: (err) => {
        this.isSaving.set(false);
        this.alertService.error(err.error?.error || 'Failed to record payment');
      }
    });
  }

  deletePayment(id: string) {
    if (!confirm('Are you sure you want to delete this payment log record?')) return;
    this.paymentService.deletePayment(id).subscribe({
      next: () => {
        this.alertService.success('Payment log deleted');
        this.closeDetail();
        this.loadPayments();
      },
      error: (err) => this.alertService.error(err.error?.error || 'Failed to delete payment log')
    });
  }

  exportCSV() {
    const list = this.payments();
    if (list.length === 0) {
      this.alertService.error('No payment records to export');
      return;
    }

    const headers = ['Date', 'Reference', 'Tenant', 'Type', 'Amount', 'Currency', 'Payment Method', 'Subaccount Routed', 'Subaccount Code', 'Customer Name', 'Customer Phone', 'Status', 'Dispute Status', 'Notes'];
    const rows = list.map(p => [
      `"${new Date(p.created_at).toLocaleString()}"`,
      `"${p.reference}"`,
      `"${p.tenant_name || p.tenant?.name || 'Platform'}"`,
      `"${p.payment_type}"`,
      p.amount,
      `"${p.currency}"`,
      `"${p.payment_method}"`,
      p.is_subaccount_routed ? 'YES' : 'NO',
      `"${p.subaccount_code || '-'}"`,
      `"${p.customer_name || '-'}"`,
      `"${p.customer_phone || '-'}"`,
      `"${p.status}"`,
      `"${p.dispute_status || 'none'}"`,
      `"${(p.notes || '').replace(/"/g, '""')}"`
    ]);

    const csvContent = 'data:text/csv;charset=utf-8,' + [headers.join(','), ...rows.map(e => e.join(','))].join('\n');
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', `payments_report_${new Date().toISOString().slice(0,10)}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    this.alertService.success('CSV Export downloaded successfully');
  }
}
