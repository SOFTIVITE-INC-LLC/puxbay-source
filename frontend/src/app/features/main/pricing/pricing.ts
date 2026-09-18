import { Component, ViewEncapsulation, signal, computed, inject, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { PricingService, PricingPlan } from '../../../core/services/pricing.service';
import { AuthService } from '../../../core/services/auth.service';
import { BillingService } from '../../../core/services/billing.service';
import { Router } from '@angular/router';
import { SeoService } from '../../../core/services/seo.service';
@Component({
  selector: 'app-pricing',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './pricing.html',
  encapsulation: ViewEncapsulation.None,
})
export class Pricing implements OnInit, OnDestroy {
  private pricingService = inject(PricingService);
  private auth = inject(AuthService);
  private billingService = inject(BillingService);
  private router = inject(Router);
  private seo = inject(SeoService);
  
  billingCycle = signal<'monthly' | 'yearly'>('monthly');
  plans = signal<PricingPlan[]>([]);
  isLoading = signal(true);

  ngOnInit() {
    this.seo.setPageSeo({
      title: 'Pricing & Plans | Puxbay',
      description: 'Transparent, flexible pricing for businesses of every size. Start free or choose a plan that scales with your growth.',
      keywords: 'Puxbay pricing, POS pricing, inventory management pricing, e-commerce plans, business software pricing',
    });

    this.seo.setJsonLd({
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      'name': 'Puxbay',
      'applicationCategory': 'BusinessApplication',
      'operatingSystem': 'Web, iOS, Android',
      'url': 'https://puxbay.com/pricing',
      'description': 'All-in-one POS, inventory management, and e-commerce platform with transparent, flexible pricing.',
      'offers': {
        '@type': 'AggregateOffer',
        'priceCurrency': 'USD',
        'lowPrice': '0',
        'offerCount': '4',
        'description': 'Free plan available. Paid plans scale with your business.'
      }
    });

    this.seo.setBreadcrumbJsonLd([
      { name: 'Home', url: 'https://puxbay.com/' },
      { name: 'Pricing', url: 'https://puxbay.com/pricing' }
    ]);

    this.pricingService.getPublicPricingPlans().subscribe({
      next: (plans) => {
        this.plans.set(plans);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load pricing plans', err);
        this.isLoading.set(false);
      }
    });
  }

  ngOnDestroy() {
    this.seo.removeJsonLd();
  }

  selectPlan(plan: PricingPlan) {
    if (this.auth.isAuthenticated()) {
      this.router.navigate(['/checkout', plan.id], { queryParams: { cycle: this.billingCycle() } });
    } else {
      this.router.navigate(['/register']);
    }
  }

  toggleBilling() {
    this.billingCycle.update(cycle => cycle === 'monthly' ? 'yearly' : 'monthly');
  }

  // ROI Calculator State
  storeCount = signal(1);
  registerCount = signal(3);

  estimatedSavings = computed(() => {
    // Basic mock formula: (Registers * $120 legacy cost) + (Stores * $500 manual inventory loss) 
    const legacyHardwareCost = this.registerCount() * 120;
    const inventoryLoss = this.storeCount() * 500;
    
    // Attempt to find a plan for ROI base or default
    const currentPlans = this.plans();
    let puxbayCost = 0;
    if (currentPlans.length > 0) {
      const proPlan = currentPlans.find(p => p.name.toLowerCase().includes('pro')) || currentPlans[0];
      puxbayCost = this.billingCycle() === 'monthly' ? proPlan.price_monthly : proPlan.price_yearly;
    } else {
      puxbayCost = this.billingCycle() === 'monthly' ? 499 : 399; // fallback mock
    }
    
    const savings = (legacyHardwareCost + inventoryLoss) - puxbayCost;
    return Math.max(0, savings);
  });
}

