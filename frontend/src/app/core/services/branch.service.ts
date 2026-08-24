import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, map, tap } from 'rxjs';
import { Branch } from '../models/branch.models';

export interface BranchCreateInput {
  name: string;
  address?: string;
  phone?: string;
  primary_color?: string;
  currency_symbol?: string;
  currency_code?: string;
  branch_type?: string;
}

export interface BranchUpdateInput {
  name?: string;
  address?: string;
  phone?: string;
  primary_color?: string;
  currency_symbol?: string;
  currency_code?: string;
  receipt_header?: string;
  receipt_footer?: string;
  low_stock_threshold?: number;
}

@Injectable({
  providedIn: 'root'
})
export class BranchService {
  private api = inject(ApiService);
  private readonly baseUrl = '/branches';
  
  private readonly STORAGE_KEY = 'puxbay_active_branch';

  branches = signal<Branch[]>([]);
  loading = signal<boolean>(false);

  // Restored from localStorage on init so refresh doesn't lose branch context
  activeBranch = signal<Branch | null>(this.restoreActiveBranch());

  private restoreActiveBranch(): Branch | null {
    try {
      const raw = localStorage.getItem(this.STORAGE_KEY);
      return raw ? (JSON.parse(raw) as Branch) : null;
    } catch {
      return null;
    }
  }

  setActiveBranch(branch: Branch | null) {
    this.activeBranch.set(branch);
    if (branch) {
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(branch));
    } else {
      localStorage.removeItem(this.STORAGE_KEY);
    }
  }

  getBranches(params?: any): Observable<Branch[]> {
    this.loading.set(true);
    return this.api.get<{ data: Branch[]; total: number }>(this.baseUrl, { params }).pipe(
      map(res => res?.data || []),
      tap(branches => {
        this.branches.set(branches);
        this.loading.set(false);
      })
    );
  }

  getBranch(id: string): Observable<Branch> {
    return this.api.get<Branch>(`${this.baseUrl}/${id}`);
  }

  createBranch(input: BranchCreateInput): Observable<Branch> {
    return this.api.post<Branch>(this.baseUrl, input).pipe(
      tap(b => this.branches.update(list => [...list, b]))
    );
  }

  updateBranch(id: string, input: BranchUpdateInput): Observable<Branch> {
    return this.api.put<Branch>(`${this.baseUrl}/${id}`, input).pipe(
      tap(b => this.branches.update(list => list.map(item => item.id === b.id ? b : item)))
    );
  }

  deleteBranch(id: string): Observable<void> {
    return this.api.delete<void>(`${this.baseUrl}/${id}`).pipe(
      tap(() => this.branches.update(list => list.filter(item => item.id !== id)))
    );
  }

  getNetworkMetrics(): Observable<any> {
    return this.api.get<any>(`${this.baseUrl}/network/metrics`);
  }

  getBranchMetrics(id: string): Observable<any> {
    return this.api.get<any>(`${this.baseUrl}/${id}/metrics`);
  }
}
