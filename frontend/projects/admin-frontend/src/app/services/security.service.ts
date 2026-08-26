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

export const AVAILABLE_PERMISSIONS = [
  { id: 'dashboard:read', name: 'View Dashboard & Stats' },
  { id: 'tenants:read', name: 'View Tenants' },
  { id: 'tenants:write', name: 'Manage Tenants (Suspend/Impersonate)' },
  { id: 'domains:read', name: 'View Domains' },
  { id: 'domains:write', name: 'Manage Domains' },
  { id: 'billing:read', name: 'View Billing Data (Invoices, Renewals)' },
  { id: 'billing:write', name: 'Manage Billing & Subscriptions' },
  { id: 'pricing_plans:read', name: 'View Pricing Plans' },
  { id: 'pricing_plans:write', name: 'Manage Pricing Plans' },
  { id: 'promo_codes:read', name: 'View Promo Codes' },
  { id: 'promo_codes:write', name: 'Manage Promo Codes' },
  { id: 'content:read', name: 'View Content (Blog, FAQs, Legal)' },
  { id: 'content:write', name: 'Manage Content (Blog, FAQs, Legal)' },
  { id: 'referrals:read', name: 'View Referrals' },
  { id: 'broadcasts:read', name: 'View Broadcasts' },
  { id: 'broadcasts:write', name: 'Manage Broadcasts' },
  { id: 'apps:read', name: 'View App Marketplace' },
  { id: 'apps:write', name: 'Manage App Marketplace' },
  { id: 'webhooks:read', name: 'View Webhook Logs' },
  { id: 'webhooks:write', name: 'Manage Webhooks' },
  { id: 'backups:read', name: 'View System Backups' },
  { id: 'backups:write', name: 'Manage System Backups' },
  { id: 'api_keys:read', name: 'View API Keys' },
  { id: 'api_keys:write', name: 'Manage API Keys' },
  { id: 'security:read', name: 'View Audit & Telemetry Logs' },
  { id: 'security:write', name: 'Manage Security & Roles' },
  { id: 'admin_users:read', name: 'View Admin Users' },
  { id: 'admin_users:write', name: 'Manage Admin Users (Add/Delete)' },
  { id: 'settings:write', name: 'Manage Settings' }
];

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

  // Admin Users
  getAdminUsers(): Observable<{ data: any[] }> {
    return this.http.get<{ data: any[] }>(`${this.base}/users`);
  }

  createAdminUser(user: any): Observable<any> {
    return this.http.post<any>(`${this.base}/users`, user);
  }

  updateAdminUserRole(id: string, roleData: { admin_role_id: string | null, is_superuser: boolean, permissions?: string }): Observable<any> {
    return this.http.put(`${this.base}/users/${id}/role`, roleData);
  }

  deleteAdminUser(id: string): Observable<any> {
    return this.http.delete(`${this.base}/users/${id}`);
  }
}
