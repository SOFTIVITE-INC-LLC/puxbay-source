import { ToastService } from '../../../core/services/toast';
import { ExportService } from '../../../core/services/export.service';
import { Component, inject, OnInit, signal } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { CatalogService } from '../../../core/services/catalog.service';
import { BranchService } from '../../../core/services/branch.service';
import { Product, Category } from '../../../core/models/models';
import { A11yModule } from '@angular/cdk/a11y';
import { forkJoin, Subject } from 'rxjs';
import { debounceTime } from 'rxjs/operators';

@Component({
  selector: 'app-inventory',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, A11yModule, AppCurrencyPipe],
  templateUrl: './inventory.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .dark .glass-panel {
      background: rgba(0, 0, 0, 0.2);
    }
  `,
})
export class Inventory implements OnInit {
  toastService = inject(ToastService);
  exportService = inject(ExportService);
  catalogService = inject(CatalogService);
  branchService = inject(BranchService);
  private router = inject(Router);

  readonly Math = Math;
  isModalOpen = signal(false);
  isBarcodeModalOpen = signal(false);
  isImportModalOpen = signal(false);
  modalTitle = signal('Add Product');
  selectedFile = signal<File | null>(null);
  
  currentProduct = signal<Partial<Product>>({
    is_active: true,
    is_online: true,
    track_inventory: true,
    current_stock: 0,
    cost_price: 0,
    selling_price: 0,
    wholesale_price: 0,
    category_id: undefined,
    description: '',
    stock_unit: 'pcs'
  });
  
  barcodeData = signal<{image: string, data: string} | null>(null);

  // Stock Adjustment State
  isAdjustStockModalOpen = signal(false);
  productToAdjust = signal<Product | null>(null);
  adjustQuantity = signal<number>(0);
  adjustReason = signal<string>('restock');
  adjustingStock = signal(false);

  // Bulk Selection State
  selectedProductIds = signal<Set<string>>(new Set());
  performingBulkAction = signal(false);

  isHistoryPanelOpen = signal(false);
  productForHistory = signal<Product | null>(null);
  historyData = signal<any[] | null>(null);
  loadingHistory = signal(false);

  // Delete Confirm Modal state
  showDeleteConfirm = signal(false);
  productToDelete = signal<string | null>(null);
  showBulkDeleteConfirm = signal(false);
  bulkDeleteCount = signal(0);

  searchQuery = signal('');
  searchSubject = new Subject<string>();
  viewMode = signal<'table' | 'grid'>('table');
  selectedCategory = signal('');
  sortBy = signal('name');
  filterByLowStock = signal(false);

  // Pagination State
  currentPage = signal(1);
  limit = signal(50);
  get totalPages() {
    return Math.ceil(this.catalogService.totalProducts() / this.limit()) || 1;
  }

  get totalActive() { return this.catalogService.products().filter(p => p.is_active).length; }
  get totalInactive() { return this.catalogService.products().filter(p => !p.is_active).length; }
  get lowStock() { return this.catalogService.products().filter(p => (p.current_stock || 0) <= (p.reorder_level || 0)).length; }

  ngOnInit() {
    this.searchSubject.pipe(debounceTime(300)).subscribe(q => {
      this.searchQuery.set(q);
      this.onFilterChange();
    });
    this.loadProducts();
    this.catalogService.getCategories().subscribe();
  }

  private _lastScrollTime = 0;

  onScroll(event: Event) {
    const el = event.target as HTMLElement;
    if (el.scrollHeight - el.scrollTop <= el.clientHeight + 150) {
      if (!this.catalogService.loading() && this.currentPage() < this.totalPages) {
        // Debounce to prevent multiple fires
        const now = Date.now();
        if (now - this._lastScrollTime > 500) {
          this._lastScrollTime = now;
          this.nextPage();
        }
      }
    }
  }

  loadProducts() {
    const branch = this.branchService.activeBranch();
    const params: any = {
      page: this.currentPage(),
      limit: this.limit()
    };
    if (branch?.id) params['branch_id'] = branch.id;
    if (this.searchQuery()) params['q'] = this.searchQuery();
    if (this.selectedCategory()) params['category_id'] = this.selectedCategory();
    
    this.catalogService.getProducts(params).subscribe();
  }

  // Pagination Methods
  nextPage() {
    if (this.currentPage() < this.totalPages && !this.catalogService.loading()) {
      this.currentPage.update(p => p + 1);
      this.loadProducts();
    }
  }

  prevPage() {
    if (this.currentPage() > 1) {
      this.currentPage.update(p => p - 1);
      this.loadProducts();
    }
  }

  onFilterChange() {
    this.currentPage.set(1);
    this.loadProducts();
  }

  get filteredProducts() {
    let list = this.catalogService.products();
    const showLowStock = this.filterByLowStock();
    
    if (showLowStock) {
      list = list.filter(p => (p.current_stock || 0) <= (p.reorder_level || 0));
    }
    
    if (this.sortBy() === 'price_asc') list = [...list].sort((a, b) => (a.selling_price || 0) - (b.selling_price || 0));
    if (this.sortBy() === 'price_desc') list = [...list].sort((a, b) => (b.selling_price || 0) - (a.selling_price || 0));
    if (this.sortBy() === 'stock') list = [...list].sort((a, b) => (a.current_stock || 0) - (b.current_stock || 0));
    if (this.sortBy() === 'name') list = [...list].sort((a, b) => a.name.localeCompare(b.name));
    return list;
  }

  openAddModal() {
    this.router.navigate(['/inventory/add']);
  }

  openEditModal(product: Product) {
    this.router.navigate(['/inventory/edit', product.id]);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  saveProduct() {
    const p = this.currentProduct();
    if (p.id) {
      this.catalogService.updateProduct(p.id!, p as any).subscribe(() => this.closeModal());
    } else {
      this.catalogService.createProduct(p as any).subscribe(() => this.closeModal());
    }
  }

  deleteProduct(id: string) {
    this.productToDelete.set(id);
    this.showDeleteConfirm.set(true);
  }

  confirmDelete() {
    const id = this.productToDelete();
    if (!id) return;
    this.catalogService.deleteProduct(id).subscribe();
    this.showDeleteConfirm.set(false);
    this.productToDelete.set(null);
  }

  cancelDelete() {
    this.showDeleteConfirm.set(false);
    this.productToDelete.set(null);
  }

  toggleActive(product: Product) {
    const originalState = product.is_active;
    product.is_active = !product.is_active;
    this.catalogService.updateProduct(product.id, { is_active: product.is_active } as any).subscribe({
      error: () => {
        product.is_active = originalState;
        this.toastService.showError('Failed to update status.');
      }
    });
  }

  generateBarcode(id: string) {
    // We would use ApiService directly for raw HTTP calls if CatalogService didn't wrap it, 
    // but we can just use fetch or api directly. For now, mock or handle via service.
    // Assuming backend returns base64
    const api = (this.catalogService as any).api;
    api.get(`/products/${id}/barcode?format=base64`).subscribe((res: any) => {
      this.barcodeData.set(res);
      this.isBarcodeModalOpen.set(true);
    });
  }

  closeBarcodeModal() {
    this.isBarcodeModalOpen.set(false);
    this.barcodeData.set(null);
  }

  openImportModal() {
    this.isImportModalOpen.set(true);
    this.selectedFile.set(null);
  }

  closeImportModal() {
    this.isImportModalOpen.set(false);
    this.selectedFile.set(null);
  }

  onFileSelected(event: any) {
    const file = event.target.files[0];
    if (file) {
      this.selectedFile.set(file);
    }
  }

  uploadProducts() {
    const file = this.selectedFile();
    if (!file) return;

    this.catalogService.importProducts(file).subscribe({
      next: (res) => {
        this.toastService.showSuccess(`Successfully imported products!`);
        this.closeImportModal();
        this.loadProducts(); // refresh
      },
      error: (err) => {
        this.toastService.showError('Failed to import products. Please check the file format.');
        console.error(err);
      }
    });
  }

  downloadTemplate() {
    const csvContent = "name,sku,category,selling_price,wholesale_price,min_wholesale_qty,cost_price,stock_quantity,low_stock_threshold,barcode,expiry_date,batch_number,invoice_waybill_number,description,is_active,image_url,mfg_date,country_of_origin,manufacturer_name,manufacturer_address\nSample Product,SKU-001,General,15.00,12.00,10,10.00,100,10,123456789,2025-12-31,BATCH001,INV12345,A sample product description,TRUE,https://example.com/img.jpg,2024-01-01,USA,Acme Corp,123 Business Rd";
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement("a");
    const url = URL.createObjectURL(blob);
    link.setAttribute("href", url);
    link.setAttribute("download", "product_import_template.csv");
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  // --- Bulk Actions ---
  toggleSelectAll() {
    const allIds = this.filteredProducts.map(p => p.id!);
    const current = this.selectedProductIds();
    if (current.size === allIds.length) {
      this.selectedProductIds.set(new Set());
    } else {
      this.selectedProductIds.set(new Set(allIds));
    }
  }

  toggleSelect(id: string) {
    const current = new Set(this.selectedProductIds());
    if (current.has(id)) current.delete(id);
    else current.add(id);
    this.selectedProductIds.set(current);
  }

  executeBulkAction(action: 'activate' | 'deactivate' | 'delete') {
    const ids = Array.from(this.selectedProductIds());
    if (!ids.length) return;
    
    if (action === 'delete') {
      this.bulkDeleteCount.set(ids.length);
      this.showBulkDeleteConfirm.set(true);
      return;
    }

    this.performingBulkAction.set(true);
    
    const observables = ids.map(id => {
      return this.catalogService.updateProduct(id, { is_active: action === 'activate' } as any);
    });

    forkJoin(observables).subscribe({
      next: () => {
        this.selectedProductIds.set(new Set());
        this.performingBulkAction.set(false);
      },
      error: () => {
        this.performingBulkAction.set(false);
      }
    });
  }

  confirmBulkDelete() {
    const ids = Array.from(this.selectedProductIds());
    if (!ids.length) return;
    this.showBulkDeleteConfirm.set(false);
    this.performingBulkAction.set(true);
    const observables = ids.map(id => this.catalogService.deleteProduct(id));
    forkJoin(observables).subscribe({
      next: () => {
        this.selectedProductIds.set(new Set());
        this.performingBulkAction.set(false);
      },
      error: () => { this.performingBulkAction.set(false); }
    });
  }

  cancelBulkDelete() {
    this.showBulkDeleteConfirm.set(false);
  }

  // --- Stock Adjustment ---
  openAdjustStockModal(product: Product) {
    this.productToAdjust.set(product);
    this.adjustQuantity.set(0);
    this.adjustReason.set('restock');
    this.isAdjustStockModalOpen.set(true);
  }

  closeAdjustStockModal() {
    this.isAdjustStockModalOpen.set(false);
    this.productToAdjust.set(null);
  }

  saveStockAdjustment() {
    const p = this.productToAdjust();
    const qty = this.adjustQuantity();
    const reason = this.adjustReason();
    if (!p || !p.id || qty === 0) return;

    this.adjustingStock.set(true);
    this.catalogService.adjustStock(p.id, qty, reason).subscribe({
      next: () => {
        this.adjustingStock.set(false);
        this.closeAdjustStockModal();
      },
      error: () => {
        this.adjustingStock.set(false);
        this.toastService.showError('Failed to adjust stock.');
      }
    });
  }

  // --- History ---
  openHistoryPanel(product: Product) {
    this.productForHistory.set(product);
    this.isHistoryPanelOpen.set(true);
    this.loadingHistory.set(true);
    this.historyData.set(null);
    this.catalogService.getProductHistory(product.id!).subscribe({
      next: (res) => {
        this.historyData.set(res.history || []);
        this.loadingHistory.set(false);
      },
      error: () => {
        this.loadingHistory.set(false);
        this.toastService.showError('Failed to load history.');
      }
    });
  }

  closeHistoryPanel() {
    this.isHistoryPanelOpen.set(false);
    this.productForHistory.set(null);
    this.historyData.set(null);
  }

  toggleLowStockFilter() {
    this.filterByLowStock.set(!this.filterByLowStock());
  }

  exportCsv() {
    const branchId = this.branchService.activeBranch()?.id;
    this.exportService.exportProducts(branchId);
  }
}
