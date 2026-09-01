import { ToastService } from '../../../core/services/toast';
import { Component, inject, OnInit, OnDestroy, signal } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { ImageUrlPipe } from '../../../core/pipes/image-url.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, ActivatedRoute, RouterModule } from '@angular/router';
import { CatalogService } from '../../../core/services/catalog.service';
import { Product } from '../../../core/models/models';
import { Html5Qrcode, Html5QrcodeSupportedFormats } from 'html5-qrcode';
import { SettingsService } from '../../../core/services/settings.service';
import { AlertService } from '../../../core/services/alert.service';

const MAX_IMAGE_BYTES = 2 * 1024 * 1024; // 2 MB

@Component({
  selector: 'app-product-form',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe, ImageUrlPipe],
  templateUrl: './product-form.html',
})
export class ProductForm implements OnInit, OnDestroy {
  toastService = inject(ToastService);
  catalogService = inject(CatalogService);
  settingsService = inject(SettingsService);
  alertService = inject(AlertService);
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  readonly Math = Math;

  isEditing = signal(false);
  isSaving = signal(false);
  isUploadingImage = signal(false);
  isUploadingGallery = signal(false);
  productId = signal<string | null>(null);
  activeTab: 'basic' | 'pricing' | 'inventory' | 'details' = 'basic';

  // Gallery images (up to 5 extra images)
  galleryImages = signal<{ id: string; image_url: string; order: number }[]>([]);
  readonly MAX_GALLERY = 5;

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
      // Load gallery images
      this.catalogService.getProductImages(id).subscribe(res => {
        this.galleryImages.set(res.data || []);
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

  /** Returns the current product image URL for preview */
  get currentImage(): string | null {
    const f = this.form() as any;
    return f.image_url || f.image || null;
  }

  onImageFileSelected(event: any) {
    const file: File = event.target?.files?.[0];
    if (!file) return;

    // Client-side 2 MB guard for instant feedback
    if (file.size > MAX_IMAGE_BYTES) {
      this.alertService.alert(
        `Image is too large (${(file.size / 1024 / 1024).toFixed(2)} MB). Maximum allowed size is 2 MB.`,
        'File Too Large',
        'danger'
      );
      event.target.value = '';
      return;
    }

    this.isUploadingImage.set(true);
    this.settingsService.uploadImage(file, 'product').subscribe({
      next: (res) => {
        this.isUploadingImage.set(false);
        this.form.update(f => ({ ...f, image_url: res.url, image: res.url }));
        this.toastService.showSuccess('Product image uploaded!');
      },
      error: (err) => {
        this.isUploadingImage.set(false);
        const msg = err?.error?.error || 'Failed to upload image';
        this.alertService.alert(msg, 'Upload Error', 'danger');
      }
    });
  }

  onGalleryFileSelected(event: any) {
    const files: FileList = event.target?.files;
    if (!files || files.length === 0) return;
    const id = this.productId();
    if (!id) {
      this.toastService.showError('Save the product first before adding gallery images.');
      event.target.value = '';
      return;
    }
    const remaining = this.MAX_GALLERY - this.galleryImages().length;
    const toUpload = Array.from(files).slice(0, remaining);

    toUpload.forEach(file => {
      if (file.size > MAX_IMAGE_BYTES) {
        this.alertService.alert(
          `"${file.name}" is too large (${(file.size / 1024 / 1024).toFixed(2)} MB). Max 2 MB.`,
          'File Too Large', 'danger'
        );
        return;
      }
      this.isUploadingGallery.set(true);
      this.catalogService.addProductImage(id, file).subscribe({
        next: (img) => {
          this.galleryImages.update(imgs => [...imgs, img]);
          this.isUploadingGallery.set(false);
          this.toastService.showSuccess('Gallery image added!');
        },
        error: (err) => {
          this.isUploadingGallery.set(false);
          const msg = err?.error?.error || 'Failed to upload gallery image';
          this.alertService.alert(msg, 'Upload Error', 'danger');
        }
      });
    });
    event.target.value = '';
  }

  removeGalleryImage(imageId: string) {
    const id = this.productId();
    if (!id) return;
    this.catalogService.deleteProductImage(id, imageId).subscribe({
      next: () => {
        this.galleryImages.update(imgs => imgs.filter(i => i.id !== imageId));
        this.toastService.showInfo('Gallery image removed.');
      },
      error: () => this.toastService.showError('Failed to remove image.')
    });
  }

  removeProductImage() {
    this.form.update(f => ({ ...f, image_url: null, image: null }));
    this.toastService.showInfo('Image removed.');
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
