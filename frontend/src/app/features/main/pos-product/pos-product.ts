import { Component, ViewEncapsulation, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { Testimonials } from '../landing/components/testimonials/testimonials';
import { inject } from '@angular/core';
import { SettingsService as AppSettingsService } from '../../../core/services/settings.service';

@Component({
  selector: 'app-pos-product',
  standalone: true,
  imports: [RouterModule, CommonModule, Testimonials],
  templateUrl: './pos-product.html',
  encapsulation: ViewEncapsulation.None,
})
export class PosProduct {
  settingsService = inject(AppSettingsService);
  // Simulator State
  cart = signal([
    { name: 'Artisan Coffee Beans', price: 24.99 },
    { name: 'Ceramic Mug', price: 12.00 }
  ]);

  cartTotal = signal(36.99);
  isCheckingOut = signal(false);
  checkoutSuccess = signal(false);

  simulateScan() {
    if (this.checkoutSuccess()) {
      this.resetDemo();
    }

    const items = [
      { name: 'Organic Matcha', price: 18.50 },
      { name: 'Oat Milk Carton', price: 4.99 },
      { name: 'Espresso Tamper', price: 35.00 }
    ];

    const randomItem = items[Math.floor(Math.random() * items.length)];
    this.cart.update(c => [...c, randomItem]);
    this.cartTotal.update(t => t + randomItem.price);
  }

  simulateCheckout() {
    if (this.cart().length === 0 || this.isCheckingOut()) return;

    this.isCheckingOut.set(true);

    // Simulate sub-second offline sync processing
    setTimeout(() => {
      this.isCheckingOut.set(false);
      this.checkoutSuccess.set(true);
    }, 600);
  }

  resetDemo() {
    this.cart.set([]);
    this.cartTotal.set(0);
    this.checkoutSuccess.set(false);
  }
}
