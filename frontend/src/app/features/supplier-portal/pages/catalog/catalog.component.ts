import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SupplierPortalService, SupplierProduct, SupplierPriceChangeRequest } from '../../services/supplier-portal.service';
import { AppCurrencyPipe } from '../../../../core/pipes/app-currency.pipe';
import { ToastService } from '../../../../core/services/toast';

@Component({
  selector: 'app-supplier-portal-catalog',
  standalone: true,
  imports: [CommonModule, AppCurrencyPipe, FormsModule],
  templateUrl: './catalog.component.html'
})
export class SupplierPortalCatalogComponent implements OnInit {
  portalService = inject(SupplierPortalService);
  private toast = inject(ToastService);

  catalog = signal<SupplierProduct[]>([]);
  priceRequests = signal<SupplierPriceChangeRequest[]>([]);
  activeTab = signal<'catalog' | 'requests'>('catalog');
  loading = signal<boolean>(false);

  showPriceModal = signal<boolean>(false);
  selectedProduct: SupplierProduct | null = null;
  proposedCost = 0;
  effectiveDate = '';
  reason = '';

  ngOnInit() {
    this.loadData();
  }

  loadData() {
    this.loading.set(true);
    this.portalService.getCatalog().subscribe({
      next: (res) => {
        this.catalog.set(res || []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });

    this.portalService.getPriceRequests().subscribe({
      next: (reqs) => this.priceRequests.set(reqs || [])
    });
  }

  openPriceModal(item: SupplierProduct) {
    this.selectedProduct = item;
    this.proposedCost = item.unit_cost;
    const date = new Date();
    date.setDate(date.getDate() + 14);
    this.effectiveDate = date.toISOString().split('T')[0];
    this.reason = '';
    this.showPriceModal.set(true);
  }

  submitPriceRequest() {
    if (!this.selectedProduct || this.proposedCost <= 0 || !this.reason) {
      this.toast.showError('Please provide a valid proposed cost and justification reason');
      return;
    }

    const payload: Partial<SupplierPriceChangeRequest> = {
      product_id: this.selectedProduct.product_id,
      current_cost: this.selectedProduct.unit_cost,
      proposed_cost: Number(this.proposedCost),
      effective_date: new Date(this.effectiveDate).toISOString(),
      reason: this.reason
    };

    this.portalService.createPriceRequest(payload).subscribe({
      next: () => {
        this.toast.showSuccess('Price change proposal submitted for merchant review!');
        this.showPriceModal.set(false);
        this.activeTab.set('requests');
        this.loadData();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to submit price request')
    });
  }

  // ── Bulk CSV Import ──
  showBulkModal = signal<boolean>(false);
  csvPreview = signal<any[]>([]);
  csvError = signal<string>('');
  importingBulk = signal<boolean>(false);

  openBulkModal() {
    this.csvPreview.set([]);
    this.csvError.set('');
    this.showBulkModal.set(true);
  }

  onCSVFile(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input?.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const text = e.target?.result as string;
        const lines = text.trim().split('\n');
        const headers = lines[0].split(',').map(h => h.trim().toLowerCase().replace(/[^a-z_]/g, '_'));
        const rows = lines.slice(1).map(line => {
          const cols = line.split(',');
          const row: any = {};
          headers.forEach((h, i) => row[h] = cols[i]?.trim() || '');
          return {
            sku: row['sku'] || row['product_sku'] || '',
            product_name: row['name'] || row['product_name'] || '',
            unit_cost: parseFloat(row['cost'] || row['unit_cost'] || '0'),
            min_order_qty: parseFloat(row['min_qty'] || row['min_order_qty'] || '1'),
          };
        }).filter(r => r.sku && r.product_name);
        this.csvPreview.set(rows);
        this.csvError.set('');
      } catch {
        this.csvError.set('Failed to parse CSV. Ensure columns: sku, name/product_name, cost/unit_cost, min_qty/min_order_qty');
      }
    };
    reader.readAsText(file);
  }

  submitBulkImport() {
    const items = this.csvPreview();
    if (!items.length) { this.toast.showError('No valid rows to import'); return; }
    this.importingBulk.set(true);
    this.portalService.bulkImportCatalog(items).subscribe({
      next: (res) => {
        this.toast.showSuccess(res.message);
        this.importingBulk.set(false);
        this.showBulkModal.set(false);
        this.loadData();
      },
      error: (err) => {
        this.toast.showError(err.error?.error || 'Bulk import failed');
        this.importingBulk.set(false);
      }
    });
  }
}
