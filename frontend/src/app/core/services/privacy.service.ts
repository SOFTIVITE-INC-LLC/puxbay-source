import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class PrivacyService {
  requests = signal<any[]>([]);
  getRequests(): Observable<any[]> { return this.api.get<any[]>('/privacy/requests').pipe(tap(res => this.requests.set(res))); }

  private api = inject(ApiService);
  
  loading = signal<boolean>(false);

  exportData(): Observable<{message: string; tenant_id: string}> {
    this.loading.set(true);
    return this.api.post<{message: string; tenant_id: string}>('/privacy/export', {}).pipe(
      tap(() => this.loading.set(false))
    );
  }

  deleteAccount(reason?: string): Observable<{message: string; tenant_id: string}> {
    this.loading.set(true);
    return this.api.post<{message: string; tenant_id: string}>('/privacy/delete-account', { reason }).pipe(
      tap(() => this.loading.set(false))
    );
  }

  anonymizeCustomer(id: string): Observable<any> { return this.api.post<any>(`/privacy/anonymize/${id}`, {}); }
}
