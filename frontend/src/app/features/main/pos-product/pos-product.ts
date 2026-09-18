import { Component, ViewEncapsulation, signal, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { Testimonials } from '../landing/components/testimonials/testimonials';
import { inject } from '@angular/core';
import { SettingsService as AppSettingsService } from '../../../core/services/settings.service';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-pos-product',
  standalone: true,
  imports: [RouterModule, CommonModule, Testimonials],
  templateUrl: './pos-product.html',
  encapsulation: ViewEncapsulation.None,
})
export class PosProduct implements OnInit, OnDestroy {
  settingsService = inject(AppSettingsService);
  private seo = inject(SeoService);

  // Simulator State
  cart = signal([
    { name: 'Artisan Coffee Beans', price: 24.99 },
    { name: 'Ceramic Mug', price: 12.00 }
  ]);

  cartTotal = signal(36.99);
  isCheckingOut = signal(false);
  checkoutSuccess = signal(false);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Point of Sale System — Fast, Offline-Ready POS | Puxbay',
      description: 'Lightning-fast cloud POS with offline mode, barcode scanning, multi-branch support, and custom receipts. Built for modern retail.',
      keywords: 'POS system, point of sale, offline POS, barcode scanner POS, cloud POS, retail POS',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      'name': 'Puxbay POS',
      'applicationCategory': 'BusinessApplication',
      'operatingSystem': 'Web, iOS, Android',
      'description': 'Lightning-fast cloud POS with offline mode, barcode scanning, multi-branch support, and custom receipts.',
      'url': 'https://puxbay.com/product/pos',
      'offers': {
        '@type': 'Offer',
        'priceCurrency': 'USD',
        'price': '0',
        'description': 'Free plan available'
      }
    });

    this.seo.setBreadcrumbJsonLd([
      { name: 'Home', url: 'https://puxbay.com/' },
      { name: 'Products', url: 'https://puxbay.com/features' },
      { name: 'POS', url: 'https://puxbay.com/product/pos' }
    ]);
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }

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
