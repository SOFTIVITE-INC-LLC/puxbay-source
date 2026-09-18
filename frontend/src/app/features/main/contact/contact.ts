import { Component, ViewEncapsulation, inject, signal, OnInit, OnDestroy } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { RouterModule } from '@angular/router';
import { SeoService } from '../../../core/services/seo.service';

export interface PlatformContact {
  company_name: string;
  headquarters_address: string;
  contact_phone: string;
  support_phone: string;
  contact_email: string;
  sales_email: string;
  support_email: string;
  working_hours: string;
}

@Component({
  selector: 'app-contact',
  standalone: true,
  imports: [RouterModule, ReactiveFormsModule],
  templateUrl: './contact.html',
  encapsulation: ViewEncapsulation.None,
})
export class Contact implements OnInit, OnDestroy {
  fb = inject(FormBuilder);
  http = inject(HttpClient);
  private seo = inject(SeoService);
  
  status = signal<'idle' | 'submitting' | 'success' | 'error'>('idle');

  contactInfo = signal<PlatformContact>({
    company_name: 'Puxbay / Softivite',
    headquarters_address: 'No. 12 Independence Avenue, Ridge, Accra, Ghana',
    contact_phone: '+233 (0) 30 123 4567',
    support_phone: '+233 (0) 50 123 4567',
    contact_email: 'support@puxbay.com',
    sales_email: 'sales@puxbay.com',
    support_email: 'support@puxbay.com',
    working_hours: 'Mon - Fri, 8:00 AM - 6:00 PM GMT'
  });

  constructor() {
    this.loadContactInfo();
  }

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Contact Us — Get in Touch with Puxbay',
      description: 'Have questions? Reach out to the Puxbay team for sales inquiries, technical support, or partnership opportunities.',
      keywords: 'contact Puxbay, Puxbay support, sales inquiry, customer support, partnership',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'ContactPage',
      'name': 'Contact Puxbay',
      'url': 'https://puxbay.com/contact',
      'mainEntity': {
        '@type': 'Organization',
        'name': 'Puxbay',
        'email': 'support@puxbay.com',
        'telephone': '+233 (0) 30 123 4567',
        'address': {
          '@type': 'PostalAddress',
          'streetAddress': 'No. 12 Independence Avenue, Ridge',
          'addressLocality': 'Accra',
          'addressCountry': 'GH'
        },
        'contactPoint': [
          {
            '@type': 'ContactPoint',
            'contactType': 'sales',
            'email': 'sales@puxbay.com'
          },
          {
            '@type': 'ContactPoint',
            'contactType': 'customer support',
            'email': 'support@puxbay.com',
            'telephone': '+233 (0) 50 123 4567'
          }
        ]
      }
    });
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }

  loadContactInfo() {
    if (typeof localStorage !== 'undefined') {
      try {
        const saved = localStorage.getItem('puxbay_platform_contact');
        if (saved) {
          this.contactInfo.update(curr => ({ ...curr, ...JSON.parse(saved) }));
        }
      } catch (_) {}
    }

    this.http.get<PlatformContact>('/api/v1/public/contact-info').subscribe({
      next: (res) => {
        if (res) {
          this.contactInfo.set(res);
        }
      },
      error: () => {
        // Fallback to local storage or defaults
      }
    });
  }

  contactForm = this.fb.group({
    name: ['', Validators.required],
    email: ['', [Validators.required, Validators.email]],
    subject: ['', Validators.required],
    message: ['', Validators.required]
  });

  submit() {
    if (this.contactForm.invalid) return;
    this.status.set('submitting');
    
    // Using a mock URL or real API depending on backend
    this.http.post('/api/v1/marketing/contact', this.contactForm.value).subscribe({
      next: () => {
        this.status.set('success');
        this.contactForm.reset();
      },
      error: () => {
        this.status.set('error');
      }
    });
  }
}
