import { Component, ViewEncapsulation, OnInit, OnDestroy, inject } from '@angular/core';

import { RouterModule } from '@angular/router';
import { Testimonials } from '../landing/components/testimonials/testimonials';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-inventory-product',
  standalone: true,
  imports: [RouterModule, Testimonials],
  templateUrl: './inventory-product.html',
  encapsulation: ViewEncapsulation.None,
})
export class InventoryProduct implements OnInit, OnDestroy {
  private seo = inject(SeoService);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Inventory Management — Real-Time Stock Tracking | Puxbay',
      description: 'Multi-warehouse inventory management with low-stock alerts, batch tracking, supplier POs, and AI-powered forecasting.',
      keywords: 'inventory management, stock tracking, warehouse management, low stock alerts, batch tracking, purchase orders',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      'name': 'Puxbay Inventory Management',
      'applicationCategory': 'BusinessApplication',
      'operatingSystem': 'Web',
      'description': 'Multi-warehouse inventory management with low-stock alerts, batch tracking, supplier POs, and AI-powered forecasting.',
      'url': 'https://puxbay.com/product/inventory',
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
      { name: 'Inventory', url: 'https://puxbay.com/product/inventory' }
    ]);
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }
}
