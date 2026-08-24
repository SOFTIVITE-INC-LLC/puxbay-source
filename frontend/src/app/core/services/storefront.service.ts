import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { StorefrontSettings, Product, Order } from '../models/models';

export interface Coupon {
  id: string;
  code: string;
  discount_type: string;
  value: number;
  min_purchase: number;
  is_active: boolean;
  valid_from: string;
  valid_to: string;
}

export interface CouponValidationResult {
  coupon: Coupon;
  discount_amount: number;
  new_total: number;
}

export interface ProductSearchResult {
  products: Product[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ProductDetailResult {
  product: Product;
  reviews: any[];
  avg_rating: number;
}

@Injectable({
  providedIn: 'root'
})
export class StorefrontService {
  private api = inject(ApiService);
  
  settings = signal<StorefrontSettings | null>(null);
  coupons = signal<Coupon[]>([]);
  loading = signal<boolean>(false);

  // --- Settings ---

  getSettings(): Observable<StorefrontSettings> {
    this.loading.set(true);
    return this.api.get<StorefrontSettings>('/storefront/settings').pipe(
      tap(res => {
        this.settings.set(res);
        this.loading.set(false);
      })
    );
  }

  updateSettings(settings: Partial<StorefrontSettings>): Observable<StorefrontSettings> {
    return this.api.put<StorefrontSettings>('/storefront/settings', settings).pipe(
      tap(res => this.settings.set(res))
    );
  }

  // --- Public Storefront API ---

  searchProducts(params?: any): Observable<ProductSearchResult> {
    return this.api.get<ProductSearchResult>('/storefront/products', { params });
  }

  getProduct(id: string): Observable<ProductDetailResult> {
    return this.api.get<ProductDetailResult>(`/storefront/products/${id}`);
  }

  trackOrder(orderNumber: string): Observable<Order> {
    return this.api.get<Order>('/storefront/track-order', { params: { order_number: orderNumber } });
  }

  submitReview(productId: string, customerId: string, rating: number, comment?: string): Observable<any> {
    return this.api.post<any>(`/storefront/products/${productId}/reviews`, { customer_id: customerId, rating, comment });
  }

  subscribeNewsletter(email: string): Observable<{message: string}> {
    return this.api.post<{message: string}>('/storefront/newsletter/subscribe', { email });
  }

  applyCoupon(code: string, cartTotal: number): Observable<CouponValidationResult> {
    return this.api.post<CouponValidationResult>('/storefront/coupon/apply', { code, cart_total: cartTotal });
  }

  // --- Admin Coupons ---

  listCoupons(): Observable<{coupons: Coupon[]}> {
    return this.api.get<{coupons: Coupon[]}>('/storefront/coupons').pipe(
      tap(res => this.coupons.set(res.coupons || []))
    );
  }

  createCoupon(coupon: Partial<Coupon>): Observable<Coupon> {
    return this.api.post<Coupon>('/storefront/coupons', coupon).pipe(
      tap(c => this.coupons.update(list => [c, ...list]))
    );
  }

  updateCoupon(id: string, coupon: Partial<Coupon>): Observable<Coupon> {
    return this.api.put<Coupon>(`/storefront/coupons/${id}`, coupon).pipe(
      tap(c => this.coupons.update(list => list.map(item => item.id === id ? c : item)))
    );
  }

  listProducts(): Observable<any[]> { return this.api.get<any[]>('/storefront/products'); }
  getCart(): Observable<any> { return this.api.get<any>('/storefront/cart'); }
  addToCart(data: any): Observable<any> { return this.api.post<any>('/storefront/cart/add', data); }
  updateCart(data: any): Observable<any> { return this.api.put<any>('/storefront/cart/update', data); }
  removeFromCart(id: string): Observable<any> { return this.api.delete<any>(`/storefront/cart/remove/${id}`); }
  checkout(data: any): Observable<any> { return this.api.post<any>('/storefront/checkout', data); }
  toggleWishlist(id: string): Observable<any> { return this.api.post<any>(`/storefront/wishlist/toggle/${id}`, {}); }
  removeCoupon(data: any): Observable<any> { return this.api.post<any>('/storefront/coupons/remove', data); }
}
