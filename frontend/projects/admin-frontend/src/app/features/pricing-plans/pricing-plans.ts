import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AdminPricingService, PricingPlan, PlanFeature } from '../../services/admin-pricing.service';
import { AlertService } from '../../services/alert.service';

@Component({
  selector: 'app-pricing-plans',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './pricing-plans.html',
})
export class PricingPlansComponent implements OnInit {
  private service = inject(AdminPricingService);
  private alert = inject(AlertService);

  plans = signal<PricingPlan[]>([]);
  isLoading = signal(true);
  isModalOpen = signal(false);
  isSaving = signal(false);
  
  editingPlan = signal<PricingPlan | null>(null);
  
  // Form Model
  formPlan = signal<PricingPlan>({
    name: '',
    slug: '',
    description: '',
    price_monthly: 0,
    price_yearly: 0,
    currency: 'USD',
    is_popular: false,
    button_text: 'Get Started',
    order_index: 0,
    max_branches: 1,
    max_staff: 1,
    features: []
  });

  ngOnInit() {
    this.loadPlans();
  }

  loadPlans() {
    this.isLoading.set(true);
    this.service.getPlans().subscribe({
      next: (data) => {
        this.plans.set(data || []);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Error loading plans', err);
        this.isLoading.set(false);
      }
    });
  }

  openCreateModal() {
    this.editingPlan.set(null);
    this.formPlan.set({
      name: '',
      slug: '',
      description: '',
      price_monthly: 0,
      price_yearly: 0,
      currency: 'USD',
      is_popular: false,
      button_text: 'Get Started',
      order_index: this.plans().length,
      max_branches: 1,
      max_staff: 1,
      features: []
    });
    this.isModalOpen.set(true);
  }

  openEditModal(plan: PricingPlan) {
    this.editingPlan.set(plan);
    // Deep clone features to avoid mutating state directly
    this.formPlan.set(JSON.parse(JSON.stringify(plan)));
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  addFeature() {
    const current = this.formPlan();
    current.features.push({
      text: '',
      is_available: true,
      order_index: current.features.length
    });
    this.formPlan.set({...current});
  }

  removeFeature(index: number) {
    const current = this.formPlan();
    current.features.splice(index, 1);
    this.formPlan.set({...current});
  }

  savePlan() {
    this.isSaving.set(true);
    const plan = this.formPlan();
    
    // Auto-generate slug if empty
    if (!plan.slug) {
      plan.slug = plan.name.toLowerCase().replace(/[^a-z0-9]+/g, '-');
    }

    const request = this.editingPlan() 
      ? this.service.updatePlan(plan.id!, plan)
      : this.service.createPlan(plan);

    request.subscribe({
      next: () => {
        this.isSaving.set(false);
        this.closeModal();
        this.loadPlans();
      },
      error: (err) => {
        console.error('Failed to save plan', err);
        this.isSaving.set(false);
      }
    });
  }

  moveLeft(index: number) {
    if (index === 0) return;
    this.swapAndSave(index, index - 1);
  }

  moveRight(index: number) {
    const plans = this.plans();
    if (index === plans.length - 1) return;
    this.swapAndSave(index, index + 1);
  }

  private swapAndSave(idx1: number, idx2: number) {
    const plans = [...this.plans()];
    
    // Swap positions
    const temp = plans[idx1];
    plans[idx1] = plans[idx2];
    plans[idx2] = temp;

    // Update order_index to match new array order
    plans.forEach((p, i) => p.order_index = i);
    this.plans.set(plans);

    // Save the swapped plans to persist the order
    this.service.updatePlan(plans[idx1].id!, plans[idx1]).subscribe();
    this.service.updatePlan(plans[idx2].id!, plans[idx2]).subscribe();
  }

  async deletePlan(id: string) {
    const confirmed = await this.alert.confirm({
      title: 'Delete Pricing Plan',
      message: 'Are you sure you want to delete this pricing plan? This action cannot be undone.',
      confirmText: 'Delete',
      cancelText: 'Cancel',
      type: 'danger'
    });
    if (confirmed) {
      this.service.deletePlan(id).subscribe({
        next: () => { this.alert.success('Pricing plan deleted.'); this.loadPlans(); },
        error: () => this.alert.error('Failed to delete pricing plan.')
      });
    }
  }
}
