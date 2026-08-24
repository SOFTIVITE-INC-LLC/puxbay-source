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

// AuthResponse no longer contains tokens — set as HttpOnly cookies by backend.
export interface AuthResponse {
  user: User;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/auth';

  currentUser = signal<User | null>(null);
  isAuthenticated = signal<boolean>(false);
  loading = signal<boolean>(false);

  constructor() {
    this.restoreSession();
  }

  private restoreSession() {
    // Restore session by calling the backend with the HttpOnly pux_session cookie.
    // The browser sends it automatically due to withCredentials: true.
    this.http.get<User>(this.apiUrl + '/session', { withCredentials: true }).subscribe({
      next: (user) => {
        if (user && user.id) {
          this.currentUser.set(user);
          this.isAuthenticated.set(true);
        }
      },
      error: () => {
        this.isAuthenticated.set(false);
      }
    });
  }

  login(credentials: any): Observable<AuthResponse> {
    this.loading.set(true);
    return this.http.post<AuthResponse>(`${this.apiUrl}/login`, credentials, { withCredentials: true }).pipe(
      tap({
        next: (res) => {
          if (res && res.user) {
            // Tokens set as HttpOnly cookies by the backend
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
    this.http.post(`${this.apiUrl}/logout`, {}, { withCredentials: true }).subscribe({
      next: () => {},
      error: () => {}
    });
    this.currentUser.set(null);
    this.isAuthenticated.set(false);
  }

  // getToken() is no longer applicable — tokens are HttpOnly cookies.
  // Kept for compatibility but always returns null.
  getToken(): string | null {
    return null;
  }
}
