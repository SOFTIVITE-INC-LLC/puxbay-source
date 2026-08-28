import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { SupplierPortalService, PurchaseOrder, SupplierASN, SupplierInvoice, SupplierMessage } from '../../services/supplier-portal.service';
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

  // Tab inside Details modal: 'items' | 'chat' | 'qr'
  detailsTab = signal<'items' | 'chat' | 'qr'>('items');
  messages = signal<SupplierMessage[]>([]);
  newMsgText = signal<string>('');
  sendingMsg = signal<boolean>(false);

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
    if (s === 'received' || s === 'confirmed') return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30';
    if (s === 'partially_received' || s === 'issued') return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30';
    if (s === 'cancelled' || s === 'rejected') return 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30';
    return 'bg-slate-100 dark:bg-zinc-800 text-slate-700 dark:text-zinc-300 border-slate-200 dark:border-zinc-700';
  }

  viewOrder(order: PurchaseOrder) {
    this.selectedOrder.set(order);
    this.detailsTab.set('items');
    this.loadMessages(order.po_number);
  }

  loadMessages(poNumber: string) {
    this.portalService.getMessages(poNumber).subscribe({
      next: (msgs) => this.messages.set(msgs || []),
      error: () => {}
    });
  }

  sendMessage() {
    const text = this.newMsgText().trim();
    const po = this.selectedOrder();
    if (!text || !po) return;

    this.sendingMsg.set(true);
    this.portalService.sendMessage({
      reference_id: po.po_number,
      sender_name: 'Supplier Representative',
      sender_type: 'supplier',
      message: text
    }).subscribe({
      next: (msg) => {
        this.messages.update(list => [...list, msg]);
        this.newMsgText.set('');
        this.sendingMsg.set(false);
      },
      error: () => this.sendingMsg.set(false)
    });
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
    this.asnTracking = `TRK-${Math.floor(100000 + Math.random() * 900000)}`;
    this.asnExpectedArrival = order.expected_date ? order.expected_date.split('T')[0] : '';
    this.asnPackages = 1;
    this.asnWeight = 5;
    this.asnNotes = '';
    this.showAsnModal.set(true);
  }

  submitAsn() {
    const o = this.selectedOrder();
    if (!o) return;

    const payload: Partial<SupplierASN> = {
      purchase_order_id: o.id,
      carrier: this.asnCarrier,
      tracking_number: this.asnTracking,
      dispatch_date: new Date().toISOString(),
      expected_arrival: this.asnExpectedArrival ? new Date(this.asnExpectedArrival).toISOString() : undefined,
      package_count: this.asnPackages,
      total_weight_kg: this.asnWeight,
      status: 'dispatched',
      notes: this.asnNotes
    };

    this.portalService.createShipment(payload).subscribe({
      next: () => {
        this.toast.showSuccess(`ASN for PO #${o.po_number} dispatched successfully!`);
        this.showAsnModal.set(false);
        this.loadOrders();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to create ASN')
    });
  }

  // --- Flip to Invoice ---
  openInvoiceModal(order: PurchaseOrder) {
    this.selectedOrder.set(order);
    this.invoiceNumber = `INV-${order.po_number.replace('PO-', '')}`;
    const nextMonth = new Date();
    nextMonth.setDate(nextMonth.getDate() + 30);
    this.invoiceDueDate = nextMonth.toISOString().split('T')[0];
    this.showInvoiceModal.set(true);
  }

  submitInvoice() {
    const o = this.selectedOrder();
    if (!o) return;

    this.portalService.flipPOToInvoice(o.id, {
      invoice_number: this.invoiceNumber,
      due_date: this.invoiceDueDate ? new Date(this.invoiceDueDate).toISOString() : undefined
    }).subscribe({
      next: (inv) => {
        this.toast.showSuccess(`Invoice #${inv.invoice_number} created successfully!`);
        this.showInvoiceModal.set(false);
        this.router.navigate(['/supplier-portal/invoices']);
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to create invoice')
    });
  }

  printSlip() {
    this.showPrintSlip.set(true);
    setTimeout(() => {
      window.print();
    }, 200);
  }

  exportCSV() {
    const rows = this.filteredOrders.map(o => ({
      PO_Number: o.po_number,
      Created_At: o.created_at,
      Status: o.status,
      Total_Amount: o.total_amount,
      Expected_Date: o.expected_date || 'N/A'
    }));

    if (!rows.length) {
      this.toast.showError('No purchase orders to export');
      return;
    }

    const header = Object.keys(rows[0]).join(',');
    const csv = [header, ...rows.map(r => Object.values(r).join(','))].join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `purchase_orders_${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    window.URL.revokeObjectURL(url);
    this.toast.showSuccess('Orders exported to CSV');
  }
}
