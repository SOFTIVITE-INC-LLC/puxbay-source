import { Component, EventEmitter, inject, Output, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { CartService } from '../../../core/store/services/cart.service';
import { ToastService } from '../../../core/store/services/toast.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-mini-cart',
  standalone: true,
  imports: [CommonModule, RouterModule, AppCurrencyPipe],
  templateUrl: './mini-cart.component.html'
})
export class MiniCartComponent {
  @Output() closeCart = new EventEmitter<void>();
  cartService = inject(CartService);
  toastService = inject(ToastService);

  copied = signal(false);

  get progressPercentage(): number {
    const total = this.cartService.cartTotal();
    const threshold = 150; // free shipping threshold
    return Math.min(100, (total / threshold) * 100);
  }

  get remainingForFreeShipping(): number {
    const total = this.cartService.cartTotal();
    const threshold = 150;
    return Math.max(0, threshold - total);
  }

  updateQuantity(productId: string, currentQty: number, change: number) {
    const newQty = currentQty + change;
    if (newQty < 1) {
      this.cartService.removeItem(productId).subscribe();
    } else {
      this.cartService.updateQuantity(productId, newQty).subscribe();
    }
  }

  remove(productId: string) {
    this.cartService.removeItem(productId).subscribe();
  }

  getShareableLink(): string {
    if (typeof window === 'undefined') return '';
    const items = this.cartService.cartItems().map(i => `${i.product_id}:${i.quantity}`).join(',');
    return `${window.location.origin}/store/cart?shared_cart=${encodeURIComponent(items)}`;
  }

  copyShareLink() {
    const link = this.getShareableLink();
    if (!link) return;
    navigator.clipboard.writeText(link).then(() => {
      this.copied.set(true);
      this.toastService.show('Cart link copied to clipboard!', 'success');
      setTimeout(() => this.copied.set(false), 2500);
    });
  }

  shareOnWhatsApp() {
    const details = this.cartService.cartDetails();
    if (details.length === 0) return;

    let text = `*My Shopping Cart*\n\n`;
    details.forEach(item => {
      text += `• ${item.quantity}x ${item.product.name} — ${(item.product.selling_price * item.quantity).toFixed(2)}\n`;
    });
    text += `\n*Total:* ${this.cartService.cartTotal().toFixed(2)}\n`;
    text += `\nView / Order Cart: ${this.getShareableLink()}`;

    const url = `https://wa.me/?text=${encodeURIComponent(text)}`;
    if (typeof window !== 'undefined') {
      window.open(url, '_blank');
    }
  }
}
