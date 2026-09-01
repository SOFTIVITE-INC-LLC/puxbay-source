import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { OrderService } from '../../../core/store/services/order.service';
import { ToastService } from '../../../core/store/services/toast.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-track-order',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe],
  templateUrl: './track-order.component.html'
})
export class TrackOrderComponent implements OnInit {
  orderService = inject(OrderService);
  toastService = inject(ToastService);
  route = inject(ActivatedRoute);

  orderNumber = signal('');
  isTracking = signal(false);
  order = signal<any | null>(null);
  errorMsg = signal('');
  copied = signal(false);

  ngOnInit() {
    this.route.queryParamMap.subscribe(params => {
      const code = params.get('code') || params.get('order_number') || params.get('tracking_code');
      if (code) {
        this.orderNumber.set(code.toUpperCase().trim());
        this.track();
      }
    });
  }

  onCodeInput(val: string) {
    this.orderNumber.set(val.toUpperCase());
  }

  track() {
    const num = this.orderNumber().trim().toUpperCase();
    if (!num || this.isTracking()) return;

    this.isTracking.set(true);
    this.errorMsg.set('');
    this.order.set(null);

    this.orderService.trackOrder(num).subscribe({
      next: (res) => {
        this.order.set(res.order || res);
        this.isTracking.set(false);
      },
      error: (err) => {
        this.errorMsg.set(err.error?.error || 'Order not found. Please check your 8-character tracking code.');
        this.isTracking.set(false);
      }
    });
  }

  getStepStatus(status: string, step: number): string {
    const s = (status || '').toLowerCase();
    if (s === 'cancelled' || s === 'voided') {
      return step === 0 ? 'cancelled' : 'pending';
    }

    let currentIndex = 0;
    if (s === 'pending') currentIndex = 0;
    else if (s === 'processing' || s === 'preparing') currentIndex = 1;
    else if (s === 'shipped' || s === 'ready' || s === 'out_for_delivery') currentIndex = 2;
    else if (s === 'delivered' || s === 'completed') currentIndex = 3;

    if (step < currentIndex) return 'completed';
    if (step === currentIndex) return 'active';
    return 'pending';
  }

  copyTrackingCode() {
    const o = this.order();
    const code = o?.order_number || this.orderNumber();
    if (!code) return;
    navigator.clipboard.writeText(code).then(() => {
      this.copied.set(true);
      this.toastService.show('Tracking code copied to clipboard!', 'success');
      setTimeout(() => this.copied.set(false), 2500);
    });
  }

  shareOnWhatsApp() {
    const o = this.order();
    if (!o) return;
    const code = o.order_number || this.orderNumber();
    const url = `${window.location.origin}/store/track-order?code=${code}`;
    const text = `*Order Tracking Update*\n\n• *Tracking Code:* ${code}\n• *Status:* ${o.status.toUpperCase()}\n• *Total:* GHS ${Number(o.total || 0).toFixed(2)}\n\nTrack live status here:\n${url}`;
    window.open(`https://wa.me/?text=${encodeURIComponent(text)}`, '_blank');
  }
}
