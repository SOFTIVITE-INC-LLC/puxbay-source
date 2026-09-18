import { Component, OnInit, OnDestroy, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { LegalService, LegalDocument } from '../../../core/services/legal.service';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-cookies',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './cookies.html',
})
export class Cookies implements OnInit, OnDestroy {
  private legalService = inject(LegalService);
  private seo = inject(SeoService);

  doc = signal<LegalDocument | null>(null);
  isLoading = signal(true);
  hasError = signal(false);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Cookie Policy | Puxbay',
      description: 'Learn about how Puxbay uses cookies and similar technologies on our platform.',
      keywords: 'cookie policy, cookies, tracking, Puxbay cookies',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'WebPage',
      'name': 'Cookie Policy',
      'description': 'How Puxbay uses cookies and similar technologies on our platform.',
      'url': 'https://puxbay.com/cookie-policy',
      'isPartOf': {
        '@type': 'WebSite',
        'name': 'Puxbay',
        'url': 'https://puxbay.com'
      }
    });

    this.seo.setBreadcrumbJsonLd([
      { name: 'Home', url: 'https://puxbay.com/' },
      { name: 'Cookie Policy', url: 'https://puxbay.com/cookie-policy' }
    ]);

    this.legalService.getLegalDocument('cookie').subscribe({
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

