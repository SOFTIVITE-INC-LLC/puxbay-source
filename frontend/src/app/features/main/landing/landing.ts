import { Component, ViewEncapsulation, OnInit, OnDestroy, inject } from '@angular/core';

import { RouterModule } from '@angular/router';
import { Hero } from './components/hero/hero';
import { FeatureOverview } from './components/feature-overview/feature-overview';
import { Testimonials } from './components/testimonials/testimonials';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-landing',
  standalone: true,
  imports: [RouterModule, Hero, FeatureOverview, Testimonials],
  templateUrl: './landing.html',
  styleUrls: ['./landing.css'],
  encapsulation: ViewEncapsulation.None,
})
export class Landing implements OnInit, OnDestroy {
  private seo = inject(SeoService);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Puxbay — All-in-One POS, Inventory & E-Commerce Platform',
      description: 'Puxbay is the modern point-of-sale, inventory management, and e-commerce platform trusted by 2,000+ businesses. AI-powered insights, offline mode, and more.',
      keywords: 'POS, point of sale, inventory management, e-commerce, storefront, retail software, AI analytics, Puxbay',
    });

    this.seo.setJsonLd([
      {
        '@context': 'https://schema.org',
        '@type': 'WebSite',
        'name': 'Puxbay',
        'url': 'https://puxbay.com',
        'description': 'All-in-one POS, inventory management, and e-commerce platform for modern businesses.',
        'potentialAction': {
          '@type': 'SearchAction',
          'target': 'https://puxbay.com/store?q={search_term_string}',
          'query-input': 'required name=search_term_string'
        }
      },
      {
        '@context': 'https://schema.org',
        '@type': 'Organization',
        'name': 'Puxbay',
        'legalName': 'Softivite Inc LLC',
        'url': 'https://puxbay.com',
        'logo': 'https://puxbay.com/logo.png',
        'sameAs': [],
        'contactPoint': {
          '@type': 'ContactPoint',
          'email': 'support@puxbay.com',
          'contactType': 'customer support'
        }
      }
    ]);
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }
}
