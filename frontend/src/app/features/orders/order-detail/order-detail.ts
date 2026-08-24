import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { OrderService } from '../../../core/services/order.service';
import { Order } from '../../../core/models/order.models';
import { ReceiptComponent } from '../../../shared/components/receipt/receipt.component';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

import { PosOverrideModalComponent } from '../../../shared/components/pos-override-modal/pos-override-modal';

@Component({
  selector: 'app-order-detail',
  standalone: true,
  imports: [CommonModule, ReceiptComponent, AppCurrencyPipe, PosOverrideModalComponent],
  templateUrl: './order-detail.html'
})
export class OrderDetail implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  public orderService = inject(OrderService);

  order = signal<Order | null>(null);
  loading = signal(true);
  voidingOrder = signal(false);
  completingOrder = signal(false);
  showVoidConfirm = signal(false);
  showOverrideModal = signal(false);

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.loadOrder(id);
    } else {
      this.goBack();
    }
  }

  loadOrder(id: string) {
    this.loading.set(true);
    this.orderService.getOrder(id).subscribe({
      next: (order) => {
        this.order.set(order);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
        this.goBack();
      }
    });
  }

  goBack() {
    this.router.navigate(['/orders']);
  }

  getCustomerName(order: Order | null): string {
    if (!order) return 'Unknown';
    if (order.customer) return `${order.customer.first_name || ''} ${order.customer.last_name || ''}`.trim() || 'Walk-in Customer';
    if (order.customer_id) return 'Walk-in Customer';
    return 'Walk-in Customer';
  }

  getDeliveryAddress(order: Order | null): string | null {
    if (!order || !order.notes) return null;
    const match = order.notes.match(/Delivery Address:\s*(.*)/i);
    return match ? match[1].trim() : null;
  }

  openGoogleMaps(address: string) {
    const query = encodeURIComponent(address);
    window.open(`https://www.google.com/maps/search/?api=1&query=${query}`, '_blank');
  }

  confirmVoidOrder() { this.showVoidConfirm.set(true); }
  cancelVoid() { this.showVoidConfirm.set(false); }

  voidOrder(overridePin?: string) {
    const o = this.order();
    if (!o?.id) return;
    this.voidingOrder.set(true);
    this.showVoidConfirm.set(false);
    this.orderService.voidOrder(o.id, overridePin).subscribe({
      next: () => {
        this.voidingOrder.set(false);
        this.showOverrideModal.set(false);
        this.loadOrder(o.id!);
      },
      error: (err) => { 
        this.voidingOrder.set(false); 
        if (err.status === 403 && err.error?.error?.includes('Manager override')) {
          this.showOverrideModal.set(true);
        } else if (overridePin) {
          alert(err.error?.error || 'Invalid override PIN');
        }
      }
    });
  }

  markCompleted() {
    const o = this.order();
    if (!o?.id) return;
    this.completingOrder.set(true);
    this.orderService.completeOrder(o.id).subscribe({
      next: () => {
        this.completingOrder.set(false);
        this.loadOrder(o.id!);
      },
      error: () => { this.completingOrder.set(false); }
    });
  }

  quickReturn() {
    const o = this.order();
    if (!o?.id) return;
    this.router.navigate(['/returns'], { queryParams: { new: 'true', order_id: o.id } });
  }

  printReceipt() {
    window.print();
  }
}
