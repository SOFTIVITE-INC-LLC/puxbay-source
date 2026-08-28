import { ToastService } from '../../../core/services/toast';
import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule, DatePipe } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { SupplierService } from '../../../core/services/supplier.service';
import { CatalogService } from '../../../core/services/catalog.service';
import { Supplier } from '../../../core/models/financial.models';
import { AlertService } from '../../../core/services/alert.service';
import { SearchableSelectComponent } from '../../../shared/components/searchable-select/searchable-select';

@Component({
  selector: 'app-suppliers',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule, AppCurrencyPipe, DatePipe, SearchableSelectComponent],
  templateUrl: './suppliers.html',
})
export class Suppliers implements OnInit {
  toastService = inject(ToastService);
  supplierService = inject(SupplierService);
  alertService = inject(AlertService);
  catalogService = inject(CatalogService);

  // View state
  searchQuery = signal('');
  statusFilter = signal<'all' | 'active' | 'inactive'>('all');
  sortOrder = signal<'name' | 'name_desc' | 'balance' | 'newest'>('name');
  viewMode = signal<'grid' | 'list'>('grid');
  isModalOpen = signal(false);
  isEditing = signal(false);
  saving = signal(false);
  deletingId = signal<string | null>(null);
  currentSupplier = signal<Partial<Supplier>>({ payment_terms: 'net30' });

  // Catalog tab state
  catalogProducts = signal<any[]>([]);
  allProducts = signal<any[]>([]);
  activeTab = signal<'details' | 'catalog' | 'ledger'>('details');

  productOptions = computed(() =>
    this.allProducts().map(p => ({
      label: p.name,
      value: p.id,
      sublabel: `SKU: ${p.sku}${p.current_stock !== undefined ? ` • Stock: ${p.current_stock}` : ''}`
    }))
  );

  newCatalogProduct = signal<{ product_id: string; supplier_sku: string; unit_cost: number; min_order_qty: number }>({
    product_id: '', supplier_sku: '', unit_cost: 0, min_order_qty: 1
  });
  addingToCatalog = signal(false);

  // Ledger tab state
  ledgerEntries = signal<any[]>([]);
  newPayment = signal<{ entry_type: string; amount: number; reference_id: string; notes: string; transaction_date: string }>({
    entry_type: 'payment', amount: 0, reference_id: '', notes: '', transaction_date: ''
  });
  addingPayment = signal(false);

  // Top-level View
  mainView = signal<'suppliers' | 'proposals' | 'rmas' | 'dock_schedule'>('suppliers');

  // Price Proposals state
  priceProposals = signal<any[]>([]);
  loadingProposals = signal(false);

  // Defect Claims / RMAs state
  rmas = signal<any[]>([]);
  loadingRMAs = signal(false);
  isRMAModalOpen = signal(false);
  newRMA = signal<{ supplier_id: string; product_id: string; quantity: number; reason: string; photo_url: string; purchase_order_id: string }>({
    supplier_id: '', product_id: '', quantity: 1, reason: 'damaged_in_transit', photo_url: '', purchase_order_id: ''
  });
  savingRMA = signal(false);

  // Dock Schedule state
  dockSlots = signal<any[]>([]);
  loadingDockSlots = signal(false);

  // Merchant Broadcast Announcement state
  isAnnouncementModalOpen = signal(false);
  announcementTitle = '';
  announcementContent = '';
  announcementPriority: 'info' | 'warning' | 'urgent' = 'info';
  broadcastingAnnouncement = signal(false);

  // Disburse Payout state
  disbursingInvoiceId = signal<string | null>(null);

  // Invite modal state
  isInviteModalOpen = signal(false);
  inviteSupplier = signal<Supplier | null>(null);
  inviteEmail = '';
  invitePassword = '';
  sendingInvite = signal(false);

  ngOnInit() {
    this.supplierService.getSuppliers({ limit: -1 }).subscribe();
    this.catalogService.getProducts({ limit: -1 }).subscribe(res => this.allProducts.set((res as any).data || []));
  }

  setMainView(view: 'suppliers' | 'proposals' | 'rmas' | 'dock_schedule') {
    this.mainView.set(view);
    if (view === 'proposals') this.loadPriceProposals();
    if (view === 'rmas') this.loadRMAs();
    if (view === 'dock_schedule') this.loadDockSlots();
  }

  // ── Price Proposals Hub ──
  loadPriceProposals() {
    this.loadingProposals.set(true);
    this.supplierService.getAllPriceRequests().subscribe({
      next: (res) => {
        this.priceProposals.set(res || []);
        this.loadingProposals.set(false);
      },
      error: () => this.loadingProposals.set(false)
    });
  }

  approvePriceProposal(req: any) {
    this.supplierService.approvePriceProposal(req.id).subscribe({
      next: () => {
        this.toastService.showSuccess(`Approved price change for ${req.product?.name}`);
        this.loadPriceProposals();
      },
      error: () => this.toastService.showError('Failed to approve price change')
    });
  }

  rejectPriceProposal(req: any) {
    this.supplierService.rejectPriceProposal(req.id).subscribe({
      next: () => {
        this.toastService.showSuccess(`Rejected price change for ${req.product?.name}`);
        this.loadPriceProposals();
      },
      error: () => this.toastService.showError('Failed to reject price change')
    });
  }

  // ── Defect Claims & RMAs Hub ──
  loadRMAs() {
    this.loadingRMAs.set(true);
    this.supplierService.getRMAs().subscribe({
      next: (res) => {
        this.rmas.set(res || []);
        this.loadingRMAs.set(false);
      },
      error: () => this.loadingRMAs.set(false)
    });
  }

  openCreateRMAModal() {
    this.newRMA.set({
      supplier_id: this.supplierService.suppliers()[0]?.id || '',
      product_id: this.allProducts()[0]?.id || '',
      quantity: 1,
      reason: 'damaged_in_transit',
      photo_url: '',
      purchase_order_id: ''
    });
    this.isRMAModalOpen.set(true);
  }

  closeRMAModal() {
    this.isRMAModalOpen.set(false);
  }

  submitRMA() {
    const data = this.newRMA();
    if (!data.supplier_id || !data.product_id || data.quantity <= 0) {
      this.toastService.showError('Please select a supplier, product, and valid quantity');
      return;
    }
    this.savingRMA.set(true);
    this.supplierService.createRMA({
      supplier_id: data.supplier_id,
      product_id: data.product_id,
      quantity: Number(data.quantity),
      reason: data.reason,
      photo_url: data.photo_url || undefined,
      purchase_order_id: data.purchase_order_id || undefined
    }).subscribe({
      next: (rma) => {
        this.savingRMA.set(false);
        this.toastService.showSuccess(`Logged RMA #${rma.rma_number} for defect review!`);
        this.closeRMAModal();
        this.loadRMAs();
      },
      error: () => {
        this.savingRMA.set(false);
        this.toastService.showError('Failed to log RMA claim');
      }
    });
  }

  resolveRMACredit(rma: any) {
    const amount = Number(prompt(`Enter credit note refund amount for RMA #${rma.rma_number}:`, '0')) || 0;
    this.supplierService.resolveRMA(rma.id, {
      status: 'refunded',
      resolution_notes: `Credit note issued by store manager for ${rma.quantity} units`,
      credit_amount: amount
    }).subscribe({
      next: () => {
        this.toastService.showSuccess(`Issued credit note for RMA #${rma.rma_number}!`);
        this.loadRMAs();
      },
      error: () => this.toastService.showError('Failed to resolve RMA')
    });
  }

  // ── Dock Schedule Hub ──
  loadDockSlots() {
    this.loadingDockSlots.set(true);
    this.supplierService.getBranchDockSlots().subscribe({
      next: (res) => {
        this.dockSlots.set(res || []);
        this.loadingDockSlots.set(false);
      },
      error: () => this.loadingDockSlots.set(false)
    });
  }

  // ── Merchant Broadcast Announcement ──
  openAnnouncementModal() {
    this.announcementTitle = '';
    this.announcementContent = '';
    this.announcementPriority = 'info';
    this.isAnnouncementModalOpen.set(true);
  }

  closeAnnouncementModal() {
    this.isAnnouncementModalOpen.set(false);
  }

  broadcastAnnouncement() {
    if (!this.announcementTitle || !this.announcementContent) {
      this.toastService.showError('Please provide a title and announcement message');
      return;
    }
    this.broadcastingAnnouncement.set(true);
    this.supplierService.createAnnouncement({
      title: this.announcementTitle,
      content: this.announcementContent,
      priority: this.announcementPriority
    }).subscribe({
      next: () => {
        this.broadcastingAnnouncement.set(false);
        this.toastService.showSuccess('Announcement broadcasted to all supplier dashboards!');
        this.closeAnnouncementModal();
      },
      error: () => {
        this.broadcastingAnnouncement.set(false);
        this.toastService.showError('Failed to broadcast announcement');
      }
    });
  }

  // ── 1-Click Invoice Disbursal ──
  disburseInvoice(invoiceId: string) {
    this.disbursingInvoiceId.set(invoiceId);
    this.supplierService.disburseInvoicePayout(invoiceId).subscribe({
      next: () => {
        this.disbursingInvoiceId.set(null);
        this.toastService.showSuccess('Disbursed invoice payout successfully!');
      },
      error: () => {
        this.disbursingInvoiceId.set(null);
        this.toastService.showError('Failed to disburse payout');
      }
    });
  }

  // ── KPIs ──
  get totalSuppliers() { return this.supplierService.suppliers().length; }
  get activeSuppliers() { return this.supplierService.suppliers().filter(s => s.is_active !== false).length; }
  get outstandingBalance() { return this.supplierService.suppliers().reduce((sum, s) => sum + (s.credit_balance ?? 0), 0); }
  get portalActiveCount() { return this.supplierService.suppliers().filter(s => !!s.portal_email).length; }

  get filteredSuppliers(): Supplier[] {
    let list = [...this.supplierService.suppliers()];

    // Text search
    const q = this.searchQuery().toLowerCase();
    if (q) {
      list = list.filter(s =>
        s.name?.toLowerCase().includes(q) ||
        s.email?.toLowerCase().includes(q) ||
        s.contact_email?.toLowerCase().includes(q) ||
        s.phone?.toLowerCase().includes(q)
      );
    }

    // Status filter
    if (this.statusFilter() === 'active') list = list.filter(s => s.is_active !== false);
    if (this.statusFilter() === 'inactive') list = list.filter(s => s.is_active === false);

    // Sort
    switch (this.sortOrder()) {
      case 'name': list.sort((a, b) => a.name.localeCompare(b.name)); break;
      case 'name_desc': list.sort((a, b) => b.name.localeCompare(a.name)); break;
      case 'balance': list.sort((a, b) => (b.credit_balance ?? 0) - (a.credit_balance ?? 0)); break;
      case 'newest': break; // default from API order
    }

    return list;
  }

  supplierInitials(name: string): string {
    return (name || '?').split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2);
  }

  paymentTermsLabel(terms: string | undefined): string {
    const map: Record<string, string> = {
      due_on_receipt: 'Due on Receipt',
      net15: 'Net 15', net30: 'Net 30', net45: 'Net 45',
      net60: 'Net 60', net90: 'Net 90',
    };
    return map[terms || ''] || terms || '—';
  }

  // ── CRUD ──
  openAddModal() {
    this.currentSupplier.set({ payment_terms: 'net30', is_active: true });
    this.isEditing.set(false);
    this.activeTab.set('details');
    this.isModalOpen.set(true);
  }

  openEditModal(supplier: Supplier) {
    this.currentSupplier.set({ ...supplier });
    this.isEditing.set(true);
    this.activeTab.set('details');
    this.isModalOpen.set(true);
  }

  closeModal() { this.isModalOpen.set(false); }

  saveSupplier() {
    this.saving.set(true);
    const s = this.currentSupplier();
    const obs = this.isEditing() && s.id
      ? this.supplierService.updateSupplier(s.id, s as any)
      : this.supplierService.createSupplier(s as any);

    obs.subscribe({
      next: () => { this.saving.set(false); this.closeModal(); this.toastService.showSuccess('Supplier saved!'); },
      error: () => { this.saving.set(false); this.toastService.showError('Failed to save supplier.'); }
    });
  }

  async deleteSupplier(supplier: Supplier) {
    if (!(await this.alertService.confirm(`Delete supplier "${supplier.name}"?`, 'Delete Supplier'))) return;
    this.deletingId.set(supplier.id);
    this.supplierService.deleteSupplier(supplier.id).subscribe({
      next: () => this.deletingId.set(null),
      error: () => { this.deletingId.set(null); this.toastService.showError('Failed to delete supplier.'); }
    });
  }

  updateField(field: keyof Supplier, value: any) {
    this.currentSupplier.update(s => ({ ...s, [field]: value }));
  }

  // ── Catalog ──
  openCatalog(supplier: Supplier) {
    this.currentSupplier.set({ ...supplier });
    this.isEditing.set(true);
    this.activeTab.set('catalog');
    this.supplierService.getSupplierProducts(supplier.id).subscribe(res => this.catalogProducts.set(res || []));
    if (this.allProducts().length === 0) {
      this.catalogService.getProducts({ limit: -1 }).subscribe(res => this.allProducts.set((res as any).data || []));
    }
    this.isModalOpen.set(true);
  }

  addCatalogProduct() {
    const s = this.currentSupplier();
    if (!s.id) return;
    this.addingToCatalog.set(true);
    this.supplierService.addSupplierProduct(s.id, this.newCatalogProduct()).subscribe({
      next: (res) => {
        this.addingToCatalog.set(false);
        this.catalogProducts.update(list => [...list, res]);
        this.newCatalogProduct.set({ product_id: '', supplier_sku: '', unit_cost: 0, min_order_qty: 1 });
        this.toastService.showSuccess('Product linked!');
      },
      error: () => { this.addingToCatalog.set(false); this.toastService.showError('Failed to add product.'); }
    });
  }

  async removeCatalogProduct(cp: any) {
    const s = this.currentSupplier();
    if (!s.id) return;
    if (!(await this.alertService.confirm(`Remove "${cp.product?.name || 'this product'}" from catalog?`))) return;
    this.supplierService.removeSupplierProduct(s.id, cp.product_id).subscribe({
      next: () => {
        this.catalogProducts.update(list => list.filter(p => p.id !== cp.id));
        this.toastService.showSuccess('Product removed from catalog.');
      },
      error: () => this.toastService.showError('Failed to remove product.')
    });
  }

  // ── Ledger ──
  openLedger(supplier: Supplier) {
    this.currentSupplier.set({ ...supplier });
    this.isEditing.set(true);
    this.activeTab.set('ledger');
    this.supplierService.getSupplierLedger(supplier.id).subscribe(res => this.ledgerEntries.set(res || []));
    this.isModalOpen.set(true);
  }

  addPayment() {
    const s = this.currentSupplier();
    if (!s.id) return;
    this.addingPayment.set(true);
    const p = this.newPayment();
    this.supplierService.addSupplierLedger(s.id, {
      entry_type: p.entry_type,
      amount: p.amount,
      reference_id: p.reference_id || undefined,
      notes: p.notes || undefined,
      transaction_date: p.transaction_date || undefined
    }).subscribe({
      next: (res) => {
        this.addingPayment.set(false);
        this.ledgerEntries.update(list => [res, ...list]);
        this.newPayment.set({ entry_type: 'payment', amount: 0, reference_id: '', notes: '', transaction_date: '' });
        // Refresh supplier balance
        this.supplierService.getSuppliers().subscribe();
        this.toastService.showSuccess('Transaction recorded!');
      },
      error: () => { this.addingPayment.set(false); this.toastService.showError('Failed to record transaction.'); }
    });
  }

  exportLedgerCSV() {
    const entries = this.ledgerEntries();
    const name = this.currentSupplier().name || 'supplier';
    const rows = [
      ['Date', 'Type', 'Reference', 'Amount', 'Balance', 'Notes'],
      ...entries.map(e => [
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
    a.download = `${name.replace(/\s+/g, '_')}_ledger.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  // ── Invite to Portal ──
  openInviteModal(supplier: Supplier) {
    this.inviteSupplier.set(supplier);
    this.inviteEmail = supplier.portal_email || supplier.email || supplier.contact_email || '';
    this.invitePassword = '';
    this.isInviteModalOpen.set(true);
  }

  closeInviteModal() { this.isInviteModalOpen.set(false); }

  sendInvite() {
    const s = this.inviteSupplier();
    if (!s) return;
    this.sendingInvite.set(true);
    this.supplierService.inviteToPortal(s.id, { email: this.inviteEmail, password: this.invitePassword }).subscribe({
      next: () => {
        this.sendingInvite.set(false);
        this.closeInviteModal();
        this.toastService.showSuccess(`Portal access granted to ${this.inviteEmail}`);
        // Update local supplier data
        this.supplierService.getSuppliers().subscribe();
      },
      error: (err) => {
        this.sendingInvite.set(false);
        this.toastService.showError('Failed to grant portal access.');
      }
    });
  }
}
