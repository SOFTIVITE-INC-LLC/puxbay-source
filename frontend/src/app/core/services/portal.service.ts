
import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap, of } from 'rxjs';

export interface PortalConfig {

  store_name: string;
  theme_color: string;
  logo_url: string;
  is_live: boolean;
  welcome_message?: string;
}

export interface PublicProduct {
  id: string;
  name: string;
  description: string;
  price: number;
  image_url: string;
  brand: string;
}

export interface Category {
  id: string;
  name: string;
}

@Injectable({
  providedIn: 'root'
})
export class PortalService {
  private api = inject(ApiService);
  
  config = signal<PortalConfig | null>(null);
  products = signal<PublicProduct[]>([]);
  categories = signal<Category[]>([]);
  availableBrands = signal<string[]>([]);
  loading = signal<boolean>(false);
  error = signal<string | null>(null);

  getConfig(): Observable<PortalConfig> {
    this.loading.set(true);
    return this.api.get<PortalConfig>('/public-portal/config').pipe(
      tap(res => {
        this.config.set(res);
        this.loading.set(false);
      })
    );
  }
  
  loadStore(domain: string) {
    this.loading.set(true);
    
    // Load storefront details
    this.api.get<PortalConfig>(`/api/v1/storefront/${domain}`).subscribe({
      next: (store) => {
        this.config.set(store);
      },
      error: () => {
        this.error.set('Store not found');
      }
    });

    // Load storefront products
    this.api.get<PublicProduct[]>(`/api/v1/storefront/${domain}/products`).subscribe({
      next: (products) => {
        this.products.set(products);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Failed to load products');
        this.loading.set(false);
      }
    });
  }
}
