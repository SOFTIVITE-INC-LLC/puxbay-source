import { Injectable, inject, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { SessionService } from './session.service';
import { CartItem, CartPayload, CartResponse } from '../models/cart.model';
import { Product } from '../models/product.model';

@Injectable({
  providedIn: 'root'
})
export class CartService {
  private http = inject(HttpClient);
  private sessionService = inject(SessionService);
  private apiUrl = '/api/v1/storefront/cart';

  // Reactive state
  cartItems = signal<CartItem[]>([]);
  products = signal<Product[]>([]);

  cartCount = computed(() =>
    this.cartItems().reduce((sum, item) => sum + item.quantity, 0)
  );

  cartDetails = computed(() => {
    return this.cartItems().map(item => {
      const p = this.products().find(prod => prod.id === item.product_id);
      return { ...item, product: p };
    }).filter(item => item.product) as (CartItem & { product: Product })[];
  });

  cartTotal = computed(() => {
    return this.cartDetails().reduce(
      (sum, item) => sum + (item.quantity * item.product.selling_price), 0
    );
  });

  loadCart(): void {
    this.getCart().subscribe({
      next: (res) => {
        this.cartItems.set(res.cart || []);
        if (this.cartItems().length > 0) {
          this.loadProducts();
        }
      }
    });
  }

  private loadProducts(): void {
    this.http.get<any>('/api/v1/storefront/products?page_size=200').subscribe({
      next: (res) => this.products.set(res.products || [])
    });
  }

  getCart(): Observable<CartResponse> {
    const sessionId = this.sessionService.getSessionId();
    if (!sessionId) {
      return new Observable<CartResponse>(subscriber => {
        subscriber.next({ cart: [] });
        subscriber.complete();
      });
    }
    return this.http.get<CartResponse>(this.apiUrl, {
      headers: { 'X-Session-ID': sessionId }
    });
  }

  addToCart(payload: CartPayload): Observable<any> {
    const sessionId = this.sessionService.getOrCreateSessionId();
    return this.http.post(`${this.apiUrl}/add`, payload, {
      headers: { 'X-Session-ID': sessionId }
    }).pipe(
      tap(() => {
        // Optimistic local update
        this.cartItems.update(items => {
          const existing = items.find(i => i.product_id === payload.product_id);
          if (existing) {
            return items.map(i =>
              i.product_id === payload.product_id
                ? { ...i, quantity: i.quantity + payload.quantity }
                : i
            );
          }
          return [...items, { product_id: payload.product_id, quantity: payload.quantity }];
        });
        this.loadProducts();
      })
    );
  }

  updateQuantity(productId: string, quantity: number): Observable<any> {
    const sessionId = this.sessionService.getSessionId();
    if (!sessionId) return new Observable(s => { s.next({}); s.complete(); });

    // Optimistic local update
    this.cartItems.update(items =>
      items.map(i => i.product_id === productId ? { ...i, quantity } : i)
    );

    return this.http.put(`${this.apiUrl}/update`, { product_id: productId, quantity }, {
      headers: { 'X-Session-ID': sessionId }
    });
  }

  removeItem(productId: string): Observable<any> {
    const sessionId = this.sessionService.getSessionId();
    if (!sessionId) {
      return new Observable<any>(subscriber => {
        subscriber.next({});
        subscriber.complete();
      });
    }

    // Optimistic local update
    this.cartItems.update(items => items.filter(i => i.product_id !== productId));

    return this.http.delete(`${this.apiUrl}/remove/${productId}`, {
      headers: { 'X-Session-ID': sessionId }
    });
  }

  clearCart(): void {
    this.cartItems.set([]);
    this.sessionService.clearSession();
  }
}
