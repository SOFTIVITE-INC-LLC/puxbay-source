import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule, ActivatedRoute, Router } from '@angular/router';
import { CatalogService } from '../../../core/services/catalog.service';
import { OrderService } from '../../../core/services/order.service';
import { Product } from '../../../core/models/product.models';
import { Order } from '../../../core/models/order.models';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-product-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe],
  templateUrl: './product-detail.html'
})
export class ProductDetail implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private catalogService = inject(CatalogService);
  private orderService = inject(OrderService);

  product = signal<Product | null>(null);
  orders = signal<Order[]>([]);
  movements = signal<any[]>([]);
  batches = signal<any[]>([]);
  loading = signal(true);
  activeTab = signal<'overview' | 'orders' | 'movements' | 'batches'>('overview');

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) { this.router.navigate(['/inventory']); return; }
    this.loadAll(id);
  }

  loadAll(id: string) {
    this.loading.set(true);
    this.catalogService.getProduct(id).subscribe({
      next: (prod) => {
        this.product.set(prod);
        // Load order history, stock movements and batches in parallel
        this.orderService.getOrders({ product_id: id, limit: 50, page: 1 }).subscribe({
          next: (res) => this.orders.set(res.data || [])
        });
        this.catalogService.getProductHistory(id).subscribe({
          next: (res) => this.movements.set(res?.history || [])
        });
        this.catalogService.getBatches(id).subscribe({
          next: (res) => {
            this.batches.set(res.data || []);
            this.loading.set(false);
          },
          error: () => this.loading.set(false)
        });
      },
      error: () => this.router.navigate(['/inventory'])
    });
  }

  setTab(tab: 'overview' | 'orders' | 'movements' | 'batches') {
    this.activeTab.set(tab);
  }

  getStatusClass(status: string | undefined): string {
    switch (status?.toLowerCase()) {
      case 'completed': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400';
      case 'pending': return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400';
      case 'void': case 'cancelled': return 'bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400';
      default: return 'bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-400';
    }
  }

  getMovementClass(type: string | undefined): string {
    switch (type?.toLowerCase()) {
      case 'sale': return 'text-rose-500';
      case 'receive': case 'restock': return 'text-emerald-500';
      case 'adjustment': return 'text-amber-500';
      default: return 'text-zinc-500';
    }
  }

  getMovementSign(type: string | undefined): string {
    const outTypes = ['sale', 'return', 'waste', 'transfer_out'];
    return outTypes.includes(type?.toLowerCase() || '') ? '-' : '+';
  }

  getDaysUntilExpiry(dateStr: string | undefined): number {
    if (!dateStr) return 999;
    const today = new Date(); today.setHours(0,0,0,0);
    const exp = new Date(dateStr); exp.setHours(0,0,0,0);
    return Math.ceil((exp.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));
  }

  getStockStatusClass(product: Product | null): string {
    if (!product) return '';
    if (!product.track_inventory) return 'text-zinc-500';
    const stock = product.current_stock || 0;
    const reorder = product.reorder_level || 0;
    if (stock <= 0) return 'text-rose-600 dark:text-rose-400';
    if (stock <= reorder) return 'text-amber-600 dark:text-amber-400';
    return 'text-emerald-600 dark:text-emerald-400';
  }

  getMargin(product: Product | null): string {
    if (!product || !product.selling_price || !product.cost_price) return '—';
    const margin = ((product.selling_price - product.cost_price) / product.selling_price) * 100;
    return margin.toFixed(1) + '%';
  }
}
