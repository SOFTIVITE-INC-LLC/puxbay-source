import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';

export interface User {
  id: string;
  email: string;
  username: string;
  role?: string;
  is_superuser?: boolean;
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
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/auth'; // Using the global auth endpoint

  currentUser = signal<User | null>(null);
  isAuthenticated = signal<boolean>(false);
  loading = signal<boolean>(false);

  constructor() {
    this.restoreSession();
  }

  private restoreSession() {
    const token = localStorage.getItem('admin_auth_token');
    if (token) {
      this.isAuthenticated.set(true);
      try {
        const userStr = localStorage.getItem('admin_user');
        if (userStr) {
          this.currentUser.set(JSON.parse(userStr));
        }
      } catch (e) {
        console.error('Failed to parse user from local storage', e);
      }
    }
  }

  login(credentials: any): Observable<AuthResponse> {
    this.loading.set(true);
    return this.http.post<AuthResponse>(`${this.apiUrl}/login`, credentials).pipe(
      tap({
        next: (res) => {
          if (res && res.tokens && res.tokens.access) {
            localStorage.setItem('admin_auth_token', res.tokens.access);
            localStorage.setItem('admin_user', JSON.stringify(res.user));
            this.currentUser.set(res.user);
            this.isAuthenticated.set(true);
          }
          this.loading.set(false);
        },
        error: () => this.loading.set(false)
      })
    );
  }

  logout() {
    localStorage.removeItem('admin_auth_token');
    localStorage.removeItem('admin_user');
    this.currentUser.set(null);
    this.isAuthenticated.set(false);
  }

  getToken(): string | null {
    return localStorage.getItem('admin_auth_token');
  }
}
