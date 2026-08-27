import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { CatalogService } from '../../../core/services/catalog.service';
import { StockBatch } from '../../../core/models/inventory.models';
import { Product } from '../../../core/models/product.models';
import { Topbar } from '../../../core/layout/topbar/topbar';
import { AlertService } from '../../../core/services/alert.service';
import { SearchableSelectComponent } from '../../../shared/components/searchable-select/searchable-select';

@Component({
  selector: 'app-batch-tracker',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, SearchableSelectComponent],
  templateUrl: './batch-tracker.html'
})
export class BatchTracker implements OnInit {
  private catalogService = inject(CatalogService);
  private alertService = inject(AlertService);

  // Data
  expiringBatches = signal<StockBatch[]>([]);
  allProducts = signal<Product[]>([]);
  loading = signal<boolean>(true);

  productOptions = computed(() =>
    this.allProducts().map(p => ({
      label: p.name,
      value: p.id,
      sublabel: `SKU: ${p.sku}${p.current_stock !== undefined ? ` • Stock: ${p.current_stock}` : ''}`
    }))
  );

  // Modals
  isBatchModalOpen = signal<boolean>(false);
  
  // Form State
  selectedProductId = signal<string>('');
  batchForm = signal({
    batch_number: '',
    quantity: 0,
    expiry_date: '',
    manufacture_date: ''
  });

  // Derived state for dashboard
  expiredBatches = computed(() => {
    const today = new Date();
    return this.expiringBatches().filter(b => b.expiry_date && new Date(b.expiry_date) < today);
  });
  
  expiringSoonBatches = computed(() => {
    const today = new Date();
    const thirtyDays = new Date();
    thirtyDays.setDate(today.getDate() + 30);
    return this.expiringBatches().filter(b => {
      if (!b.expiry_date) return false;
      const exp = new Date(b.expiry_date);
      return exp >= today && exp <= thirtyDays;
    });
  });

  ngOnInit() {
    this.loadData();
  }

  loadData() {
    this.loading.set(true);
    // Load expiring batches globally (next 90 days to show on dashboard)
    this.catalogService.getExpiringBatches(90).subscribe({
      next: (res) => {
        this.expiringBatches.set(res.data || []);
      },
      complete: () => {
        // Also load all products for the select dropdown without pagination
        this.catalogService.getProducts({ limit: -1 }).subscribe(prodRes => {
           this.allProducts.set(prodRes.data || []);
           this.loading.set(false);
        });
      }
    });
  }

  openBatchModal() {
    this.batchForm.set({
      batch_number: `BATCH-${Date.now().toString().slice(-6)}`,
      quantity: 1,
      expiry_date: '',
      manufacture_date: ''
    });
    this.selectedProductId.set('');
    this.isBatchModalOpen.set(true);
  }

  closeBatchModal() {
    this.isBatchModalOpen.set(false);
  }

  saveBatch() {
    if (!this.selectedProductId() || !this.batchForm().batch_number) return;
    
    this.catalogService.createBatch(this.selectedProductId(), {
      ...this.batchForm()
    }).subscribe({
      next: () => {
        this.closeBatchModal();
        this.loadData(); // refresh list
      },
      error: (err) => console.error(err)
    });
  }

  async deleteBatch(batchId: string | undefined) {
    if (!batchId) return;
    if (await this.alertService.confirm('Are you sure you want to delete this batch?', 'Delete Batch')) {
      this.catalogService.deleteBatch(batchId).subscribe({
        next: () => this.loadData()
      });
    }
  }

  getDaysUntilExpiry(dateStr: string | undefined): number {
    if (!dateStr) return 999;
    const today = new Date();
    today.setHours(0,0,0,0);
    const exp = new Date(dateStr);
    exp.setHours(0,0,0,0);
    const diffTime = exp.getTime() - today.getTime();
    return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  }
}
