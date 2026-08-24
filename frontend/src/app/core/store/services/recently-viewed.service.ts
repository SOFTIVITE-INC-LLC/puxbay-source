import { Injectable, inject, signal } from '@angular/core';
import { Product } from '../models/product.model';
import { ProductService } from './product.service';
import { forkJoin, map, of, catchError } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class RecentlyViewedService {
  private readonly STORAGE_KEY = 'storefront_recently_viewed';
  private readonly MAX_ITEMS = 8;
  
  private productService = inject(ProductService);
  
  recentlyViewedProducts = signal<Product[]>([]);

  addProduct(productId: string) {
    if (typeof localStorage === 'undefined') return;
    
    let viewed = this.getViewedIds();
    
    // Remove if already exists to move to front
    viewed = viewed.filter(id => id !== productId);
    
    // Add to front
    viewed.unshift(productId);
    
    // Cap at MAX_ITEMS
    if (viewed.length > this.MAX_ITEMS) {
      viewed = viewed.slice(0, this.MAX_ITEMS);
    }
    
    localStorage.setItem(this.STORAGE_KEY, JSON.stringify(viewed));
  }

  loadRecentlyViewed() {
    const ids = this.getViewedIds();
    if (ids.length === 0) {
      this.recentlyViewedProducts.set([]);
      return;
    }

    // Fetch products by id. A simple way is to use getProduct for each, 
    // or if the backend supports a comma separated list. 
    // We'll use forkJoin with getProduct for simplicity since it's cached or fast enough for a small list.
    const requests = ids.map(id => this.productService.getProduct(id).pipe(
      map(res => res.product),
      catchError(() => of(null))
    ));

    forkJoin(requests).subscribe(products => {
      // Filter out nulls and set
      const validProducts = products.filter((p): p is Product => p !== null);
      this.recentlyViewedProducts.set(validProducts);
    });
  }

  private getViewedIds(): string[] {
    if (typeof localStorage === 'undefined') return [];
    try {
      return JSON.parse(localStorage.getItem(this.STORAGE_KEY) || '[]');
    } catch {
      return [];
    }
  }
}
