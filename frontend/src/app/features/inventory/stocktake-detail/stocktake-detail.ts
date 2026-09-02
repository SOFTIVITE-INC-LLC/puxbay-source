import { Component, OnInit, inject, OnDestroy, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { ApiService } from '../../../core/services/api.service';
import { ToastService } from '../../../core/services/toast';
import { AlertService } from '../../../core/services/alert.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

export interface StocktakeCategoryStat {
  categoryName: string;
  totalItems: number;
  matchedItems: number;
  discrepancyItems: number;
  expectedUnits: number;
  countedUnits: number;
  varianceUnits: number;
  varianceValue: number;
  accuracyRate: number;
}

@Component({
  selector: 'app-stocktake-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe],
  templateUrl: './stocktake-detail.html',
  styles: [`
    :host { display: block; }
    @media print {
      .no-print { display: none !important; }
      .print-only { display: block !important; }
      body { background: white !important; color: black !important; }
    }
  `]
})
export class StocktakeDetailComponent implements OnInit, OnDestroy {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private api = inject(ApiService);
  private toast = inject(ToastService);
  private alertService = inject(AlertService);

  id = signal<string | null>(null);
  session = signal<any>(null);
  loading = signal<boolean>(true);
  portalUrl = signal<string>('');
  qrUrl = signal<string>('');
  
  // Tab navigation
  activeTab = signal<'live' | 'analysis'>('live');

  // Filtering & Sorting
  searchQuery = signal<string>('');
  filterStatus = signal<'all' | 'discrepancy' | 'shortage' | 'surplus' | 'matched'>('all');
  sortBy = signal<'variance-desc' | 'variance-asc' | 'name' | 'value-impact'>('variance-desc');

  // Real-time metadata
  lastSynced = signal<Date>(new Date());
  isRefreshing = signal<boolean>(false);
  editingEntryId = signal<string | null>(null);
  editCountValue = signal<number>(0);

  private pollInterval: any;

  // Computed Key Performance Metrics
  totalItems = computed(() => this.session()?.entries?.length || 0);

  totalExpectedUnits = computed(() => {
    const entries = this.session()?.entries || [];
    return entries.reduce((acc: number, e: any) => acc + (Number(e.expected_stock) || 0), 0);
  });

  totalCountedUnits = computed(() => {
    const entries = this.session()?.entries || [];
    return entries.reduce((acc: number, e: any) => acc + (Number(e.actual_stock) || 0), 0);
  });

  netVarianceUnits = computed(() => {
    return this.totalCountedUnits() - this.totalExpectedUnits();
  });

  matchedCount = computed(() => {
    const entries = this.session()?.entries || [];
    return entries.filter((e: any) => Number(e.difference) === 0).length;
  });

  shortageCount = computed(() => {
    const entries = this.session()?.entries || [];
    return entries.filter((e: any) => Number(e.difference) < 0).length;
  });

  surplusCount = computed(() => {
    const entries = this.session()?.entries || [];
    return entries.filter((e: any) => Number(e.difference) > 0).length;
  });

  discrepancyCount = computed(() => {
    return this.shortageCount() + this.surplusCount();
  });

  accuracyRate = computed(() => {
    const total = this.totalItems();
    if (total === 0) return 100;
    return Math.round((this.matchedCount() / total) * 1000) / 10;
  });

  // Financial Valuation Metrics
  totalShrinkageValue = computed(() => {
    const entries = this.session()?.entries || [];
    return entries.reduce((acc: number, e: any) => {
      const diff = Number(e.difference) || 0;
      if (diff < 0) {
        const cost = Number(e.product?.cost_price) || Number(e.product?.selling_price) || 0;
        return acc + (Math.abs(diff) * cost);
      }
      return acc;
    }, 0);
  });

  totalSurplusValue = computed(() => {
    const entries = this.session()?.entries || [];
    return entries.reduce((acc: number, e: any) => {
      const diff = Number(e.difference) || 0;
      if (diff > 0) {
        const cost = Number(e.product?.cost_price) || Number(e.product?.selling_price) || 0;
        return acc + (diff * cost);
      }
      return acc;
    }, 0);
  });

  netVarianceValue = computed(() => {
    return this.totalSurplusValue() - this.totalShrinkageValue();
  });

  // Category Breakdown Stats
  categoryStats = computed<StocktakeCategoryStat[]>(() => {
    const entries = this.session()?.entries || [];
    const catMap = new Map<string, {
      total: number;
      matched: number;
      discrepancy: number;
      expected: number;
      counted: number;
      diff: number;
      valDiff: number;
    }>();

    for (const e of entries) {
      const catName = e.product?.category?.name || 'Uncategorized';
      if (!catMap.has(catName)) {
        catMap.set(catName, { total: 0, matched: 0, discrepancy: 0, expected: 0, counted: 0, diff: 0, valDiff: 0 });
      }
      const data = catMap.get(catName)!;
      const expected = Number(e.expected_stock) || 0;
      const counted = Number(e.actual_stock) || 0;
      const diff = Number(e.difference) || 0;
      const cost = Number(e.product?.cost_price) || Number(e.product?.selling_price) || 0;

      data.total += 1;
      data.expected += expected;
      data.counted += counted;
      data.diff += diff;
      data.valDiff += (diff * cost);

      if (diff === 0) {
        data.matched += 1;
      } else {
        data.discrepancy += 1;
      }
    }

    const result: StocktakeCategoryStat[] = [];
    catMap.forEach((val, key) => {
      result.push({
        categoryName: key,
        totalItems: val.total,
        matchedItems: val.matched,
        discrepancyItems: val.discrepancy,
        expectedUnits: val.expected,
        countedUnits: val.counted,
        varianceUnits: val.diff,
        varianceValue: val.valDiff,
        accuracyRate: val.total > 0 ? Math.round((val.matched / val.total) * 1000) / 10 : 100
      });
    });

    return result.sort((a, b) => Math.abs(b.varianceValue) - Math.abs(a.varianceValue));
  });

  // Top Discrepancies
  topShortages = computed(() => {
    const entries = [...(this.session()?.entries || [])];
    return entries
      .filter((e: any) => Number(e.difference) < 0)
      .sort((a: any, b: any) => {
        const costA = Number(a.product?.cost_price) || Number(a.product?.selling_price) || 0;
        const costB = Number(b.product?.cost_price) || Number(b.product?.selling_price) || 0;
        return (Math.abs(b.difference) * costB) - (Math.abs(a.difference) * costA);
      })
      .slice(0, 5);
  });

  topSurpluses = computed(() => {
    const entries = [...(this.session()?.entries || [])];
    return entries
      .filter((e: any) => Number(e.difference) > 0)
      .sort((a: any, b: any) => {
        const costA = Number(a.product?.cost_price) || Number(a.product?.selling_price) || 0;
        const costB = Number(b.product?.cost_price) || Number(b.product?.selling_price) || 0;
        return (b.difference * costB) - (a.difference * costA);
      })
      .slice(0, 5);
  });

  // Filtered and Sorted Entries
  filteredEntries = computed(() => {
    const entries = this.session()?.entries || [];
    const query = this.searchQuery().toLowerCase().trim();
    const filter = this.filterStatus();
    const sort = this.sortBy();

    let list = entries.filter((e: any) => {
      // Search match
      if (query) {
        const name = (e.product?.name || '').toLowerCase();
        const sku = (e.product?.sku || '').toLowerCase();
        const barcode = (e.product?.barcode || '').toLowerCase();
        const category = (e.product?.category?.name || '').toLowerCase();
        if (!name.includes(query) && !sku.includes(query) && !barcode.includes(query) && !category.includes(query)) {
          return false;
        }
      }

      // Filter match
      const diff = Number(e.difference) || 0;
      if (filter === 'discrepancy') return diff !== 0;
      if (filter === 'shortage') return diff < 0;
      if (filter === 'surplus') return diff > 0;
      if (filter === 'matched') return diff === 0;

      return true;
    });

    // Sorting
    return list.sort((a: any, b: any) => {
      const diffA = Number(a.difference) || 0;
      const diffB = Number(b.difference) || 0;
      const costA = Number(a.product?.cost_price) || Number(a.product?.selling_price) || 0;
      const costB = Number(b.product?.cost_price) || Number(b.product?.selling_price) || 0;

      if (sort === 'variance-desc') {
        return Math.abs(diffB) - Math.abs(diffA);
      } else if (sort === 'variance-asc') {
        return Math.abs(diffA) - Math.abs(diffB);
      } else if (sort === 'value-impact') {
        return (Math.abs(diffB) * costB) - (Math.abs(diffA) * costA);
      } else if (sort === 'name') {
        return (a.product?.name || '').localeCompare(b.product?.name || '');
      }
      return 0;
    });
  });

  ngOnInit() {
    this.id.set(this.route.snapshot.paramMap.get('id'));
    const tabParam = this.route.snapshot.queryParamMap.get('tab');
    if (tabParam === 'analysis' || tabParam === 'audit' || this.router.url.includes('/analysis')) {
      this.activeTab.set('analysis');
    }

    if (this.id()) {
      this.loadSession();
      // Poll every 5 seconds for live analytics updates
      this.pollInterval = setInterval(() => {
        if (this.session()?.status !== 'completed') {
          this.loadSession(true);
        }
      }, 5000);
    }
  }

  ngOnDestroy() {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
    }
  }

  loadSession(silent = false) {
    if (!silent) this.loading.set(true);
    this.api.get(`/inventory/stocktakes/${this.id()}`).subscribe({
      next: (res: any) => {
        try {
          this.session.set(res?.data || res);
          this.lastSynced.set(new Date());
          
          // Build portal URL based on current host
          const baseUrl = window.location.origin;
          const token = this.session()?.access_token || '';
          this.portalUrl.set(`${baseUrl}/stocktake/portal/${token}`);
          this.qrUrl.set(`https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=${encodeURIComponent(this.portalUrl())}`);
        } catch (e) {
          console.error('Error processing session:', e);
        } finally {
          if (!silent) {
            this.loading.set(false);
          }
          this.isRefreshing.set(false);
        }
      },
      error: (err) => {
        console.error('Failed to load session:', err);
        if (!silent) this.toast.showError('Failed to load stocktake session: ' + (err.message || 'Unknown error'));
        if (!silent) {
          this.loading.set(false);
        }
        this.isRefreshing.set(false);
        
        if (!silent && this.pollInterval) {
          clearInterval(this.pollInterval);
        }
      }
    });
  }

  refreshNow() {
    this.isRefreshing.set(true);
    this.loadSession(false);
  }

  startEditCount(entry: any) {
    if (this.session()?.status === 'completed') return;
    this.editingEntryId.set(entry.id);
    this.editCountValue.set(Number(entry.actual_stock) || 0);
  }

  cancelEditCount() {
    this.editingEntryId.set(null);
  }

  saveCount(entry: any) {
    const token = this.session()?.access_token;
    if (!token) return;

    const newCount = Math.max(0, Number(this.editCountValue()) || 0);
    this.api.post(`/public/stocktake/${token}/update`, {
      product_id: entry.product_id,
      quantity: newCount,
      mode: 'set'
    }).subscribe({
      next: () => {
        this.toast.showSuccess(`Updated ${entry.product?.name || 'item'} to ${newCount}`);
        this.editingEntryId.set(null);
        this.loadSession(true);
      },
      error: (err) => {
        console.error('Failed to update count:', err);
        this.toast.showError('Failed to update count');
      }
    });
  }

  async finalizeSession() {
    const disc = this.discrepancyCount();
    const netVal = this.netVarianceValue();
    const formattedVal = netVal >= 0 ? `+${netVal.toFixed(2)}` : netVal.toFixed(2);

    const message = `Are you sure you want to finalize this stocktake session?\n\n` +
      `• Total Discrepancies: ${disc} products\n` +
      `• Net Inventory Valuation Impact: ${formattedVal}\n` +
      `\nAll variances will be permanently reconciled and committed as stock movements.`;

    if (await this.alertService.confirm(message, 'Finalize & Reconcile Stocktake')) {
      this.loading.set(true);
      this.api.post(`/inventory/stocktakes/${this.id()}/finalize`, {}).subscribe({
        next: () => {
          this.loading.set(false);
          this.toast.showSuccess('Stocktake finalized and inventory reconciled successfully!');
          this.loadSession(false);
        },
        error: (err) => {
          this.loading.set(false);
          console.error('Finalize error:', err);
          this.toast.showError('Failed to finalize stocktake: ' + (err.error?.error || err.message));
        }
      });
    }
  }

  exportCSV() {
    const entries = this.session()?.entries || [];
    if (entries.length === 0) {
      this.toast.showError('No entries available to export');
      return;
    }

    const headers = [
      'Product Name',
      'SKU',
      'Barcode',
      'Category',
      'Expected Stock',
      'Actual Counted',
      'Variance Qty',
      'Unit Cost Price',
      'Unit Selling Price',
      'Total Valuation Impact',
      'Status'
    ];

    const rows = entries.map((e: any) => {
      const diff = Number(e.difference) || 0;
      const cost = Number(e.product?.cost_price) || 0;
      const price = Number(e.product?.selling_price) || 0;
      const valueImpact = diff * (cost || price);
      let status = 'Accurate';
      if (diff < 0) status = 'Shortage';
      else if (diff > 0) status = 'Surplus';

      return [
        `"${(e.product?.name || '').replace(/"/g, '""')}"`,
        `"${e.product?.sku || ''}"`,
        `"${e.product?.barcode || ''}"`,
        `"${(e.product?.category?.name || 'Uncategorized').replace(/"/g, '""')}"`,
        e.expected_stock,
        e.actual_stock,
        diff,
        cost.toFixed(2),
        price.toFixed(2),
        valueImpact.toFixed(2),
        status
      ];
    });

    const csvContent = [headers.join(','), ...rows.map((r: any[]) => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    const sessionName = (this.session()?.name || 'stocktake').replace(/[^a-zA-Z0-9_-]/g, '_');
    a.download = `Stocktake_Analysis_${sessionName}_${new Date().toISOString().slice(0,10)}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    this.toast.showSuccess('Audit CSV report downloaded successfully');
  }

  printReport() {
    window.print();
  }

  copyPortalUrl() {
    if (this.portalUrl()) {
      navigator.clipboard.writeText(this.portalUrl()).then(() => {
        this.toast.showSuccess('Portal URL copied to clipboard');
      }).catch(err => {
        console.error('Could not copy text: ', err);
        this.toast.showError('Failed to copy URL');
      });
    }
  }

  openPortalUrl() {
    if (this.portalUrl()) {
      window.open(this.portalUrl(), '_blank');
    }
  }

  goBack() {
    this.router.navigate(['/inventory']);
  }
}
