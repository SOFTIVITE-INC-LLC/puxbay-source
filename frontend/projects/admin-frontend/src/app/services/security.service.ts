import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface AdminRole {
  id: string;
  name: string;
  permissions: any;
  created_at: string;
}

export interface IPAllowlist {
  id: string;
  ip_address: string;
  description: string;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class SecurityService {
  private http = inject(HttpClient);
  private base = '/api/v1/admin';

  getRoles(): Observable<{ data: AdminRole[] }> {
    return this.http.get<{ data: AdminRole[] }>(`${this.base}/admin-roles`);
  }

  createRole(role: any): Observable<AdminRole> {
    return this.http.post<AdminRole>(`${this.base}/admin-roles`, role);
  }

  updateRolePermissions(id: string, permissions: any): Observable<any> {
    return this.http.put(`${this.base}/admin-roles/${id}`, { permissions });
  }

  getIPAllowlist(): Observable<{ data: IPAllowlist[] }> {
    return this.http.get<{ data: IPAllowlist[] }>(`${this.base}/ip-allowlist`);
  }

  addIP(ip: any): Observable<IPAllowlist> {
    return this.http.post<IPAllowlist>(`${this.base}/ip-allowlist`, ip);
  }

  removeIP(id: string): Observable<any> {
    return this.http.delete(`${this.base}/ip-allowlist/${id}`);
  }

  // API Keys
  getAPIKeys(): Observable<{ data: any[] }> {
    return this.http.get<{ data: any[] }>(`${this.base}/api-keys`);
  }

  createAPIKey(name: string): Observable<any> {
    return this.http.post<any>(`${this.base}/api-keys`, { name });
  }

  revokeAPIKey(id: string): Observable<any> {
    return this.http.delete(`${this.base}/api-keys/${id}`);
  }
}
