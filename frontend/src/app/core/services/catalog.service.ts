import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Product, Category } from '../models/product.models';
import { Observable, tap } from 'rxjs';
import { OfflineService } from './offline.service';

export interface ProductListResponse {
  data: Product[];
  total: number;
  page: number;
  limit: number;
}

export interface ProductCreateInput {
  name: string;
  description?: string;
  sku: string;
  barcode?: string;
  category_id?: string;
  cost_price: number;
  selling_price: number;
  wholesale_price?: number;
  track_inventory: boolean;
  current_stock: number;
  reorder_level: number;
  stock_unit: string;
  branch_id?: string;
  is_active?: boolean;
  is_online?: boolean;
  expiry_date?: string;
  manufacturing_date?: string;
  minimum_wholesale_quantity?: number;
  batch_number?: string;
  invoice_waybill_number?: string;
  country_of_origin?: string;
  manufacturer_name?: string;
  manufacturer_address?: string;
}

@Injectable({
  providedIn: 'root'
})
export class CatalogService {
  private api = inject(ApiService);
  private offlineService = inject(OfflineService);

  // State
  products = signal<Product[]>([]);
  categories = signal<Category[]>([]);
  loading = signal<boolean>(false);
  totalProducts = signal<number>(0);

  // --- Products ---

  getProducts(params?: any): Observable<ProductListResponse> {
    this.loading.set(true);
    return this.api.get<ProductListResponse>('/products', { params }).pipe(
      tap(res => {
        if (params?.page > 1) {
          this.products.update(prev => [...prev, ...(res.data || [])]);
        } else {
          this.products.set(res.data || []);
        }
        this.totalProducts.set(res.total || 0);
        this.loading.set(false);
      })
    );
  }

  getProduct(id: string): Observable<Product> {
    return this.api.get<Product>(`/products/${id}`);
  }

  createProduct(product: ProductCreateInput): Observable<Product> {
    return this.api.post<Product>('/products', product).pipe(
      tap(newProd => this.products.update(prods => [newProd, ...prods]))
    );
  }

  updateProduct(id: string, product: ProductCreateInput): Observable<Product> {
    return this.api.put<Product>(`/products/${id}`, product).pipe(
      tap(updatedProd => this.products.update(prods => prods.map(p => p.id === id ? updatedProd : p)))
    );
  }

  deleteProduct(id: string): Observable<any> {
    return this.api.delete(`/products/${id}`).pipe(
      tap(() => this.products.update(prods => prods.filter(p => p.id !== id)))
    );
  }

  importProducts(file: File): Observable<{message: string; count: number}> {
    const formData = new FormData();
    formData.append('file', file);
    return this.api.post<{message: string; count: number}>('/products/import', formData);
  }

  generateProductBarcode(id: string, format: string = 'base64'): Observable<{image: string, data?: string}> {
    return this.api.get<{image: string, data?: string}>(`/barcodes/products/${id}?format=${format}`);
  }

  generateProductQR(id: string, format: string = 'base64'): Observable<{image: string}> {
    return this.api.get<{image: string}>(`/barcodes/products/${id}/qr?format=${format}`);
  }

  adjustStock(product_id: string, quantity: number, reason: string): Observable<any> {
    return this.api.post('/inventory/receive', {
      product_id,
      quantity,
      reason
    }).pipe(
      tap(() => {
        this.products.update(prods => prods.map(p => {
          if (p.id === product_id) {
            return { ...p, current_stock: (p.current_stock || 0) + quantity };
          }
          return p;
        }));
      })
    );
  }

  getProductHistory(product_id: string): Observable<any> {
    return this.api.get(`/inventory/products/${product_id}/history`);
  }

  // --- Categories ---
  getCategories(): Observable<Category[]> {
    return this.api.get<Category[]>('/categories').pipe(
      tap(res => this.categories.set(res || []))
    );
  }

  createCategory(category: Partial<Category>): Observable<Category> {
    return this.api.post<Category>('/categories', category).pipe(
      tap(newCat => this.categories.update(cats => [newCat, ...cats]))
    );
  }

  // --- Batch & Expiry Tracking ---
  getBatches(productId: string): Observable<{ data: any[]; total: number }> {
    return this.api.get(`/inventory/products/${productId}/batches`);
  }

  createBatch(productId: string, batch: {
    batch_number: string;
    quantity: number;
    expiry_date?: string;
    manufacture_date?: string;
  }): Observable<any> {
    return this.api.post(`/inventory/products/${productId}/batches`, batch);
  }

  updateBatch(batchId: string, batch: {
    batch_number: string;
    quantity: number;
    expiry_date?: string;
    manufacture_date?: string;
  }): Observable<any> {
    return this.api.put(`/inventory/batches/${batchId}`, batch);
  }

  deleteBatch(batchId: string): Observable<any> {
    return this.api.delete(`/inventory/batches/${batchId}`);
  }

  getExpiringBatches(days = 30): Observable<{ data: any[]; total: number }> {
    return this.api.get(`/inventory/expiring-batches?days=${days}`);
  }

  // --- Product Gallery Images ---
  getProductImages(productId: string): Observable<{ data: { id: string; image_url: string; order: number }[] }> {
    return this.api.get(`/products/${productId}/images`);
  }

  addProductImage(productId: string, file: File): Observable<{ id: string; image_url: string; order: number }> {
    const formData = new FormData();
    formData.append('file', file);
    return this.api.post(`/products/${productId}/images`, formData);
  }

  deleteProductImage(productId: string, imageId: string): Observable<any> {
    return this.api.delete(`/products/${productId}/images/${imageId}`);
  }
}
