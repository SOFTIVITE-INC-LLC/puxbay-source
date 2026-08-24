import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { Router } from '@angular/router';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { InventoryService, POCreateInput, TransferCreateInput } from '../../../core/services/inventory.service';
import { CatalogService } from '../../../core/services/catalog.service';
import { SupplierService } from '../../../core/services/supplier.service';
import { BranchService } from '../../../core/services/branch.service';
import { ToastService } from '../../../core/services/toast';

import { SearchableSelectComponent } from '../../../shared/components/searchable-select/searchable-select';

type ActiveView = 'transfers' | 'purchase-orders' | 'stocktakes' | 'low-stock';

@Component({
  selector: 'app-supply-chain',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe, SearchableSelectComponent],
  templateUrl: './supply-chain.html',
})
export class SupplyChain implements OnInit {
  private inventoryService = inject(InventoryService);
  private catalogService = inject(CatalogService);
  private supplierService = inject(SupplierService);
  private branchService = inject(BranchService);
  private toastr = inject(ToastService);
  private router = inject(Router);

  readonly Math = Math;

  tabs = [
    { id: 'transfers' as const, label: 'Stock Transfers' },
    { id: 'purchase-orders' as const, label: 'Purchase Orders' },
    { id: 'stocktakes' as const, label: 'Stocktakes' },
    { id: 'low-stock' as const, label: 'Low Stock Alerts' },
  ];

  activeView = signal<ActiveView>('transfers');
  isDrawerOpen = signal(false);

  transfers = signal<any[]>([]);
  purchaseOrders = signal<any[]>([]);
  stocktakes = signal<any[]>([]);

  products = this.catalogService.products;
  suppliers = this.supplierService.suppliers;
  branches = this.branchService.branches;

  productOptions = computed(() => this.products().map(p => ({ label: p.name, value: p.id })));
  supplierOptions = computed(() => this.suppliers().map(s => ({ label: s.name, value: s.id })));
  branchOptions = computed(() => this.branches().map(b => ({ label: b.name, value: b.id })));

  lowStockProducts = computed(() =>
    this.products().filter(p => p.track_inventory && (p.current_stock || 0) <= (p.reorder_level || 0))
  );

  // Group low-stock products by their preferred supplier
  lowStockBySupplier = computed(() => {
    const map = new Map<string, { supplier: any; products: any[] }>();
    for (const p of this.lowStockProducts()) {
      const supplierId = p.supplier_id || '__none__';
      if (!map.has(supplierId)) {
        const supplier = this.suppliers().find(s => s.id === supplierId);
        map.set(supplierId, { supplier: supplier || null, products: [] });
      }
      map.get(supplierId)!.products.push(p);
    }
    return Array.from(map.values());
  });

  newTransfer: any = { reference_no: '', from_branch_id: '', to_branch_id: '', notes: '' };
  newTransferItems: { product_id: string; quantity: number }[] = [{ product_id: '', quantity: 1 }];

  newPO: any = { po_number: '', supplier_id: '', notes: '' };
  newPOItems: { product_id: string; quantity_ordered: number; unit_cost: number }[] = [{ product_id: '', quantity_ordered: 1, unit_cost: 0 }];

  selectedTransfer = signal<any>(null);
  selectedPO = signal<any>(null);

  getProductName(id: string): string {
    return this.products().find(p => p.id === id)?.name || id;
  }

  getBranchName(id: string): string {
    return this.branches().find(b => b.id === id)?.name || id;
  }

  getSupplierName(id: string): string {
    return this.suppliers().find(s => s.id === id)?.name || id;
  }

  newStocktake: any = { name: '', notes: '' };

  poTotal = computed(() => this.newPOItems.reduce((sum, i) => sum + (i.quantity_ordered * i.unit_cost), 0));

  ngOnInit() {
    this.loadCurrentTab();
    this.catalogService.getProducts({ limit: -1 }).subscribe();
    this.supplierService.getSuppliers({ limit: -1 }).subscribe();
    this.branchService.getBranches({ limit: -1 }).subscribe();
  }

  loadCurrentTab() {
    this.inventoryService.listTransfers().subscribe({ next: (r: any) => this.transfers.set(r?.data || []), error: () => {} });
    this.inventoryService.listPOs().subscribe({ next: (r: any) => this.purchaseOrders.set(r?.data || []), error: () => {} });
    this.inventoryService.listStocktakes().subscribe({ next: (r: any) => this.stocktakes.set(r?.data || []), error: () => {} });
  }

  openCreateDrawer() {
    this.selectedTransfer.set(null);
    this.selectedPO.set(null);
    this.isDrawerOpen.set(true);
  }

  viewTransfer(t: any) {
    this.inventoryService.getTransfer(t.id).subscribe(transfer => {
      this.selectedTransfer.set(transfer);
      this.selectedPO.set(null);
      this.isDrawerOpen.set(true);
    });
  }

  viewPO(po: any) {
    this.inventoryService.getPO(po.id).subscribe(fullPO => {
      this.selectedPO.set(fullPO);
      this.selectedTransfer.set(null);
      this.isDrawerOpen.set(true);
    });
  }

  createTransfer() {
    if (!this.newTransfer.reference_no || !this.newTransfer.from_branch_id || !this.newTransfer.to_branch_id) {
      this.toastr.showError('Please fill in all required fields (Reference, From Branch, To Branch).');
      return;
    }

    const validItems = this.newTransferItems.filter(i => i.product_id && i.quantity > 0);
    if (validItems.length === 0) {
      this.toastr.showError('Please add at least one valid product to transfer.');
      return;
    }

    const input: TransferCreateInput = {
      ...this.newTransfer,
      items: validItems
    };
    this.inventoryService.createTransfer(input).subscribe({
      next: (t) => {
        this.transfers.update(list => [t, ...list]);
        this.isDrawerOpen.set(false);
        this.toastr.showSuccess('Stock transfer created!');
        this.newTransfer = { reference_no: '', from_branch_id: '', to_branch_id: '', notes: '' };
        this.newTransferItems = [{ product_id: '', quantity: 1 }];
      },
      error: (err) => this.toastr.showError(err.error?.error || 'Failed to create transfer.')
    });
  }

  createPO() {
    if (!this.newPO.po_number || !this.newPO.supplier_id) {
      this.toastr.showError('Please fill in all required fields (PO Number, Supplier).');
      return;
    }

    const validItems = this.newPOItems.filter(i => i.product_id && i.quantity_ordered > 0);
    if (validItems.length === 0) {
      this.toastr.showError('Please add at least one valid product.');
      return;
    }

    const input: POCreateInput = {
      ...this.newPO,
      branch_id: this.branches()[0]?.id || '',
      items: validItems
    };
    this.inventoryService.createPO(input).subscribe({
      next: (po) => {
        this.purchaseOrders.update(list => [po, ...list]);
        this.isDrawerOpen.set(false);
        this.toastr.showSuccess('Purchase order created!');
        this.newPO = { po_number: '', supplier_id: '', notes: '' };
        this.newPOItems = [{ product_id: '', quantity_ordered: 1, unit_cost: 0 }];
      },
      error: (err) => this.toastr.showError(err.error?.error || 'Failed to create PO.')
    });
  }

  createStocktake() {
    this.inventoryService.createStocktake(this.newStocktake).subscribe({
      next: (s: any) => {
        this.stocktakes.update(list => [s, ...list]);
        this.isDrawerOpen.set(false);
        this.toastr.showSuccess('Stocktake session started!');
        this.newStocktake = { name: '', notes: '' };
      },
      error: () => this.toastr.showError('Failed to create stocktake session.')
    });
  }

  viewStocktake(s: any) {
    this.router.navigate(['/inventory/stocktake', s.id]);
  }

  approveTransfer(t: any) {
    this.inventoryService.approveTransfer(t.id).subscribe({
      next: () => { t.status = 'approved'; this.toastr.showSuccess('Transfer approved'); },
      error: () => this.toastr.showError('Failed to approve')
    });
  }

  shipTransfer(t: any) {
    this.inventoryService.shipTransfer(t.id).subscribe({
      next: () => { t.status = 'shipped'; this.toastr.showSuccess('Transfer shipped'); },
      error: () => this.toastr.showError('Failed to ship')
    });
  }

  receiveTransfer(t: any) {
    this.inventoryService.receiveTransfer(t.id).subscribe({
      next: () => { t.status = 'received'; this.toastr.showSuccess('Transfer received'); },
      error: () => this.toastr.showError('Failed to complete')
    });
  }

  quickReceivePO(po: any) {
    if (!po.items) {
      this.toastr.showError('No items to receive');
      return;
    }
    const input = {
      items: po.items.map((i: any) => ({ item_id: i.id, quantity_received: i.quantity_ordered - i.quantity_received }))
    };
    this.inventoryService.receivePO(po.id, input).subscribe({
      next: () => { po.status = 'received'; this.toastr.showSuccess('PO Received'); },
      error: () => this.toastr.showError('Failed to receive PO')
    });
  }

  finalizeStocktake(s: any) {
    this.inventoryService.finalizeStocktake(s.id).subscribe({
      next: () => { s.status = 'completed'; this.toastr.showSuccess('Stocktake finalized'); },
      error: () => this.toastr.showError('Failed to finalize')
    });
  }

  // 1-Click generate PO from low-stock group
  generatePOFromLowStock(group: { supplier: any; products: any[] }) {
    if (!group.supplier) {
      this.toastr.showError('No supplier assigned to these products. Assign a supplier first.');
      return;
    }
    this.activeView.set('purchase-orders');
    this.selectedPO.set(null);
    this.newPO = {
      po_number: `PO-${new Date().getFullYear()}-${Date.now().toString().slice(-4)}`,
      supplier_id: group.supplier.id,
      notes: `Auto-generated from low stock alert`
    };
    this.newPOItems = group.products.map(p => ({
      product_id: p.id,
      quantity_ordered: Math.max(1, (p.reorder_level || 0) - (p.current_stock || 0)),
      unit_cost: p.cost_price || 0
    }));
    this.isDrawerOpen.set(true);
  }

  // Transfer status timeline steps
  transferSteps = ['requested', 'approved', 'shipped', 'received'];
  poSteps = ['pending', 'ordered', 'partial', 'received'];

  getStepIndex(status: string, steps: string[]): number {
    const i = steps.indexOf(status);
    return i >= 0 ? i : 0;
  }
}
