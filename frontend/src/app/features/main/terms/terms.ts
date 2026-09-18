import { Component, OnInit, OnDestroy, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { LegalService, LegalDocument } from '../../../core/services/legal.service';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-terms',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './terms.html',
})
export class Terms implements OnInit, OnDestroy {
  private legalService = inject(LegalService);
  private seo = inject(SeoService);

  doc = signal<LegalDocument | null>(null);
  isLoading = signal(true);
  hasError = signal(false);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Terms of Service | Puxbay',
      description: 'Review Puxbay\'s terms of service covering usage, responsibilities, and legal agreements.',
      keywords: 'terms of service, terms and conditions, user agreement, Puxbay terms',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'WebPage',
      'name': 'Terms of Service',
      'description': 'Puxbay terms of service covering usage, responsibilities, and legal agreements.',
      'url': 'https://puxbay.com/terms',
      'isPartOf': {
        '@type': 'WebSite',
        'name': 'Puxbay',
        'url': 'https://puxbay.com'
      }
    });

    this.seo.setBreadcrumbJsonLd([
      { name: 'Home', url: 'https://puxbay.com/' },
      { name: 'Terms of Service', url: 'https://puxbay.com/terms' }
    ]);

    this.legalService.getLegalDocument('terms').subscribe({
      next: (d) => { this.doc.set(d); this.isLoading.set(false); },
      error: () => { this.hasError.set(true); this.isLoading.set(false); }
    });
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }

  formatDate(dateStr?: string): string {
    if (!dateStr) return '';
    return new Date(dateStr).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
  }
}

