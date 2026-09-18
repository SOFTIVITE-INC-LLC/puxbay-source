import { Component, OnInit, OnDestroy, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { LegalService, LegalDocument } from '../../../core/services/legal.service';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-privacy',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './privacy.html',
})
export class Privacy implements OnInit, OnDestroy {
  private legalService = inject(LegalService);
  private seo = inject(SeoService);

  doc = signal<LegalDocument | null>(null);
  isLoading = signal(true);
  hasError = signal(false);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Privacy Policy | Puxbay',
      description: 'Read Puxbay\'s privacy policy to understand how we collect, use, and protect your personal information.',
      keywords: 'privacy policy, data protection, personal information, Puxbay privacy',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'WebPage',
      'name': 'Privacy Policy',
      'description': 'How Puxbay collects, uses, and protects your personal information.',
      'url': 'https://puxbay.com/privacy-policy',
      'isPartOf': {
        '@type': 'WebSite',
        'name': 'Puxbay',
        'url': 'https://puxbay.com'
      }
    });

    this.seo.setBreadcrumbJsonLd([
      { name: 'Home', url: 'https://puxbay.com/' },
      { name: 'Privacy Policy', url: 'https://puxbay.com/privacy-policy' }
    ]);

    this.legalService.getLegalDocument('privacy').subscribe({
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

