import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { Tenant } from '../models/tenant.models';

export interface Plan {
  id: string;
  name: string;
  price: number;
  features: string[];
}

export interface Broadcast {
  id: string;
  title: string;
  message: string;
  type: string;
  created_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class AdminService {
  private api = inject(ApiService);
  
  tenants = signal<Tenant[]>([]);
  plans = signal<Plan[]>([]);
  broadcasts = signal<Broadcast[]>([]);
  health = signal<any>({ status: 'unknown', latency: 0 });
  loading = signal<boolean>(false);

  getTenants(): Observable<{tenants: Tenant[]}> {
    this.loading.set(true);
    return this.api.get<{tenants: Tenant[]}>('/admin/tenants').pipe(
      tap(res => {
        this.tenants.set(res.tenants || []);
        this.loading.set(false);
      })
    );
  }

  getSystemHealth(): Observable<any> {
    return this.api.get<any>('/admin/health').pipe(
      tap(res => this.health.set(res))
    );
  }

  suspendTenant(id: string): Observable<any> {
    return this.api.post<any>(`/admin/tenants/${id}/suspend`, {}).pipe(
      tap(() => {
        this.tenants.update(list => list.map(t => t.id === id ? { ...t, status: 'suspended' } : t));
      })
    );
  }

  impersonateTenant(id: string): Observable<{token: string}> {
    return this.api.post<{token: string}>(`/admin/tenants/${id}/impersonate`, {});
  }

  getPlans(): Observable<Plan[]> {
    return this.api.get<Plan[]>('/admin/plans').pipe(
      tap(res => this.plans.set(res || []))
    );
  }

  createPlan(plan: Partial<Plan>): Observable<Plan> {
    return this.api.post<Plan>('/admin/plans', plan).pipe(
      tap(p => this.plans.update(list => [p, ...list]))
    );
  }

  getBroadcasts(): Observable<Broadcast[]> {
    return this.api.get<Broadcast[]>('/admin/broadcasts').pipe(
      tap(res => this.broadcasts.set(res || []))
    );
  }

  createBroadcast(b: Partial<Broadcast>): Observable<Broadcast> {
    return this.api.post<Broadcast>('/admin/broadcasts', b).pipe(
      tap(res => this.broadcasts.update(list => [res, ...list]))
    );
  }

  updateFeatureFlags(flags: any): Observable<any> {
    return this.api.post<any>('/admin/feature-flags', flags);
  }

  listDomains(): Observable<any[]> { return this.api.get<any[]>('/settings/domains'); }
  createDomain(data: any): Observable<any> { return this.api.post<any>('/settings/domains', data); }
  deleteDomain(id: string): Observable<any> { return this.api.delete<any>(`/settings/domains/${id}`); }
  verifyDomain(id: string): Observable<any> { return this.api.post<any>(`/settings/domains/${id}/verify`, {}); }
  setPrimaryDomain(id: string): Observable<any> { return this.api.post<any>(`/settings/domains/${id}/primary`, {}); }
}
