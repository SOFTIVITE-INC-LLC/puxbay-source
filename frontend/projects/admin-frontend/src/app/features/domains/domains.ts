import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DomainService, Domain, DomainDiagnostics } from '../../services/domain.service';
import { AlertService } from '../../services/alert.service';

@Component({
  selector: 'app-domains',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './domains.html',
})
export class DomainsComponent implements OnInit {
  private service = inject(DomainService);
  private alert = inject(AlertService);

  domains = signal<Domain[]>([]);
  total = signal(0);
  isLoading = signal(true);

  // Pagination & Filtering state
  page = signal(1);
  limit = 20;
  search = signal('');
  status = signal('all');

  // Multi-select state
  selectedIds = signal<Set<number>>(new Set());

  // Diagnostics state
  diagnostics = signal<{ [id: number]: { isLoading: boolean, data?: DomainDiagnostics } }>({});

  ngOnInit() {
    this.loadDomains();
  }

  loadDomains() {
    this.isLoading.set(true);
    const offset = (this.page() - 1) * this.limit;
    
    this.service.getDomains(this.limit, offset, this.search(), this.status()).subscribe({
      next: (res) => {
        this.domains.set(res.data || []);
        this.total.set(res.total || 0);
        this.selectedIds.set(new Set()); // Reset selection on load
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load domains', err);
        this.isLoading.set(false);
      }
    });
  }

  onSearch(term: string) {
    this.search.set(term);
    this.page.set(1);
    this.loadDomains();
  }

  onStatusChange(status: string) {
    this.status.set(status);
    this.page.set(1);
    this.loadDomains();
  }

  changePage(newPage: number) {
    if (newPage < 1 || newPage > this.totalPages()) return;
    this.page.set(newPage);
    this.loadDomains();
  }

  totalPages(): number {
    return Math.ceil(this.total() / this.limit);
  }

  toggleSelection(id: number) {
    const current = new Set(this.selectedIds());
    if (current.has(id)) {
      current.delete(id);
    } else {
      current.add(id);
    }
    this.selectedIds.set(current);
  }

  toggleAll(event: any) {
    if (event.target.checked) {
      this.selectedIds.set(new Set(this.domains().map(d => d.id)));
    } else {
      this.selectedIds.set(new Set());
    }
  }

  async bulkVerify() {
    const ids = Array.from(this.selectedIds());
    if (ids.length === 0) return;

    const confirmed = await this.alert.confirm({
      title: 'Bulk DNS Verification',
      message: `Run DNS verification for ${ids.length} selected domain${ids.length > 1 ? 's' : ''}? This may take a few seconds.`,
      confirmText: 'Run Verification',
      cancelText: 'Cancel',
      type: 'info'
    });
    if (confirmed) {
      this.service.bulkVerifyDomains(ids).subscribe({
        next: () => {
          this.alert.success(`Bulk verification complete for ${ids.length} domain${ids.length > 1 ? 's' : ''}.`, 'Verification Done');
          this.loadDomains();
        },
        error: () => this.alert.error('Failed to run bulk verification.')
      });
    }
  }

  loadDiagnostics(id: number) {
    // Toggle off if already loaded
    if (this.diagnostics()[id]) {
      const current = { ...this.diagnostics() };
      delete current[id];
      this.diagnostics.set(current);
      return;
    }

    // Load
    this.diagnostics.update(curr => ({ ...curr, [id]: { isLoading: true } }));
    this.service.getDomainDiagnostics(id).subscribe({
      next: (data) => {
        this.diagnostics.update(curr => ({ ...curr, [id]: { isLoading: false, data } }));
      },
      error: (err) => {
        console.error('Failed to load diagnostics', err);
        this.diagnostics.update(curr => ({ ...curr, [id]: { isLoading: false } }));
      }
    });
  }

  async deleteDomain(id: number) {
    const confirmed = await this.alert.confirm({
      title: 'Remove Custom Domain',
      message: 'Are you sure you want to remove this custom domain? The tenant will need to re-add it.',
      confirmText: 'Remove Domain',
      cancelText: 'Cancel',
      type: 'danger'
    });
    if (confirmed) {
      this.service.deleteDomain(id).subscribe({
        next: () => { this.alert.success('Domain removed successfully.'); this.loadDomains(); },
        error: () => this.alert.error('Failed to remove domain.')
      });
    }
  }
}
