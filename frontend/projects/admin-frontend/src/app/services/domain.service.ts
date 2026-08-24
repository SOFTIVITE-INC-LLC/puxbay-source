import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Domain {
  id: number;
  tenant_id: string;
  domain: string;
  is_primary: boolean;
  is_verified: boolean;
  dns_checked_at: string;
  tenant?: {
    name: string;
    subdomain: string;
  };
}

export interface DomainResponse {
  data: Domain[];
  total: number;
}

export interface DomainDiagnostics {
  domain: string;
  cname: string;
  cname_error: string;
  ips: string[];
  ips_error: string;
  status: string;
}

@Injectable({
  providedIn: 'root'
})
export class DomainService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/domains';

  getDomains(limit: number = 100, offset: number = 0, search: string = '', status: string = 'all'): Observable<DomainResponse> {
    let url = `${this.apiUrl}?limit=${limit}&offset=${offset}`;
    if (search) url += `&search=${encodeURIComponent(search)}`;
    if (status && status !== 'all') url += `&status=${status}`;
    return this.http.get<DomainResponse>(url);
  }

  verifyDomain(id: number): Observable<any> {
    return this.http.post(`${this.apiUrl}/${id}/verify`, {});
  }

  bulkVerifyDomains(ids: number[]): Observable<any> {
    return this.http.post(`${this.apiUrl}/verify-bulk`, { ids });
  }

  getDomainDiagnostics(id: number): Observable<DomainDiagnostics> {
    return this.http.get<DomainDiagnostics>(`${this.apiUrl}/${id}/diagnostics`);
  }

  deleteDomain(id: number): Observable<any> {
    return this.http.delete(`${this.apiUrl}/${id}`);
  }
}
