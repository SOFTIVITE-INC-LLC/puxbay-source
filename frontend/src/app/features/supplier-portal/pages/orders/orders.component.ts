import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { Router } from '@angular/router';
import { SupplierPortalService, SupplierProfile, PurchaseOrder } from '../../services/supplier-portal.service';
import { AppCurrencyPipe } from '../../../../core/pipes/app-currency.pipe';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-supplier-portal-orders',
  standalone: true,
  imports: [CommonModule, DatePipe, AppCurrencyPipe, FormsModule],
  templateUrl: './orders.component.html'
})
export class SupplierPortalOrdersComponent implements OnInit {
  portalService = inject(SupplierPortalService);
  private router = inject(Router);

  orders = signal<PurchaseOrder[]>([]);
  statusFilter = signal<string>('all');
  selectedOrder = signal<PurchaseOrder | null>(null);
  
  currentSupplier = signal<SupplierProfile | null>(null);

  ngOnInit() {
    this.portalService.currentSupplier$.subscribe(s => {
      if (s) {
        this.currentSupplier.set(s);
      } else {
        // Try fetching me if not loaded
        this.portalService.getMe().subscribe({
          error: () => this.router.navigate(['/supplier-portal/login'])
        });
      }
    });

    this.loadOrders();
  }

  loadOrders() {
    this.portalService.getPurchaseOrders().subscribe({
      next: (res) => this.orders.set(res || []),
      error: () => this.router.navigate(['/supplier-portal/login'])
    });
  }

  get filteredOrders() {
    const list = this.orders();
    const filter = this.statusFilter();
    if (filter === 'all') return list;
    return list.filter(o => o.status === filter);
  }

  statusClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s === 'received') return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400';
    if (s === 'partially_received') return 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400';
    if (s === 'cancelled') return 'bg-rose-100 text-rose-800 dark:bg-rose-900/30 dark:text-rose-400';
    if (s === 'issued') return 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400';
    return 'bg-zinc-100 text-zinc-800 dark:bg-zinc-800 dark:text-zinc-300';
  }

  viewOrder(order: any) {
    this.selectedOrder.set(order);
  }

  closeOrder() {
    this.selectedOrder.set(null);
  }

  logout() {
    this.portalService.logout();
    this.router.navigate(['/supplier-portal/login']);
  }
}
