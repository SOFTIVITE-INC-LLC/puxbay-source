import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { PublicPortalService } from '../../core/services/public-portal.service';
import { ToastService } from '../../core/services/toast';

@Component({
  selector: 'app-public-feedback',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './public-feedback.html',
  styles: `
    .glass-card {
      background: rgba(255, 255, 255, 0.7);
      backdrop-filter: blur(20px);
      border: 1px solid rgba(255, 255, 255, 0.5);
      box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.07);
    }
    .dark .glass-card {
      background: rgba(24, 24, 27, 0.7);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .gradient-text { color: #005b96; }
  `,
})
export class PublicFeedback implements OnInit {
  private route = inject(ActivatedRoute);
  private publicPortal = inject(PublicPortalService);
  private toastService = inject(ToastService);

  tenantId = signal<string | null>(null);
  
  name = signal('');
  email = signal('');
  rating = signal(0);
  comment = signal('');
  
  submitting = signal(false);
  submitted = signal(false);

  ngOnInit() {
    this.route.paramMap.subscribe(params => {
      this.tenantId.set(params.get('tenant_id'));
    });
  }

  setRating(star: number) {
    this.rating.set(star);
  }

  submit() {
    if (!this.tenantId()) {
      this.toastService.showError('Invalid store link.');
      return;
    }
    if (!this.name() || !this.email() || this.rating() === 0) {
      this.toastService.showError('Please fill in your name, email, and select a rating.');
      return;
    }

    this.submitting.set(true);
    this.publicPortal.submitFeedback(this.tenantId()!, {
      name: this.name(),
      email: this.email(),
      rating: this.rating(),
      comment: this.comment()
    }).subscribe({
      next: () => {
        this.submitting.set(false);
        this.submitted.set(true);
      },
      error: () => {
        this.submitting.set(false);
        this.toastService.showError('There was a problem submitting your feedback. Please try again.');
      }
    });
  }
}
