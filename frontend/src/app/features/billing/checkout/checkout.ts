import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, ActivatedRoute } from '@angular/router';
import { BillingService, Plan } from '../../../core/services/billing.service';
import { PricingService } from '../../../core/services/pricing.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { AlertService } from '../../../core/services/alert.service';

@Component({
  selector: 'app-checkout',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './checkout.html',
})
export class Checkout implements OnInit {
  private billingService = inject(BillingService);
  private pricingService = inject(PricingService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private alertService = inject(AlertService);

  planId = signal<string | null>(null);
  billingCycle = signal<'monthly' | 'yearly'>('monthly');
  plan = signal<any | null>(null);
  
  promoCode = signal('');
  promoLoading = signal(false);
  promoError = signal('');
  promoSuccess = signal(false);
  discountType = signal('');
  discountValue = signal(0);
  
  isProcessing = signal(false);

  // Use any to bypass TypeScript typing issues for price_monthly / price_yearly if they are added later
  originalPrice = computed(() => {
    const p = this.plan() as any;
    if (!p) return 0;
    if (this.billingCycle() === 'yearly' && p.price_yearly) return p.price_yearly;
    if (this.billingCycle() === 'monthly' && p.price_monthly) return p.price_monthly;
    return p.price || 0;
  });

  discountAmount = computed(() => {
    if (!this.promoSuccess()) return 0;
    const price = this.originalPrice();
    if (this.discountType() === 'percentage') {
      return price * (this.discountValue() / 100);
    } else if (this.discountType() === 'flat') {
      return this.discountValue();
    }
    return 0;
  });

  finalTotal = computed(() => {
    const total = this.originalPrice() - this.discountAmount();
    return total < 0 ? 0 : total;
  });

  ngOnInit() {
    this.planId.set(this.route.snapshot.paramMap.get('planId'));
    this.billingCycle.set((this.route.snapshot.queryParamMap.get('cycle') as any) || 'monthly');

    if (!this.planId()) {
      this.router.navigate(['/pricing']);
      return;
    }

    this.pricingService.getPublicPricingPlans().subscribe({
      next: (plans) => {
        const found = plans.find(p => String(p.id) === String(this.planId()));
        if (found) {
          this.plan.set(found);
        } else {
          this.router.navigate(['/pricing']);
        }
      },
      error: () => this.router.navigate(['/pricing'])
    });
  }

  applyPromo() {
    if (!this.promoCode().trim()) return;
    
    this.promoLoading.set(true);
    this.promoError.set('');
    this.promoSuccess.set(false);

    this.billingService.validatePromo(this.promoCode()).subscribe({
      next: (res) => {
        if (res.success) {
          this.promoSuccess.set(true);
          this.discountType.set(res.discount_type);
          this.discountValue.set(res.discount);
        } else {
          this.promoError.set(res.message || 'Invalid promo code');
        }
        this.promoLoading.set(false);
      },
      error: (err) => {
        this.promoError.set(err.error?.error || 'Invalid or expired promo code');
        this.promoLoading.set(false);
      }
    });
  }

  removePromo() {
    this.promoCode.set('');
    this.promoSuccess.set(false);
    this.discountType.set('');
    this.discountValue.set(0);
    this.promoError.set('');
  }

  processPayment() {
    if (!this.planId()) return;
    
    this.isProcessing.set(true);
    
    // Pass the actual promo code if it was successfully applied
    const appliedPromo = this.promoSuccess() ? this.promoCode() : undefined;
    
    this.billingService.processPayment(this.planId()!, this.billingCycle(), appliedPromo).subscribe({
      next: (res) => {
        if (res.url) {
          window.location.href = res.url;
        } else {
          // If free plan, it might just return success without URL, or navigate to billing
          this.router.navigate(['/billing']);
        }
      },
      error: async (err) => {
        console.error('Payment failed', err);
        await this.alertService.alert(err.error?.error || 'Failed to initiate payment. Please try again.', 'Payment Failed');
        this.isProcessing.set(false);
      }
    });
  }
}
