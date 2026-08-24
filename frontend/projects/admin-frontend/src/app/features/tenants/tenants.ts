import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TenantService, Tenant } from '../../services/tenant.service';

@Component({
  selector: 'app-tenants',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './tenants.html',
})
export class TenantsComponent implements OnInit {
  private service = inject(TenantService);

  tenants = signal<Tenant[]>([]);
  stats = signal<any>(null);
  isLoading = signal(true);

  // Search & Filter
  searchQuery = signal('');
  statusFilter = signal('all');
  isSearching = signal(false);

  // Drawer state
  selectedTenant = signal<Tenant | null>(null);
  isDrawerOpen = signal(false);
  isLoadingDetail = signal(false);
  tenantNotes = signal('');
  isSavingNotes = signal(false);

  // Create Modal
  isCreateModalOpen = signal(false);
  createForm = signal({ name: '', subdomain: '' });
  isCreating = signal(false);

  ngOnInit() {
    this.loadTenants();
  }

  loadTenants() {
    this.isLoading.set(true);
    this.service.getTenants().subscribe({
      next: (res) => {
        this.tenants.set(res.data || []);
        this.stats.set(res.stats || null);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load tenants', err);
        this.isLoading.set(false);
      }
    });
  }

  onSearchChange() {
    const q = this.searchQuery();
    const s = this.statusFilter();
    if (!q && s === 'all') {
      this.loadTenants();
      return;
    }
    this.isSearching.set(true);
    this.service.searchTenants(q, s).subscribe({
      next: (res) => {
        this.tenants.set(res.data || []);
        this.isSearching.set(false);
      },
      error: () => this.isSearching.set(false)
    });
  }

  setFilter(status: string) {
    this.statusFilter.set(status);
    this.onSearchChange();
  }

  // Drawer
  openDrawer(tenant: Tenant) {
    this.selectedTenant.set(tenant);
    this.isDrawerOpen.set(true);
    this.tenantNotes.set(tenant.metadata?.admin_notes || '');
    this.isLoadingDetail.set(true);
    this.service.getTenantDetail(tenant.id).subscribe({
      next: (detail) => {
        this.selectedTenant.set(detail);
        this.tenantNotes.set(detail.metadata?.admin_notes || '');
        this.isLoadingDetail.set(false);
      },
      error: () => this.isLoadingDetail.set(false)
    });
  }

  closeDrawer() {
    this.isDrawerOpen.set(false);
    this.selectedTenant.set(null);
  }

  saveNotes() {
    const t = this.selectedTenant();
    if (!t) return;
    this.isSavingNotes.set(true);
    this.service.updateTenantNotes(t.id, this.tenantNotes()).subscribe({
      next: () => this.isSavingNotes.set(false),
      error: () => this.isSavingNotes.set(false)
    });
  }

  suspendTenant(id: string) {
    this.service.suspendTenant(id).subscribe({
      next: () => {
        this.loadTenants();
        this.closeDrawer();
      },
      error: (err) => console.error('Failed to suspend tenant', err)
    });
  }

  impersonateTenant(id: string) {
    this.service.impersonateTenant(id).subscribe({
      next: (res) => {
        // Store token and open tenant app in support mode
        const url = `http://app.puxbay.com?support_token=${res.token}`;
        window.open(url, '_blank');
      },
      error: (err) => console.error('Failed to impersonate tenant', err)
    });
  }

  // Create Modal
  openCreateModal() {
    this.createForm.set({ name: '', subdomain: '' });
    this.isCreateModalOpen.set(true);
  }

  closeCreateModal() {
    this.isCreateModalOpen.set(false);
  }

  createTenant() {
    const { name, subdomain } = this.createForm();
    if (!name || !subdomain) return;
    this.isCreating.set(true);
    this.service.createTenant(name, subdomain).subscribe({
      next: () => {
        this.isCreating.set(false);
        this.closeCreateModal();
        this.loadTenants();
      },
      error: (err) => {
        console.error('Failed to create tenant', err);
        this.isCreating.set(false);
      }
    });
  }

  // CSV Export
  exportCSV() {
    const data = this.tenants();
    const headers = ['ID', 'Name', 'Subdomain', 'Type', 'Status', 'Created On'];
    const rows = data.map(t => [
      t.id, t.name, t.subdomain, t.tenant_type, t.status, t.created_on
    ]);
    const csv = [headers, ...rows].map(r => r.join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `tenants_${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  getStatusClass(status: string): string {
    switch (status) {
      case 'active': return 'bg-emerald-100 text-emerald-700';
      case 'suspended': return 'bg-rose-100 text-rose-700';
      case 'trialing': return 'bg-blue-100 text-blue-700';
      case 'past_due': return 'bg-amber-100 text-amber-700';
      default: return 'bg-slate-100 text-slate-600';
    }
  }
}
