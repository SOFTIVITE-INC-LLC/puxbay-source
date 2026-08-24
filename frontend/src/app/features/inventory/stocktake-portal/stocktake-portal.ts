import { Component, OnInit, OnDestroy, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { ApiService } from '../../../core/services/api.service';
import { ToastService } from '../../../core/services/toast';
import { Subject, Subscription } from 'rxjs';
import { debounceTime } from 'rxjs/operators';

@Component({
  selector: 'app-stocktake-portal',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './stocktake-portal.html',
  styles: [`
    :host { display: block; }
  `]
})
export class StocktakePortalComponent implements OnInit, OnDestroy {
  private route = inject(ActivatedRoute);
  private api = inject(ApiService);
  private toast = inject(ToastService);

  token = signal<string | null>(null);
  session = signal<any>(null);

  query = signal<string>('');
  results = signal<any[]>([]);
  loading = signal<boolean>(true);
  updating = signal<{ [key: string]: boolean }>({});

  private countUpdate$ = new Subject<{ product: any, count: number }>();
  private sub?: Subscription;
  private pollInterval: any;

  ngOnInit() {
    this.token.set(this.route.snapshot.paramMap.get('token'));
    if (this.token()) {
      this.loadSession();
      this.searchProduct();
      
      // Poll every 5 seconds to keep counts in sync with other users
      this.pollInterval = setInterval(() => {
        this.searchProduct(true);
      }, 5000);
    }

    this.sub = this.countUpdate$.pipe(
      debounceTime(800)
    ).subscribe(({ product, count }) => {
      this.executeSetExactCount(product, count);
    });
  }

  trackById(index: number, item: any): string {
    return item.id;
  }

  ngOnDestroy() {
    this.sub?.unsubscribe();
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
    }
  }

  loadSession() {
    this.api.get(`/public/stocktake/${this.token()}`).subscribe({
      next: (res) => {
        this.session.set(res);
        this.loading.set(false);
      },
      error: (err) => {
        console.error('Failed to load session:', err);
        this.toast.showError('Failed to load stocktake session');
        this.loading.set(false);
      }
    });
  }

  searchProduct(silent = false) {
    if (!this.token()) return;
    this.api.get(`/public/stocktake/${this.token()}/scan?q=${encodeURIComponent(this.query())}`).subscribe({
      next: (res: any) => {
        const newResults = res.results || [];
        const currentUpdating = this.updating();
        
        // Preserve current count for items that are currently being updated by the user
        const mergedResults = newResults.map((p: any) => {
          if (currentUpdating[p.id]) {
            const existing = this.results().find(r => r.id === p.id);
            if (existing) {
              return { ...p, current_count: existing.current_count };
            }
          }
          return p;
        });
        
        this.results.set(mergedResults);
      },
      error: (err) => {
        console.error('Search error:', err);
      }
    });
  }

  updateQuery(newVal: string) {
    this.query.set(newVal);
  }

  updateCount(product: any, delta: number) {
    if (!this.token()) return;
    const newCount = Math.max(0, product.current_count + delta);
    product.current_count = newCount;
    this.executeSetExactCount(product, newCount);
  }

  setExactCount(product: any, count: number) {
    if (!this.token() || count === null || count === undefined) return;
    product.current_count = Math.max(0, count);
    this.countUpdate$.next({ product, count: product.current_count });
  }

  private executeSetExactCount(product: any, count: number) {
    if (!this.token()) return;
    this.updating.update(u => ({ ...u, [product.id]: true }));

    this.api.post(`/public/stocktake/${this.token()}/update`, {
      product_id: product.id,
      quantity: count,
      mode: 'set'
    }).subscribe({
      next: (res: any) => {
        this.updating.update(u => ({ ...u, [product.id]: false }));
        this.toast.showSuccess(`Updated ${product.name}`);
        product.current_count = res.new_count ?? count;
      },
      error: (err) => {
        this.updating.update(u => ({ ...u, [product.id]: false }));
        console.error('Update error:', err);
        this.toast.showError('Failed to update count');
      }
    });
  }
}
