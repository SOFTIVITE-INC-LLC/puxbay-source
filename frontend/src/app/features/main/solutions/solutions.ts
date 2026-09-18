import { Component, ViewEncapsulation, OnInit, OnDestroy, inject } from '@angular/core';

import { RouterModule } from '@angular/router';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-solutions',
  standalone: true,
  imports: [RouterModule],
  templateUrl: './solutions.html',
  encapsulation: ViewEncapsulation.None,
})
export class Solutions implements OnInit, OnDestroy {
  private seo = inject(SeoService);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Solutions for Retail, F&B, Pharmacy & More | Puxbay',
      description: 'See how Puxbay empowers retail stores, restaurants, pharmacies, and multi-location businesses with a unified commerce platform.',
      keywords: 'retail solutions, restaurant POS, pharmacy software, multi-location business, commerce platform',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'WebPage',
      'name': 'Puxbay Solutions',
      'description': 'Industry-specific solutions for retail, F&B, pharmacy, and multi-location businesses.',
      'url': 'https://puxbay.com/solutions',
      'isPartOf': {
        '@type': 'WebSite',
        'name': 'Puxbay',
        'url': 'https://puxbay.com'
      }
    });

    this.seo.setBreadcrumbJsonLd([
      { name: 'Home', url: 'https://puxbay.com/' },
      { name: 'Solutions', url: 'https://puxbay.com/solutions' }
    ]);
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }
}
