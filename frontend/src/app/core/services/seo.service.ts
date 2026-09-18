import { Injectable, Inject, inject } from '@angular/core';
import { Title, Meta } from '@angular/platform-browser';
import { DOCUMENT } from '@angular/common';
import { Router } from '@angular/router';

export interface PageSeoConfig {
  title: string;
  description: string;
  url?: string;
  image?: string;
  imageAlt?: string;
  imageWidth?: number;
  imageHeight?: number;
  type?: string; // og:type — 'website' | 'article' | 'product'
  keywords?: string;
  noindex?: boolean;
  locale?: string;
  author?: string;
  twitterSite?: string;
}

const BASE_URL = 'https://puxbay.com';
const DEFAULT_IMAGE = `${BASE_URL}/og-image.png`;
const DEFAULT_IMAGE_ALT = 'Puxbay — All-in-One POS, Inventory & E-Commerce Platform';
const DEFAULT_IMAGE_WIDTH = 1200;
const DEFAULT_IMAGE_HEIGHT = 630;
const SITE_NAME = 'Puxbay';
const DEFAULT_LOCALE = 'en_US';
const DEFAULT_TWITTER_SITE = '@puxbay';
const DEFAULT_AUTHOR = 'Puxbay';

@Injectable({ providedIn: 'root' })
export class SeoService {
  private titleService = inject(Title);
  private metaService = inject(Meta);
  private router = inject(Router);

  constructor(@Inject(DOCUMENT) private document: Document) {}

  /**
   * Sets all page-level SEO tags: title, meta description, canonical URL,
   * Open Graph tags, Twitter Card tags, author, and locale.
   */
  setPageSeo(config: PageSeoConfig): void {
    const url = config.url || `${BASE_URL}${this.router.url}`;
    const image = config.image || DEFAULT_IMAGE;
    const imageAlt = config.imageAlt || config.title || DEFAULT_IMAGE_ALT;
    const imageWidth = config.imageWidth || DEFAULT_IMAGE_WIDTH;
    const imageHeight = config.imageHeight || DEFAULT_IMAGE_HEIGHT;
    const type = config.type || 'website';
    const locale = config.locale || DEFAULT_LOCALE;
    const author = config.author || DEFAULT_AUTHOR;
    const twitterSite = config.twitterSite || DEFAULT_TWITTER_SITE;

    // Title
    this.titleService.setTitle(config.title);

    // Meta description
    this.metaService.updateTag({ name: 'description', content: config.description });

    // Author
    this.metaService.updateTag({ name: 'author', content: author });

    // Keywords (optional)
    if (config.keywords) {
      this.metaService.updateTag({ name: 'keywords', content: config.keywords });
    }

    // Robots
    if (config.noindex) {
      this.metaService.updateTag({ name: 'robots', content: 'noindex, nofollow' });
    } else {
      this.metaService.updateTag({ name: 'robots', content: 'index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1' });
    }

    // Canonical URL
    this.setCanonicalUrl(url);

    // Open Graph
    this.metaService.updateTag({ property: 'og:title', content: config.title });
    this.metaService.updateTag({ property: 'og:description', content: config.description });
    this.metaService.updateTag({ property: 'og:url', content: url });
    this.metaService.updateTag({ property: 'og:image', content: image });
    this.metaService.updateTag({ property: 'og:image:alt', content: imageAlt });
    this.metaService.updateTag({ property: 'og:image:width', content: String(imageWidth) });
    this.metaService.updateTag({ property: 'og:image:height', content: String(imageHeight) });
    this.metaService.updateTag({ property: 'og:type', content: type });
    this.metaService.updateTag({ property: 'og:site_name', content: SITE_NAME });
    this.metaService.updateTag({ property: 'og:locale', content: locale });

    // Twitter Card
    this.metaService.updateTag({ name: 'twitter:card', content: 'summary_large_image' });
    this.metaService.updateTag({ name: 'twitter:site', content: twitterSite });
    this.metaService.updateTag({ name: 'twitter:title', content: config.title });
    this.metaService.updateTag({ name: 'twitter:description', content: config.description });
    this.metaService.updateTag({ name: 'twitter:image', content: image });
    this.metaService.updateTag({ name: 'twitter:image:alt', content: imageAlt });
  }

  /**
   * Injects or updates a JSON-LD structured data script in <head>.
   */
  setJsonLd(data: Record<string, unknown> | Record<string, unknown>[]): void {
    this.removeJsonLd();
    const script = this.document.createElement('script');
    script.type = 'application/ld+json';
    script.id = 'seo-json-ld';
    script.textContent = JSON.stringify(data);
    this.document.head.appendChild(script);
  }

  /**
   * Removes any existing JSON-LD structured data script.
   */
  removeJsonLd(): void {
    const existing = this.document.getElementById('seo-json-ld');
    if (existing) {
      existing.remove();
    }
  }

  /**
   * Generates and injects BreadcrumbList JSON-LD structured data.
   * @param items Array of breadcrumb items with name and url.
   */
  setBreadcrumbJsonLd(items: { name: string; url: string }[]): void {
    const breadcrumbLd = {
      '@context': 'https://schema.org',
      '@type': 'BreadcrumbList',
      'itemListElement': items.map((item, index) => ({
        '@type': 'ListItem',
        'position': index + 1,
        'name': item.name,
        'item': item.url
      }))
    };

    // Append breadcrumb alongside existing JSON-LD
    const existingScript = this.document.getElementById('seo-json-ld');
    if (existingScript && existingScript.textContent) {
      try {
        const existingData = JSON.parse(existingScript.textContent);
        const combined = Array.isArray(existingData)
          ? [...existingData, breadcrumbLd]
          : [existingData, breadcrumbLd];
        existingScript.textContent = JSON.stringify(combined);
      } catch {
        // If parsing fails, just set the breadcrumb
        existingScript.textContent = JSON.stringify(breadcrumbLd);
      }
    } else {
      this.setJsonLd(breadcrumbLd);
    }
  }

  /**
   * Generates and injects FAQPage JSON-LD structured data.
   * @param faqs Array of FAQ items with question and answer.
   */
  setFaqJsonLd(faqs: { question: string; answer: string }[]): void {
    const faqLd = {
      '@context': 'https://schema.org',
      '@type': 'FAQPage',
      'mainEntity': faqs.map(faq => ({
        '@type': 'Question',
        'name': faq.question,
        'acceptedAnswer': {
          '@type': 'Answer',
          'text': faq.answer
        }
      }))
    };

    const existingScript = this.document.getElementById('seo-json-ld');
    if (existingScript && existingScript.textContent) {
      try {
        const existingData = JSON.parse(existingScript.textContent);
        const combined = Array.isArray(existingData)
          ? [...existingData, faqLd]
          : [existingData, faqLd];
        existingScript.textContent = JSON.stringify(combined);
      } catch {
        existingScript.textContent = JSON.stringify(faqLd);
      }
    } else {
      this.setJsonLd(faqLd);
    }
  }

  /**
   * Sets or updates the canonical link element in <head>.
   */
  private setCanonicalUrl(url: string): void {
    let link = this.document.querySelector('link[rel="canonical"]') as HTMLLinkElement | null;
    if (!link) {
      link = this.document.createElement('link');
      link.setAttribute('rel', 'canonical');
      this.document.head.appendChild(link);
    }
    link.setAttribute('href', url);
  }
}
