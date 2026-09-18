import { Component, ViewEncapsulation, OnInit, OnDestroy, inject } from '@angular/core';

import { RouterModule } from '@angular/router';
import { Testimonials } from '../landing/components/testimonials/testimonials';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-analytics-product',
  standalone: true,
  imports: [RouterModule, Testimonials],
  templateUrl: './analytics-product.html',
  encapsulation: ViewEncapsulation.None,
})
export class AnalyticsProduct implements OnInit, OnDestroy {
  private seo = inject(SeoService);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Analytics & Reporting — AI-Powered Business Insights | Puxbay',
      description: 'Real-time dashboards, profit & loss statements, AI insights, and tax reporting to grow your bottom line.',
      keywords: 'business analytics, AI insights, sales reports, profit loss statement, tax reporting, retail analytics',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      'name': 'Puxbay Analytics',
      'applicationCategory': 'BusinessApplication',
      'operatingSystem': 'Web',
      'description': 'Real-time dashboards, profit & loss statements, AI-powered insights, and tax reporting to grow your bottom line.',
      'url': 'https://puxbay.com/product/analytics',
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
      { name: 'Analytics', url: 'https://puxbay.com/product/analytics' }
    ]);
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }
}

