import { Component, ViewEncapsulation, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { RouterModule } from '@angular/router';

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
export class Contact {
  fb = inject(FormBuilder);
  http = inject(HttpClient);
  
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

  loadContactInfo() {
    try {
      const saved = localStorage.getItem('puxbay_platform_contact');
      if (saved) {
        this.contactInfo.update(curr => ({ ...curr, ...JSON.parse(saved) }));
      }
    } catch (_) {}

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
