import { Component, inject, OnInit, signal } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { Router, ActivatedRoute } from '@angular/router';
import { BillingService } from '../../../core/services/billing.service';
import { ApiService } from '../../../core/services/api.service';

@Component({
  selector: 'app-billing',
  standalone: true,
  imports: [CommonModule, AppCurrencyPipe],
  templateUrl: './billing.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .dark .glass-panel {
      background: rgba(0, 0, 0, 0.2);
    }
  `,
})
export class Billing implements OnInit {
  billingService = inject(BillingService);
  router = inject(Router);
  route = inject(ActivatedRoute);
  api = inject(ApiService);

  verifying = signal(false);
  verifySuccess = signal(false);
  verifyError = signal('');

  ngOnInit() {
    // Check if we're returning from Paystack payment callback
    const reference = this.route.snapshot.queryParamMap.get('reference')
      || this.route.snapshot.queryParamMap.get('trxref');

    if (reference) {
      this.verifyPayment(reference);
    } else {
      this.loadBillingData();
    }
  }

  loadBillingData() {
    this.billingService.getSubscription().subscribe();
    this.billingService.getInvoices().subscribe();
  }

  verifyPayment(reference: string) {
    this.verifying.set(true);
    this.api.get<{ status: string; message: string }>(`/billing/verify/${reference}`).subscribe({
      next: () => {
        this.verifySuccess.set(true);
        this.verifying.set(false);
        // Remove query params from URL without reloading
        this.router.navigate([], { queryParams: {}, replaceUrl: true });
        // Reload fresh subscription/invoice data
        this.loadBillingData();
      },
      error: (err) => {
        this.verifyError.set(err?.error?.error || 'Payment verification failed');
        this.verifying.set(false);
        this.loadBillingData();
      }
    });
  }

  upgradePlan() {
    // Navigate to the global pricing page
    window.location.href = window.location.protocol + '//' + window.location.hostname.replace(/^[a-zA-Z0-9-]+\./, '') + (window.location.port ? ':' + window.location.port : '') + '/pricing';
  }
}
