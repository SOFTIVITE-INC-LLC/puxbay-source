import { Component, ViewEncapsulation, inject, signal, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { Testimonials } from '../landing/components/testimonials/testimonials';
import { SettingsService } from '../../../core/services/settings.service';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-storefront-product',
  standalone: true,
  imports: [RouterModule, CommonModule, Testimonials],
  templateUrl: './storefront-product.html',
  encapsulation: ViewEncapsulation.None,
})
export class StorefrontProduct implements OnInit, OnDestroy {
  settingsService = inject(SettingsService);
  private seo = inject(SeoService);
  isScanning = signal(false);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'E-Commerce Storefront — Launch Your Online Store | Puxbay',
      description: 'Go online instantly with Puxbay\'s free storefront. Mobile-optimized checkout, delivery integrations, and global payment gateways.',
      keywords: 'e-commerce storefront, online store builder, mobile commerce, delivery integration, payment gateway, Paystack, Stripe',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      'name': 'Puxbay Storefront',
      'applicationCategory': 'BusinessApplication',
      'operatingSystem': 'Web',
      'description': 'Launch your online store instantly with mobile-optimized checkout, delivery integrations, and global payment gateways.',
      'url': 'https://puxbay.com/product/storefront',
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
      { name: 'Storefront', url: 'https://puxbay.com/product/storefront' }
    ]);
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }

  toggleScan() {
    this.isScanning.set(!this.isScanning());
  }
}
