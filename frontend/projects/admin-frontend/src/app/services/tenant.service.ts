import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Tenant {
  id: string;
  name: string;
  subdomain: string;
  tenant_type: string;
  is_sandbox: boolean;
  status: string;
  created_on: string;
  referral_code?: string;
  metadata?: any;
  subscription?: {
    id: string;
    status: string;
    plan?: { name: string; price: number; };
    next_billing_date?: string;
  };
  domains?: { domain: string; is_verified: boolean; is_primary: boolean; }[];
}

export interface TenantResponse {
  data: Tenant[];
  stats: {
    total: number;
    active: number;
    suspended: number;
    trialing: number;
    revenue: number;
  };
}

export interface SearchResponse {
  data: Tenant[];
  total: number;
}

@Injectable({
  providedIn: 'root'
})
export class TenantService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/tenants';

  getTenants(): Observable<TenantResponse> {
    return this.http.get<TenantResponse>(this.apiUrl);
  }

  searchTenants(search: string, status: string): Observable<SearchResponse> {
    const params: any = {};
    if (search) params['search'] = search;
    if (status && status !== 'all') params['status'] = status;
    return this.http.get<SearchResponse>(`${this.apiUrl}/search`, { params });
  }

  getTenantDetail(id: string): Observable<Tenant> {
    return this.http.get<Tenant>(`${this.apiUrl}/${id}`);
  }

  createTenant(name: string, subdomain: string): Observable<Tenant> {
    return this.http.post<Tenant>(this.apiUrl, { name, subdomain });
  }

  suspendTenant(id: string): Observable<any> {
    return this.http.post(`${this.apiUrl}/${id}/suspend`, {});
  }

  impersonateTenant(id: string): Observable<{ token: string }> {
    return this.http.post<{ token: string }>(`${this.apiUrl}/${id}/impersonate`, {});
  }

  updateTenantNotes(id: string, notes: string): Observable<any> {
    return this.http.put(`${this.apiUrl}/${id}/notes`, { notes });
  }
}
