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
}
