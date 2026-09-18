import { Component, ViewEncapsulation, OnInit, OnDestroy, inject } from '@angular/core';

import { RouterModule } from '@angular/router';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-careers',
  standalone: true,
  imports: [RouterModule],
  templateUrl: './careers.html',
  encapsulation: ViewEncapsulation.None,
})
export class Careers implements OnInit, OnDestroy {
  private seo = inject(SeoService);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Careers at Puxbay — Join Our Team',
      description: 'Explore career opportunities at Puxbay. We\'re building the future of commerce in Africa. Join our team of innovators.',
      keywords: 'Puxbay careers, jobs, software engineer, Africa tech jobs, Softivite careers',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'WebPage',
      'name': 'Careers at Puxbay',
      'description': 'Explore career opportunities at Puxbay. We\'re building the future of commerce in Africa.',
      'url': 'https://puxbay.com/careers',
      'about': {
        '@type': 'Organization',
        'name': 'Puxbay',
        'legalName': 'Softivite Inc LLC',
        'url': 'https://puxbay.com',
        'logo': 'https://puxbay.com/logo.png',
        'address': {
          '@type': 'PostalAddress',
          'addressLocality': 'Accra',
          'addressCountry': 'GH'
        }
      }
    });

    this.seo.setBreadcrumbJsonLd([
      { name: 'Home', url: 'https://puxbay.com/' },
      { name: 'Careers', url: 'https://puxbay.com/careers' }
    ]);
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }
}
