import { Component, OnInit, OnDestroy, ViewEncapsulation, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { BlogService, BlogPost } from '../../../core/services/blog.service';
import { SeoService } from '../../../core/services/seo.service';

@Component({
  selector: 'app-blog',
  standalone: true,
  imports: [RouterModule, CommonModule],
  templateUrl: './blog.html',
  encapsulation: ViewEncapsulation.None,
})
export class Blog implements OnInit, OnDestroy {
  private blogService = inject(BlogService);
  private seo = inject(SeoService);

  posts = signal<BlogPost[]>([]);
  isLoading = signal(true);
  hasError = signal(false);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Blog — Commerce Insights & Updates | Puxbay',
      description: 'Read the latest insights on retail technology, inventory management, e-commerce trends, and Puxbay product updates.',
      keywords: 'Puxbay blog, retail technology, e-commerce trends, inventory tips, commerce insights',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'CollectionPage',
      'name': 'Puxbay Blog',
      'description': 'Commerce insights, retail technology trends, and Puxbay product updates.',
      'url': 'https://puxbay.com/blog',
      'isPartOf': {
        '@type': 'WebSite',
        'name': 'Puxbay',
        'url': 'https://puxbay.com'
      },
      'publisher': {
        '@type': 'Organization',
        'name': 'Puxbay',
        'logo': {
          '@type': 'ImageObject',
          'url': 'https://puxbay.com/logo.png'
        }
      }
    });

    this.seo.setBreadcrumbJsonLd([
      { name: 'Home', url: 'https://puxbay.com/' },
      { name: 'Blog', url: 'https://puxbay.com/blog' }
    ]);

    this.blogService.getBlogPosts().subscribe({
      next: (res) => {
        this.posts.set(res.posts || []);
        this.isLoading.set(false);
      },
      error: () => {
        this.hasError.set(true);
        this.isLoading.set(false);
      }
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

