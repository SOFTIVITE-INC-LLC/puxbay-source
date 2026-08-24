import { Component, HostListener, inject, OnInit, signal, ChangeDetectionStrategy } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ApiService } from '../../../core/services/api.service';
import { OrderService } from '../../../core/services/order.service';
import { Order } from '../../../core/models/order.models';
import { Subject } from 'rxjs';
import { debounceTime } from 'rxjs/operators';

@Component({
  selector: 'app-orders',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './orders.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .dark .glass-panel {
      background: rgba(0, 0, 0, 0.2);
    }
  `,
})
export class Orders implements OnInit {
  orderService = inject(OrderService);
  router = inject(Router);
  api = inject(ApiService);
  readonly Math = Math;
  
  searchQuery = signal('');
  searchSubject = new Subject<string>();
  selectedStatus = signal('all');
  selectedOrderType = signal('');  // '' = all, 'kiosk', 'online', etc.

  // Pagination State
  currentPage = signal(1);
  limit = signal(50);
  get totalPages() {
    return Math.ceil(this.orderService.totalOrders() / this.limit()) || 1;
  }
  exporting = signal(false);

  ngOnInit() {
    this.searchSubject.pipe(debounceTime(300)).subscribe(q => {
      this.searchQuery.set(q);
      this.onFilterChange();
    });
    this.loadOrders();
  }

  private _lastScrollTime = 0;

  @HostListener('window:scroll')
  onWindowScroll() {
    const scrolled = window.scrollY + window.innerHeight;
    const total = document.documentElement.scrollHeight;
    if (total - scrolled <= 150) {
      if (!this.orderService.loading() && this.currentPage() < this.totalPages) {
        const now = Date.now();
        if (now - this._lastScrollTime > 500) {
          this._lastScrollTime = now;
          this.nextPage();
        }
      }
    }
  }

  loadOrders() {
    const params: any = {
      page: this.currentPage(),
      limit: this.limit()
    };
    if (this.selectedOrderType()) params['order_type'] = this.selectedOrderType();
    if (this.selectedStatus() !== 'all') params['status'] = this.selectedStatus();
    if (this.searchQuery()) params['q'] = this.searchQuery();
    
    this.orderService.getOrders(params).subscribe();
  }

  // Pagination Methods
  nextPage() {
    if (this.currentPage() < this.totalPages && !this.orderService.loading()) {
      this.currentPage.update(p => p + 1);
      this.loadOrders();
    }
  }

  prevPage() {
    if (this.currentPage() > 1) {
      this.currentPage.update(p => p - 1);
      this.loadOrders();
    }
  }

  onFilterChange() {
    this.currentPage.set(1);
    this.loadOrders();
  }

  // --- KPIs ---
  get totalOrders() { return this.orderService.orders().length; }
  
  get totalRevenue() { 
    return this.orderService.orders()
      .filter(o => o.status !== 'cancelled' && o.status !== 'voided')
      .reduce((sum, o) => sum + (o.total || 0), 0); 
  }
  
  get averageOrderValue() { 
    const validOrders = this.orderService.orders().filter(o => o.status !== 'cancelled' && o.status !== 'voided');
    if (validOrders.length === 0) return 0;
    return this.totalRevenue / validOrders.length;
  }
  
  get cancelledOrders() { 
    return this.orderService.orders().filter(o => o.status === 'cancelled' || o.status === 'voided').length; 
  }

  get filteredOrders() {
    const q = this.searchQuery().toLowerCase();
    return this.orderService.orders().filter(o => {
      if (!q) return true;
      return o.order_number?.toLowerCase().includes(q) 
        || o.status?.toLowerCase().includes(q)
        || (o as any).customer?.name?.toLowerCase().includes(q);
    });
  }

  setTab(tab: string) {
    this.selectedStatus.set('all');
    if (tab === 'kiosk') {
      this.selectedOrderType.set('kiosk');
    } else if (tab === 'online') {
      this.selectedOrderType.set('online');
    } else {
      this.selectedOrderType.set('');
    }
    this.onFilterChange();
  }

  getCustomerName(order: any): string {
    return order?.customer?.name || (order?.customer_id ? 'Customer' : 'Walk-in Customer');
  }

  // --- Order Actions ---
  openOrderDetails(orderId: string) {
    this.router.navigate(['/orders', orderId]);
  }
}
