import { Component, ViewEncapsulation, OnInit, OnDestroy, inject } from '@angular/core';

import { RouterModule } from '@angular/router';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-about',
  standalone: true,
  imports: [RouterModule],
  templateUrl: './about.html',
  encapsulation: ViewEncapsulation.None,
})
export class About implements OnInit, OnDestroy {
  private seo = inject(SeoService);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'About Puxbay — Our Mission & Story',
      description: 'Learn about Puxbay\'s mission to empower businesses across Africa with modern commerce technology. Built by Softivite Inc LLC.',
      keywords: 'about Puxbay, Softivite, commerce technology, Africa, retail technology',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'Organization',
      'name': 'Puxbay',
      'legalName': 'Softivite Inc LLC',
      'url': 'https://puxbay.com',
      'logo': 'https://puxbay.com/logo.png',
      'description': 'Puxbay empowers businesses across Africa with modern commerce technology — POS, inventory, e-commerce, and analytics.',
      'address': {
        '@type': 'PostalAddress',
        'addressLocality': 'Accra',
        'addressCountry': 'GH'
      }
    });
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }
}
