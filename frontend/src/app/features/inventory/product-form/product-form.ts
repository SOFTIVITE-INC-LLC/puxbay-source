import { ToastService } from '../../../core/services/toast';
import { Component, inject, OnInit, signal } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, ActivatedRoute, RouterModule } from '@angular/router';
import { CatalogService } from '../../../core/services/catalog.service';
import { Product } from '../../../core/models/models';

@Component({
  selector: 'app-product-form',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe],
  templateUrl: './product-form.html',
})
export class ProductForm implements OnInit {
  toastService = inject(ToastService);
  catalogService = inject(CatalogService);
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  readonly Math = Math;

  isEditing = signal(false);
  isSaving = signal(false);
  productId = signal<string | null>(null);
  activeTab: 'basic' | 'pricing' | 'inventory' | 'details' = 'basic';

  form = signal<Partial<Product>>({
    name: '',
    description: '',
    sku: '',
    barcode: null,
    category_id: null,
    cost_price: 0,
    selling_price: 0,
    wholesale_price: 0,
    tax_rate: 0,
    track_inventory: true,
    current_stock: 0,
    reorder_level: 0,
    stock_unit: 'pcs',
    is_active: true,
    is_online: false,
    brand: null,
    color: null,
    weight: null,
    minimum_wholesale_quantity: 1,
    batch_number: '',
    invoice_waybill_number: '',
    country_of_origin: '',
    manufacturer_name: '',
    manufacturer_address: '',
    expiry_date: '',
    manufacturing_date: '',
  } as any);

  ngOnInit() {
    this.catalogService.getCategories().subscribe();

    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.isEditing.set(true);
      this.productId.set(id);
      this.catalogService.getProduct(id).subscribe(product => {
        this.form.set({ ...product });
      });
    }
  }

  update(field: string, value: any) {
    this.form.update(f => ({ ...f, [field]: value }));
  }

  get pageTitle() {
    return this.isEditing() ? 'Edit Product' : 'Add New Product';
  }

  get breadcrumb() {
    return this.isEditing() ? 'Edit' : 'Add';
  }

  save() {
    this.isSaving.set(true);
    const data = this.form();
    const id = this.productId();

    const action = id
      ? this.catalogService.updateProduct(id, data as any)
      : this.catalogService.createProduct(data as any);

    action.subscribe({
      next: () => {
        this.isSaving.set(false);
        this.router.navigate(['/inventory']);
      },
      error: (err) => {
        this.isSaving.set(false);
        console.error(err);
        this.toastService.showError('Failed to save product. Please check your inputs.');
      }
    });
  }

  get margin(): string {
    const f = this.form() as any;
    const sell = f.selling_price ?? 0;
    const cost = f.cost_price ?? 0;
    if (sell <= 0) return '–';
    return (((sell - cost) / sell) * 100).toFixed(1) + '%';
  }

  get grossProfit(): number {
    const f = this.form() as any;
    return (f.selling_price ?? 0) - (f.cost_price ?? 0);
  }

  cancel() {
    this.router.navigate(['/inventory']);
  }
}
