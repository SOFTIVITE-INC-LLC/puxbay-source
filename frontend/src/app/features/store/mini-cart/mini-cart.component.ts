import { Component, EventEmitter, inject, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { CartService } from '../../../core/store/services/cart.service';
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
}
