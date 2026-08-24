import { Injectable, inject, signal } from '@angular/core';
import { Product } from '../models/product.model';
import { ProductService } from './product.service';
import { forkJoin, map, of, catchError } from 'rxjs';
import { ToastService } from './toast.service';

import { StorefrontAuthService } from './storefront-auth.service';
import { HttpClient } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class WishlistService {
  private readonly STORAGE_KEY = 'storefront_wishlist';
  
  private productService = inject(ProductService);
  private toastService = inject(ToastService);
  private authService = inject(StorefrontAuthService);
  private http = inject(HttpClient);
  
  wishlistIds = signal<string[]>(this.getStoredIds());
  wishlistProducts = signal<Product[]>([]);
  isLoading = signal(false);

  toggleWishlist(productId: string) {
    let ids = this.wishlistIds();
    const isSaved = ids.includes(productId);

    if (isSaved) {
      ids = ids.filter(id => id !== productId);
      this.toastService.show('Removed from wishlist', 'info');
    } else {
      ids = [...ids, productId];
      this.toastService.show('Added to wishlist', 'success');
    }

    this.wishlistIds.set(ids);
    
    // Sync to backend if logged in
    const token = this.authService.getToken();
    if (token) {
      this.http.post(`/api/v1/storefront/me/wishlist/${productId}`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      }).subscribe();
    } else {
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem(this.STORAGE_KEY, JSON.stringify(ids));
      }
    }
  }

  // Load wishlist from backend if logged in
  loadUserWishlist() {
    const token = this.authService.getToken();
    if (token) {
      this.http.get<{product_ids: string[]}>(`/api/v1/storefront/me/wishlist`, {
        headers: { Authorization: `Bearer ${token}` }
      }).subscribe(res => {
        this.wishlistIds.set(res.product_ids || []);
      });
    }
  }

  isSaved(productId: string): boolean {
    return this.wishlistIds().includes(productId);
  }

  loadWishlistProducts() {
    const ids = this.wishlistIds();
    if (ids.length === 0) {
      this.wishlistProducts.set([]);
      return;
    }

    this.isLoading.set(true);
    const requests = ids.map(id => this.productService.getProduct(id).pipe(
      map(res => res.product),
      catchError(() => of(null))
    ));

    forkJoin(requests).subscribe(products => {
      const validProducts = products.filter((p): p is Product => p !== null);
      this.wishlistProducts.set(validProducts);
      this.isLoading.set(false);
    });
  }

  private getStoredIds(): string[] {
    if (typeof localStorage === 'undefined') return [];
    try {
      return JSON.parse(localStorage.getItem(this.STORAGE_KEY) || '[]');
    } catch {
      return [];
    }
  }
}
