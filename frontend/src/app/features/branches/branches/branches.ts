import { Component, inject, OnInit, signal } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { BranchService } from '../../../core/services/branch.service';
import { Branch } from '../../../core/models/branch.models';
import { Router } from '@angular/router';
import { AlertService } from '../../../core/services/alert.service';
import { AuthService } from '../../../core/services/auth.service';
import { BillingService } from '../../../core/services/billing.service';
import { SettingsService } from '../../../core/services/settings.service';

@Component({
  selector: 'app-branches',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './branches.html',
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
export class Branches implements OnInit {
  branchService = inject(BranchService);
  private router = inject(Router);
  private alertService = inject(AlertService);
  private authService = inject(AuthService);
  private billingService = inject(BillingService);
  settingsService = inject(SettingsService);

  branches = this.branchService.branches;
  isModalOpen = signal(false);
  isEditing = signal(false);
  activeTab = signal<'general' | 'pos' | 'sync'>('general');
  saving = signal(false);

  networkMetrics = signal<any>(null);
  branchMetrics = signal<Record<string, any>>({});

  currentBranch = signal<Partial<Branch>>({
    status: 'active',
    branch_type: 'retail',
    currency_code: 'GHS',
    currency_symbol: 'GH₵',
    primary_color: '#005b96',
    low_stock_threshold: 10
  });

  ngOnInit() {
    const user = this.authService.currentUser();
    if (user?.branch_id) {
      this.router.navigate(['/dashboard']);
      return;
    }
    this.loadData();
  }

  loadData() {
    this.billingService.getSubscription().subscribe();
    this.branchService.getNetworkMetrics().subscribe(res => this.networkMetrics.set(res));

    this.branchService.getBranches().subscribe(branches => {
      // Fetch metrics for each branch
      branches.forEach(b => {
        if (b.id) {
          this.branchService.getBranchMetrics(b.id).subscribe(metrics => {
            this.branchMetrics.update(map => ({ ...map, [b.id as string]: metrics }));
          });
        }
      });
    });
  }

  openAddModal() {
    const sub = this.billingService.subscription();
    let maxBranches = 1;
    if (sub) {
      if (sub.status === 'trialing') maxBranches = 1;
      else if (sub.plan) maxBranches = sub.plan.max_branches;
      if (sub.custom_branches_count != null) maxBranches = sub.custom_branches_count;
    }
    
    if (this.branches().length >= maxBranches) {
      this.alertService.alert(
        `Your current plan limits you to ${maxBranches} branch(es). Please upgrade your plan in the Billing section to add more.`, 
        'Branch Limit Reached'
      );
      return;
    }

    this.isEditing.set(false);
    this.activeTab.set('general');
    this.currentBranch.set({
      status: 'active',
      branch_type: 'retail',
      currency_code: 'GHS',
      currency_symbol: 'GH₵',
      primary_color: '#005b96',
      low_stock_threshold: 10
    });
    this.isModalOpen.set(true);
  }

  editSettings(branch: Branch) {
    this.isEditing.set(true);
    this.activeTab.set('general');
    this.currentBranch.set({ ...branch });
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  saveBranch() {
    this.saving.set(true);
    const branchData = this.currentBranch() as any;

    if (this.isEditing() && branchData.id) {
      this.branchService.updateBranch(branchData.id, branchData).subscribe({
        next: () => {
          this.saving.set(false);
          this.closeModal();
          this.loadData();
        },
        error: () => this.saving.set(false)
      });
    } else {
      this.branchService.createBranch(branchData).subscribe({
        next: () => {
          this.saving.set(false);
          this.closeModal();
          this.loadData();
        },
        error: () => this.saving.set(false)
      });
    }
  }

  manageBranch(branch: Branch) {
    this.branchService.setActiveBranch(branch);
    this.router.navigate(['/dashboard']);
  }

  viewInventory(branch: Branch) {
    this.branchService.setActiveBranch(branch);
    this.router.navigate(['/inventory/supply-chain']);
  }

  async archiveBranch(branch: Branch) {
    if (await this.alertService.confirm(`Are you sure you want to archive ${branch.name}?`, 'Archive Branch')) {
      if (branch.id) {
        this.branchService.deleteBranch(branch.id).subscribe(() => this.loadData());
      }
    }
  }
}
