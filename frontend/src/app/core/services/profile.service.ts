import { Injectable, signal, inject } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

export interface UserProfile {
  id?: string;
  user_id?: string;
  first_name: string;
  last_name: string;
  email: string;
  username: string;
  role?: string;
  branch_id?: string;
  is_2fa_enabled?: boolean;
  is_email_verified?: boolean;
}

export interface UpdateMeInput {
  first_name?: string;
  last_name?: string;
  current_password?: string;
  new_password?: string;
}

@Injectable({
  providedIn: 'root'
})
export class ProfileService {
  private api = inject(ApiService);
  
  loading = signal<boolean>(false);
  saving = signal<boolean>(false);
  profile = signal<UserProfile | null>(null);
  error = signal<string | null>(null);

  getProfile(): Observable<any> {
    this.loading.set(true);
    this.error.set(null);
    return this.api.get<any>('/auth/user').pipe(
      tap({
        next: (res) => {
          this.profile.set(res as UserProfile);
          this.loading.set(false);
        },
        error: () => this.loading.set(false)
      })
    );
  }

  updateMe(input: UpdateMeInput): Observable<any> {
    this.saving.set(true);
    this.error.set(null);
    return this.api.put<any>('/auth/me', input).pipe(
      tap({
        next: (res) => {
          // Merge updated fields into current profile
          const current = this.profile();
          if (current) {
            this.profile.set({
              ...current,
              first_name: res.first_name ?? current.first_name,
              last_name: res.last_name ?? current.last_name,
            });
          }
          this.saving.set(false);
        },
        error: (err) => {
          this.error.set(err?.error?.error || 'Failed to update profile');
          this.saving.set(false);
        }
      })
    );
  }

  setPosPin(pin: string): Observable<any> {
    this.saving.set(true);
    this.error.set(null);
    return this.api.put<any>('/profiles/pos-pin', { pos_pin: pin }).pipe(
      tap({
        next: () => this.saving.set(false),
        error: (err) => {
          this.error.set(err?.error?.error || 'Failed to set POS PIN');
          this.saving.set(false);
        }
      })
    );
  }

  setup2FA(): Observable<{ secret: string; qr_code: string; message: string }> {
    this.saving.set(true);
    this.error.set(null);
    return this.api.post<{ secret: string; qr_code: string; message: string }>('/security/2fa/setup', {}).pipe(
      tap({
        next: () => this.saving.set(false),
        error: (err) => {
          this.error.set(err?.error?.error || 'Failed to initialize 2FA setup');
          this.saving.set(false);
        }
      })
    );
  }

  verify2FA(code: string): Observable<any> {
    this.saving.set(true);
    this.error.set(null);
    return this.api.post<any>('/security/2fa/verify', { code }).pipe(
      tap({
        next: () => {
          const current = this.profile();
          if (current) {
            this.profile.set({ ...current, is_2fa_enabled: true });
          }
          this.saving.set(false);
        },
        error: (err) => {
          this.error.set(err?.error?.error || 'Invalid 2FA verification code');
          this.saving.set(false);
        }
      })
    );
  }

  disable2FA(): Observable<any> {
    this.saving.set(true);
    this.error.set(null);
    return this.api.post<any>('/security/2fa/disable', {}).pipe(
      tap({
        next: () => {
          const current = this.profile();
          if (current) {
            this.profile.set({ ...current, is_2fa_enabled: false });
          }
          this.saving.set(false);
        },
        error: (err) => {
          this.error.set(err?.error?.error || 'Failed to disable 2FA');
          this.saving.set(false);
        }
      })
    );
  }
}
