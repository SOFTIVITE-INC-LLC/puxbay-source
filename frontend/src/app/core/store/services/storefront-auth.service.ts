import { Injectable, inject, signal, Injector } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { Router } from '@angular/router';
import { ToastService } from './toast.service';
import { WishlistService } from './wishlist.service';

export interface StorefrontCustomer {
  id: string;
  name: string;
  email: string;
  phone?: string;
  address?: string;
  is_registered: boolean;
}

export interface AuthResponse {
  token: string;
  customer: StorefrontCustomer;
}

@Injectable({
  providedIn: 'root'
})
export class StorefrontAuthService {
  private http = inject(HttpClient);
  private router = inject(Router);
  private toast = inject(ToastService);
  private injector = inject(Injector);
  
  private apiUrl = '/api/v1/storefront';
  
  currentUser = signal<StorefrontCustomer | null>(null);

  constructor() {
    const token = this.getToken();
    if (token) {
      this.fetchMe().subscribe({
        next: () => {
          setTimeout(() => {
            const wishlistService = this.injector.get(WishlistService);
            wishlistService.loadUserWishlist();
          }, 0);
        },
        error: () => this.logout()
      });
    }
  }

  getToken(): string | null {
    if (typeof localStorage === 'undefined') return null;
    return localStorage.getItem('storefront_token');
  }

  setToken(token: string) {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('storefront_token', token);
    }
  }

  register(data: any): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/auth/register`, data).pipe(
      tap(res => {
        this.setToken(res.token);
        this.currentUser.set(res.customer);
        const wishlistService = this.injector.get(WishlistService);
        wishlistService.loadUserWishlist();
      })
    );
  }

  login(data: any): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.apiUrl}/auth/login`, data).pipe(
      tap(res => {
        this.setToken(res.token);
        this.currentUser.set(res.customer);
        const wishlistService = this.injector.get(WishlistService);
        wishlistService.loadUserWishlist();
      })
    );
  }

  logout() {
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem('storefront_token');
    }
    this.currentUser.set(null);
    const wishlistService = this.injector.get(WishlistService);
    wishlistService.wishlistIds.set([]);
    wishlistService.wishlistProducts.set([]);
    this.router.navigate(['/store/login']);
  }

  fetchMe(): Observable<StorefrontCustomer> {
    return this.http.get<StorefrontCustomer>(`${this.apiUrl}/me`, {
      headers: { Authorization: `Bearer ${this.getToken()}` }
    }).pipe(
      tap(customer => this.currentUser.set(customer))
    );
  }

  updateMe(data: any): Observable<StorefrontCustomer> {
    return this.http.put<StorefrontCustomer>(`${this.apiUrl}/me`, data, {
      headers: { Authorization: `Bearer ${this.getToken()}` }
    }).pipe(
      tap(customer => {
        this.currentUser.set(customer);
        this.toast.show('Profile updated successfully!', 'success');
      })
    );
  }
  
  getOrders(): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/me/orders`, {
      headers: { Authorization: `Bearer ${this.getToken()}` }
    });
  }
}
