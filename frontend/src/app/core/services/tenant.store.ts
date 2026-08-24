import { Injectable, signal, computed } from '@angular/core';

export interface TenantState {
  tenantId: string | null;
  subdomain: string | null;
  loading: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class TenantStore {
  private state = signal<TenantState>({
    tenantId: null,
    subdomain: null,
    loading: false
  });

  // Selectors
  readonly tenantId = computed(() => this.state().tenantId);
  readonly subdomain = computed(() => this.state().subdomain);
  readonly isLoading = computed(() => this.state().loading);

  constructor() {
    this.initFromStorage();
  }

  private initFromStorage() {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('tenant_id');
      if (stored) {
        this.setTenantId(stored);
      }
    }
  }

  setTenantId(id: string) {
    this.state.update(s => ({ ...s, tenantId: id }));
    if (typeof window !== 'undefined') {
      localStorage.setItem('tenant_id', id);
    }
  }

  setSubdomain(subdomain: string) {
    this.state.update(s => ({ ...s, subdomain }));
  }

  setLoading(loading: boolean) {
    this.state.update(s => ({ ...s, loading }));
  }

  clear() {
    this.state.set({ tenantId: null, subdomain: null, loading: false });
    if (typeof window !== 'undefined') {
      localStorage.removeItem('tenant_id');
    }
  }
}
