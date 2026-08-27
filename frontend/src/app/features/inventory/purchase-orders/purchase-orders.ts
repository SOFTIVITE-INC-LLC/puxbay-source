import { ToastService } from '../../../core/services/toast';
import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { InventoryService, POCreateInput } from '../../../core/services/inventory.service';
import { SupplierService } from '../../../core/services/supplier.service';
import { CatalogService } from '../../../core/services/catalog.service';
import { PurchaseOrder } from '../../../core/models/inventory.models';
import { Supplier } from '../../../core/models/financial.models';
import { SettingsService } from '../../../core/services/settings.service';
import { SearchableSelectComponent } from '../../../shared/components/searchable-select/searchable-select';

@Component({
  selector: 'app-purchase-orders',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe, SearchableSelectComponent],
  templateUrl: './purchase-orders.html',
})
export class PurchaseOrders implements OnInit {
  toastService = inject(ToastService);
  inventoryService = inject(InventoryService);
  supplierService = inject(SupplierService);
  catalogService = inject(CatalogService);
  settingsService = inject(SettingsService);

  pos = signal<PurchaseOrder[]>([]);
  suppliers = signal<Supplier[]>([]);
  products = signal<any[]>([]);
  loading = signal(false);

  productOptions = computed(() =>
    this.products().map(p => ({
      label: p.name,
      value: p.id,
      sublabel: `Stock: ${p.current_stock ?? 0}${p.sku ? ` • SKU: ${p.sku}` : ''}`
    }))
  );

  supplierOptions = computed(() =>
    this.suppliers().map(s => ({
      label: s.name,
      value: s.id
    }))
  );

  // Modal states
  isCreateModalOpen = signal(false);
  isReceiveModalOpen = signal(false);
  creating = signal(false);
  receiving = signal(false);
  
  selectedPO = signal<PurchaseOrder | null>(null);

  // Form states
  newPO = signal<Partial<POCreateInput>>({ items: [] });
  newPOItems = signal<any[]>([]); // UI representation of items
  
  receiveItems = signal<any[]>([]); // Form state for receiving

  ngOnInit() {
    this.loadData();
  }

  loadData() {
    this.loading.set(true);
    this.inventoryService.listPOs().subscribe({
      next: (res: any) => {
        this.pos.set(res?.data || []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });

    this.supplierService.getSuppliers({ limit: -1 }).subscribe(res => this.suppliers.set(res || []));
    this.catalogService.getProducts({ limit: -1 }).subscribe(res => this.products.set(res.data || []));
  }

  getSupplierName(id: string): string {
    return this.suppliers().find(s => s.id === id)?.name || 'Unknown Supplier';
  }

  getProductName(id: string): string {
    return this.products().find(p => p.id === id)?.name || 'Unknown Product';
  }

  // --- Create PO ---
  openCreateModal() {
    this.newPO.set({ po_number: 'PO-' + Date.now(), items: [] });
    this.newPOItems.set([]);
    this.isCreateModalOpen.set(true);
  }

  closeCreateModal() {
    this.isCreateModalOpen.set(false);
  }

  addItem() {
    this.newPOItems.update(items => [...items, { product_id: '', quantity_ordered: 1, unit_cost: 0 }]);
  }

  removeItem(index: number) {
    this.newPOItems.update(items => items.filter((_, i) => i !== index));
  }

  createPO() {
    this.creating.set(true);
    const payload = {
      ...this.newPO(),
      items: this.newPOItems()
    };
    
    this.inventoryService.createPO(payload as any).subscribe({
      next: (res) => {
        this.pos.update(list => [res, ...list]);
        this.closeCreateModal();
        this.creating.set(false);
      },
      error: () => {
        this.creating.set(false);
        this.toastService.showError('Failed to create PO');
      }
    });
  }

  // --- Receive PO ---
  openReceiveModal(po: PurchaseOrder) {
    this.selectedPO.set(po);
    const itemsToReceive = (po.items || []).map(item => ({
      item_id: item.id,
      product_name: this.getProductName(item.product_id),
      quantity_ordered: item.quantity_ordered,
      quantity_previously_received: item.quantity_received || 0,
      quantity_to_receive: item.quantity_ordered - (item.quantity_received || 0) // Default to remaining
    }));
    this.receiveItems.set(itemsToReceive);
    this.isReceiveModalOpen.set(true);
  }

  closeReceiveModal() { this.isReceiveModalOpen.set(false); }

  receivePO() {
    const po = this.selectedPO();
    if (!po) return;

    this.receiving.set(true);
    const payload = {
      items: this.receiveItems().map(r => ({
        item_id: r.item_id,
        quantity_received: Number(r.quantity_to_receive)
      }))
    };

    this.inventoryService.receivePO(po.id, payload).subscribe({
      next: () => {
        this.receiving.set(false);
        this.closeReceiveModal();
        this.loadData();
      },
      error: () => {
        this.receiving.set(false);
        this.toastService.showError('Failed to receive PO');
      }
    });
  }

  // --- Helpers ---
  statusLabel(status: string): string {
    const map: Record<string, string> = {
      draft: 'Draft',
      issued: 'Issued',
      partially_received: 'Partially Received',
      received: 'Received',
      cancelled: 'Cancelled'
    };
    return map[status] || status;
  }
}
