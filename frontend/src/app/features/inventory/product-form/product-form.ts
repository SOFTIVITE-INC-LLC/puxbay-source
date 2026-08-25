import { ToastService } from '../../../core/services/toast';
import { Component, inject, OnInit, OnDestroy, signal } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, ActivatedRoute, RouterModule } from '@angular/router';
import { CatalogService } from '../../../core/services/catalog.service';
import { Product } from '../../../core/models/models';
import { Html5Qrcode, Html5QrcodeSupportedFormats } from 'html5-qrcode';

@Component({
  selector: 'app-product-form',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe],
  templateUrl: './product-form.html',
})
export class ProductForm implements OnInit, OnDestroy {
  toastService = inject(ToastService);
  catalogService = inject(CatalogService);
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  readonly Math = Math;

  isEditing = signal(false);
  isSaving = signal(false);
  productId = signal<string | null>(null);
  activeTab: 'basic' | 'pricing' | 'inventory' | 'details' = 'basic';

  // ── Barcode Camera Scanner ────────────────────────────────────
  showBarcodeScanner = signal(false);
  barcodeError = signal<string | null>(null);
  private html5QrCode: Html5Qrcode | null = null;

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

  ngOnDestroy() {
    this.stopBarcodeScanner();
  }

  openBarcodeScanner() {
    this.showBarcodeScanner.set(true);
    this.barcodeError.set(null);
    setTimeout(() => this.startBarcodeScanner(), 150);
  }

  private startBarcodeScanner() {
    try {
      this.html5QrCode = new Html5Qrcode('pf-barcode-scanner-reader', {
        formatsToSupport: [
          Html5QrcodeSupportedFormats.QR_CODE,
          Html5QrcodeSupportedFormats.EAN_13,
          Html5QrcodeSupportedFormats.EAN_8,
          Html5QrcodeSupportedFormats.UPC_A,
          Html5QrcodeSupportedFormats.UPC_E,
          Html5QrcodeSupportedFormats.CODE_128,
          Html5QrcodeSupportedFormats.CODE_39,
          Html5QrcodeSupportedFormats.CODE_93,
          Html5QrcodeSupportedFormats.ITF,
          Html5QrcodeSupportedFormats.DATA_MATRIX,
        ],
        verbose: false,
      });

      this.html5QrCode
        .start(
          { facingMode: 'environment' },
          { fps: 10, qrbox: { width: 260, height: 180 } },
          (decodedText) => this.onBarcodeScanSuccess(decodedText),
          () => {}
        )
        .catch((err) => {
          this.barcodeError.set('Camera access denied or not available.');
          console.error('Barcode scan error:', err);
        });
    } catch (e) {
      this.barcodeError.set('Could not start the camera scanner.');
    }
  }

  private onBarcodeScanSuccess(decodedText: string) {
    // Populate the barcode field
    this.update('barcode', decodedText);
    this.stopBarcodeScanner();
    setTimeout(() => this.showBarcodeScanner.set(false), 600);
  }

  closeBarcodeScanner() {
    this.stopBarcodeScanner();
    this.showBarcodeScanner.set(false);
    this.barcodeError.set(null);
  }

  private stopBarcodeScanner() {
    if (this.html5QrCode) {
      this.html5QrCode.stop().catch(() => {}).finally(() => { this.html5QrCode = null; });
    }
  }
}
