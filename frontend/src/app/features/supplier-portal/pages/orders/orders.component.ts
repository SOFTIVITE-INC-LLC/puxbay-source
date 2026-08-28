import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { SupplierPortalService, PurchaseOrder, SupplierASN, SupplierInvoice } from '../../services/supplier-portal.service';
import { AppCurrencyPipe } from '../../../../core/pipes/app-currency.pipe';
import { ToastService } from '../../../../core/services/toast';

@Component({
  selector: 'app-supplier-portal-orders',
  standalone: true,
  imports: [CommonModule, DatePipe, AppCurrencyPipe, FormsModule, RouterModule],
  templateUrl: './orders.component.html'
})
export class SupplierPortalOrdersComponent implements OnInit {
  portalService = inject(SupplierPortalService);
  private router = inject(Router);
  private toast = inject(ToastService);

  orders = signal<PurchaseOrder[]>([]);
  statusFilter = signal<string>('all');
  searchQuery = signal<string>('');
  selectedOrder = signal<PurchaseOrder | null>(null);

  // Modals state
  showAckModal = signal<boolean>(false);
  ackStatus = 'confirmed';
  ackExpectedDate = '';
  ackNotes = '';

  showAsnModal = signal<boolean>(false);
  asnCarrier = 'DHL Express';
  asnTracking = '';
  asnExpectedArrival = '';
  asnPackages = 1;
  asnWeight = 5;
  asnNotes = '';

  showInvoiceModal = signal<boolean>(false);
  invoiceNumber = '';
  invoiceDueDate = '';

  showPrintSlip = signal<boolean>(false);
  loading = signal<boolean>(false);

  ngOnInit() {
    this.loadOrders();
  }

  loadOrders() {
    this.loading.set(true);
    this.portalService.getPurchaseOrders().subscribe({
      next: (res) => {
        this.orders.set(res || []);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
        this.router.navigate(['/supplier-portal/login']);
      }
    });
  }

  get filteredOrders(): PurchaseOrder[] {
    const list = this.orders();
    const filter = this.statusFilter();
    const q = this.searchQuery().toLowerCase().trim();

    return list.filter(o => {
      const matchStatus = filter === 'all' || o.status.toLowerCase() === filter.toLowerCase();
      const matchQuery = !q || o.po_number.toLowerCase().includes(q) || (o.notes && o.notes.toLowerCase().includes(q));
      return matchStatus && matchQuery;
    });
  }

  statusClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s === 'received' || s === 'confirmed') return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
    if (s === 'partially_received' || s === 'issued') return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
    if (s === 'cancelled' || s === 'rejected') return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
    return 'bg-zinc-800 text-zinc-300 border-zinc-700';
  }

  viewOrder(order: PurchaseOrder) {
    this.selectedOrder.set(order);
  }

  closeOrder() {
    this.selectedOrder.set(null);
    this.showPrintSlip.set(false);
  }

  // --- Acknowledgment ---
  openAckModal(order: PurchaseOrder) {
    this.selectedOrder.set(order);
    this.ackStatus = 'confirmed';
    this.ackExpectedDate = order.expected_date ? order.expected_date.split('T')[0] : '';
    this.ackNotes = '';
    this.showAckModal.set(true);
  }

  submitAcknowledgment() {
    const o = this.selectedOrder();
    if (!o) return;

    this.portalService.acknowledgePO(o.id, {
      status: this.ackStatus,
      expected_date: this.ackExpectedDate,
      notes: this.ackNotes
    }).subscribe({
      next: (updated) => {
        this.toast.showSuccess(`PO #${o.po_number} ${this.ackStatus === 'confirmed' ? 'Accepted' : 'Updated'}`);
        this.showAckModal.set(false);
        this.loadOrders();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to update order acknowledgment')
    });
  }

  // --- ASN Dispatch ---
  openAsnModal(order: PurchaseOrder) {
    this.selectedOrder.set(order);
    this.asnCarrier = 'DHL Express';
    this.asnTracking = '';
    this.asnExpectedArrival = order.expected_date ? order.expected_date.split('T')[0] : '';
    this.asnPackages = 1;
    this.asnWeight = 5;
    this.asnNotes = '';
    this.showAsnModal.set(true);
  }

  submitASN() {
    const o = this.selectedOrder();
    if (!o) return;

    const payload: Partial<SupplierASN> = {
      purchase_order_id: o.id,
      carrier: this.asnCarrier,
      tracking_number: this.asnTracking,
      dispatch_date: new Date().toISOString(),
      expected_arrival: this.asnExpectedArrival ? new Date(this.asnExpectedArrival).toISOString() : undefined,
      package_count: Number(this.asnPackages) || 1,
      total_weight_kg: Number(this.asnWeight) || 0,
      notes: this.asnNotes,
      status: 'dispatched'
    };

    this.portalService.createShipment(payload).subscribe({
      next: () => {
        this.toast.showSuccess(`ASN for PO #${o.po_number} dispatched successfully!`);
        this.showAsnModal.set(false);
        this.router.navigate(['/supplier-portal/shipments']);
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to dispatch ASN')
    });
  }

  // --- Invoice Flip ---
  openInvoiceModal(order: PurchaseOrder) {
    this.selectedOrder.set(order);
    this.invoiceNumber = `INV-${order.po_number}`;
    const net30 = new Date();
    net30.setDate(net30.getDate() + 30);
    this.invoiceDueDate = net30.toISOString().split('T')[0];
    this.showInvoiceModal.set(true);
  }

  submitInvoiceFlip() {
    const o = this.selectedOrder();
    if (!o) return;

    this.portalService.flipPOToInvoice(o.id, {
      invoice_number: this.invoiceNumber,
      due_date: this.invoiceDueDate
    }).subscribe({
      next: () => {
        this.toast.showSuccess(`Invoice created for PO #${o.po_number}`);
        this.showInvoiceModal.set(false);
        this.router.navigate(['/supplier-portal/invoices']);
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to create invoice')
    });
  }

  // --- Printable Slip ---
  printOrderSlip() {
    this.showPrintSlip.set(true);
    setTimeout(() => {
      window.print();
    }, 200);
  }

  // --- CSV Export ---
  exportCSV() {
    const data = this.filteredOrders;
    if (!data.length) return;

    const headers = ['PO Number', 'Date', 'Status', 'Total Amount', 'Expected Date', 'Notes'];
    const rows = data.map(o => [
      o.po_number,
      o.created_at,
      o.status,
      o.total_amount,
      o.expected_date || '',
      `"${(o.notes || '').replace(/"/g, '""')}"`
    ]);

    const csvContent = 'data:text/csv;charset=utf-8,' + [headers.join(','), ...rows.map(e => e.join(','))].join('\n');
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', `Purchase_Orders_${new Date().toISOString().split('T')[0]}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }
}
