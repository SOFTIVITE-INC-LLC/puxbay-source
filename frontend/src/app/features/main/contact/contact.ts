import { Component, ViewEncapsulation, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { RouterModule } from '@angular/router';

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
