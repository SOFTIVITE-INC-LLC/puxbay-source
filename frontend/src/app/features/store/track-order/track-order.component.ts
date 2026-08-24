import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { OrderService } from '../../../core/store/services/order.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-track-order',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './track-order.component.html'
})
export class TrackOrderComponent {
  orderService = inject(OrderService);
  
  orderNumber = signal('');
  isTracking = signal(false);
  order = signal<any | null>(null);
  errorMsg = signal('');

  track() {
    const num = this.orderNumber().trim();
    if (!num || this.isTracking()) return;

    this.isTracking.set(true);
    this.errorMsg.set('');
    this.order.set(null);

    this.orderService.trackOrder(num).subscribe({
      next: (res) => {
        this.order.set(res.order);
        this.isTracking.set(false);
      },
      error: (err) => {
        this.errorMsg.set(err.error?.error || 'Order not found or invalid order number');
        this.isTracking.set(false);
      }
    });
  }

  getStepStatus(status: string, step: number): string {
    const statuses = ['pending', 'processing', 'shipped', 'delivered', 'cancelled'];
    const currentIndex = statuses.indexOf(status.toLowerCase());
    if (currentIndex === -1) return 'pending'; // fallback

    if (currentIndex === 4) { // cancelled
      return step === 0 ? 'cancelled' : 'pending';
    }

    if (step < currentIndex) return 'completed';
    if (step === currentIndex) return 'active';
    return 'pending';
  }
}
