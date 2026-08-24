import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

export interface Setup2FAResult {
  secret: string;
  qr_code: string;
  message: string;
}

export interface AuditLog {
  id: string;
  action: string;
  model_name: string;
  object_id?: string;
  changes?: any;
  user_id: string;
  user?: { first_name: string; last_name: string; username: string };
  created_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class SecurityService {
  logs = signal<any[]>([]);

  private api = inject(ApiService);
  
  loading = signal<boolean>(false);
  auditLogs = signal<AuditLog[]>([]);
  totalLogs = signal<number>(0);

  setup2FA(): Observable<Setup2FAResult> {
    this.loading.set(true);
    return this.api.post<Setup2FAResult>('/security/2fa/setup', {}).pipe(
      tap(() => this.loading.set(false))
    );
  }

  verify2FA(code: string): Observable<{message: string}> {
    this.loading.set(true);
    return this.api.post<{message: string}>('/security/2fa/verify', { code }).pipe(
      tap(() => this.loading.set(false))
    );
  }

  getAuditLogs(page: number = 1, limit: number = 10): Observable<{tenant_id: string; logs: AuditLog[]; total: number}> {
    this.loading.set(true);
    const offset = (page - 1) * limit;
    return this.api.get<{tenant_id: string; logs: AuditLog[]; total: number}>(`/security/audit-logs?limit=${limit}&offset=${offset}`).pipe(
      tap(res => {
        this.auditLogs.set(res.logs || []);
        this.totalLogs.set(res.total || 0);
        this.loading.set(false);
      })
    );
  }

  disable2FA(data: any): Observable<any> { return this.api.post<any>('/security/2fa/disable', data); }
  backupDashboard(): Observable<any> { return this.api.get<any>('/security/backup'); }
  restoreBackup(data: any): Observable<any> { return this.api.post<any>('/security/backup/restore', data); }
}
