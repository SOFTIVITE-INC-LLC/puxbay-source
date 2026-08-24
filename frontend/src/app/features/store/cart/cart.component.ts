import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { CartService } from '../../../core/store/services/cart.service';
import { ToastService } from '../../../core/store/services/toast.service';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-cart',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule, AppCurrencyPipe],
  templateUrl: './cart.component.html'
})
export class CartComponent implements OnInit {
  cartService = inject(CartService);
  toastService = inject(ToastService);
  settingsService = inject(StorefrontSettingsService);
  router = inject(Router);
  http = inject(HttpClient);

  isLoading = signal(true);
  couponCode = signal('');
  couponDiscount = signal(0);
  couponError = signal('');
  couponApplied = signal(false);
  applyingCoupon = signal(false);

  ngOnInit() {
    this.cartService.loadCart();
    // Give the cart a moment to hydrate
    setTimeout(() => this.isLoading.set(false), 600);
  }

  updateQuantity(productId: string, newQty: number) {
    if (newQty < 1) return;
    this.cartService.updateQuantity(productId, newQty).subscribe();
  }

  removeItem(productId: string) {
    this.cartService.removeItem(productId).subscribe({
      next: () => this.toastService.show('Item removed from cart', 'info'),
      error: () => this.toastService.show('Failed to remove item', 'error')
    });
  }

  applyCoupon() {
    const code = this.couponCode().trim();
    if (!code) return;

    this.applyingCoupon.set(true);
    this.couponError.set('');
    this.http.post<any>('/api/v1/storefront/coupon/apply', {
      code,
      cart_total: this.cartService.cartTotal()
    }).subscribe({
      next: (res) => {
        this.couponDiscount.set(res.discount_amount || 0);
        this.couponApplied.set(true);
        this.applyingCoupon.set(false);
        this.toastService.show('Coupon applied successfully', 'success');
      },
      error: (err) => {
        this.couponError.set(err.error?.error || 'Invalid coupon code');
        this.applyingCoupon.set(false);
      }
    });
  }

  removeCoupon() {
    this.couponCode.set('');
    this.couponDiscount.set(0);
    this.couponApplied.set(false);
    this.couponError.set('');
    this.toastService.show('Coupon removed', 'info');
  }

  get finalTotal(): number {
    return Math.max(0, this.cartService.cartTotal() - this.couponDiscount());
  }

  get minOrderMet(): boolean {
    const minAmount = this.settingsService.settings()?.min_order_amount || 0;
    return this.finalTotal >= minAmount;
  }

  get minOrderDifference(): number {
    const minAmount = this.settingsService.settings()?.min_order_amount || 0;
    return Math.max(0, minAmount - this.finalTotal);
  }

  proceedToCheckout() {
    if (!this.minOrderMet) {
      this.toastService.show(`Minimum order amount is not met`, 'error');
      return;
    }
    this.router.navigate(['/store/checkout']);
  }
}
