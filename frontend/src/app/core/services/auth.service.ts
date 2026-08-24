import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap, switchMap } from 'rxjs';
import { TenantStore } from './tenant.store';

export interface User {
  id: string;
  user_id: string;
  tenant_id: string;
  branch_id?: string;
  email: string;
  username: string;
  role: string;
  permissions?: string[];
  first_name: string;
  last_name: string;
  subdomain?: string;
}

export interface AuthResponse {
  tokens: {
    access: string;
    refresh: string;
  };
  user: User;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private api = inject(ApiService);
  
  currentUser = signal<User | null>(null);
  isAuthenticated = signal<boolean>(false);
  loading = signal<boolean>(false);

  isInitialized = signal<boolean>(false);
  isImpersonating = signal<boolean>(false);

  constructor() {
    this.restoreSession();
  }

  private restoreSession() {
    if (typeof window === 'undefined' || typeof localStorage === 'undefined') {
      this.isInitialized.set(true);
      return;
    }
    // Check for impersonation token in URL
    const urlParams = new URLSearchParams(window.location.search);
    const supportToken = urlParams.get('support_token');
    
    if (supportToken) {
      localStorage.setItem('auth_token', supportToken);
      this.isImpersonating.set(true);
      // Clean URL
      window.history.replaceState({}, document.title, window.location.pathname);
    } else if (localStorage.getItem('is_impersonating') === 'true') {
      this.isImpersonating.set(true);
    }

    const token = localStorage.getItem('auth_token');
    if (token) {
      this.isAuthenticated.set(true);
      
      try {
        const userStr = localStorage.getItem('user');
        if (userStr) {
          this.currentUser.set(JSON.parse(userStr));
        }
      } catch (e) {
        console.error('Failed to parse user from local storage', e);
      }

      this.api.get<User & {subdomain?: string}>('/auth/user').subscribe({
        next: (res) => {
          if (res && res.id) {
            this.currentUser.set(res);
            localStorage.setItem('user', JSON.stringify(res));
            
            if (res.subdomain && (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')) {
              localStorage.setItem('dev_tenant', res.subdomain);
              this.tenantStore?.setSubdomain(res.subdomain);
            }
            
            if (supportToken) {
              localStorage.setItem('is_impersonating', 'true');
            }
            this.isAuthenticated.set(true);
          }
          this.isInitialized.set(true);
        },
        error: () => {
          this.logout();
          this.isInitialized.set(true);
        }
      });
    } else {
      this.isInitialized.set(true);
    }
  }

  login(credentials: any): Observable<AuthResponse> {
    this.loading.set(true);
    return this.api.post<AuthResponse>('/auth/login', credentials).pipe(
      tap(res => {
        if (res && res.tokens && res.tokens.access) {
          localStorage.setItem('auth_token', res.tokens.access);
          localStorage.setItem('refresh_token', res.tokens.refresh);
          localStorage.setItem('user', JSON.stringify(res.user));
          this.currentUser.set(res.user);
          
          const anyUser = res.user as any;
          if (anyUser.subdomain && (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')) {
            localStorage.setItem('dev_tenant', anyUser.subdomain);
            this.tenantStore?.setSubdomain(anyUser.subdomain);
          }
          
          this.isAuthenticated.set(true);
        }
        this.loading.set(false);
      })
    );
  }

  changeTemporaryPassword(username: string, temporary_password: string, new_password: string): Observable<AuthResponse> {
    this.loading.set(true);
    return this.api.post<AuthResponse>('/auth/change-temporary-password', { username, temporary_password, new_password }).pipe(
      tap(res => {
        if (res && res.tokens && res.tokens.access) {
          localStorage.setItem('auth_token', res.tokens.access);
          localStorage.setItem('refresh_token', res.tokens.refresh);
          localStorage.setItem('user', JSON.stringify(res.user));
          this.currentUser.set(res.user);
          
          const anyUser = res.user as any;
          if (anyUser.subdomain && (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')) {
            localStorage.setItem('dev_tenant', anyUser.subdomain);
            this.tenantStore?.setSubdomain(anyUser.subdomain);
          }
          
          this.isAuthenticated.set(true);
        }
        this.loading.set(false);
      })
    );
  }

  register(payload: any): Observable<any> {
    this.loading.set(true);
    return this.api.post<any>('/auth/register', payload).pipe(
      tap(() => {
        this.loading.set(false);
      })
    );
  }

  tenantStore = inject(TenantStore);

  registerAndLogin(payload: any): Observable<AuthResponse> {
    return this.register(payload).pipe(
      switchMap(() => {
        if (payload.subdomain && (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')) {
          localStorage.setItem('dev_tenant', payload.subdomain);
          this.tenantStore.setSubdomain(payload.subdomain);
        }
        return this.login({
          username: payload.email,
          password: payload.password
        });
      })
    );
  }

  logout() {
    this.isAuthenticated.set(false);
    this.currentUser.set(null);
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
      localStorage.removeItem('is_impersonating');
      localStorage.removeItem('puxbay_active_branch');
    }
    this.api.post('/auth/logout', {}).subscribe({
      next: () => {},
      error: () => {}
    });
    this.isImpersonating.set(false);
  }

  hasPermission(permissionCode: string | string[]): boolean {
    const user = this.currentUser();
    if (!user) return false;
    if (user.role === 'superadmin' || user.role === 'admin') return true;
    
    const permissions = Array.isArray(permissionCode) ? permissionCode : [permissionCode];
    
    // Return true if user has AT LEAST ONE of the required permissions
    return permissions.some(code => user.permissions?.includes(code) ?? false);
  }
}
