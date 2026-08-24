import { Component, inject, OnInit, signal } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AdminService, Plan, Broadcast } from '../../../core/services/admin.service';
import { ToastrService } from 'ngx-toastr';
import { AlertService } from '../../../core/services/alert.service';

@Component({
  selector: 'app-superadmin',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './superadmin.html',
})
export class Superadmin implements OnInit {
  adminService = inject(AdminService);
  private toastr = inject(ToastrService);
  alertService = inject(AlertService);

  activeTab = signal<'tenants' | 'plans' | 'broadcasts' | 'settings'>('tenants');

  isPlanModalOpen = signal(false);
  isBroadcastModalOpen = signal(false);

  newPlan = signal<Partial<Plan>>({ name: '', price: 0, features: [] });
  newBroadcast = signal<Partial<Broadcast>>({ title: '', message: '', type: 'info' });

  featureFlags = signal({
    enable_ai_insights: true,
    beta_kiosk_mode: false,
    new_reporting_engine: true
  });

  ngOnInit() {
    this.adminService.getSystemHealth().subscribe();
    this.loadActiveTabData();
  }

  loadActiveTabData() {
    switch(this.activeTab()) {
      case 'tenants': this.adminService.getTenants().subscribe(); break;
      case 'plans': this.adminService.getPlans().subscribe(); break;
      case 'broadcasts': this.adminService.getBroadcasts().subscribe(); break;
    }
  }

  async suspendTenant(id: string) {
    if (await this.alertService.confirm('Are you sure you want to suspend this tenant?', 'Suspend Tenant')) {
      this.adminService.suspendTenant(id).subscribe({
        next: () => this.toastr.success('Tenant suspended.'),
        error: () => this.toastr.error('Failed to suspend tenant.')
      });
    }
  }

  impersonate(id: string) {
    this.adminService.impersonateTenant(id).subscribe({
      next: (res) => {
        localStorage.setItem('puxbay_token', res.token);
        this.toastr.success('Impersonation active. Redirecting...');
        setTimeout(() => window.location.href = '/', 1000);
      },
      error: () => this.toastr.error('Impersonation failed.')
    });
  }

  savePlan() {
    const p = this.newPlan();
    if (!p.name) return;
    this.adminService.createPlan(p).subscribe({
      next: () => {
        this.toastr.success('Plan created.');
        this.isPlanModalOpen.set(false);
        this.newPlan.set({ name: '', price: 0, features: [] });
      },
      error: () => this.toastr.error('Failed to create plan.')
    });
  }

  addFeatureToPlan(feature: string) {
    if (!feature) return;
    this.newPlan.update(p => ({ ...p, features: [...(p.features || []), feature] }));
  }

  removeFeatureFromPlan(index: number) {
    this.newPlan.update(p => ({ ...p, features: p.features?.filter((_, i) => i !== index) }));
  }

  sendBroadcast() {
    const b = this.newBroadcast();
    if (!b.title || !b.message) return;
    this.adminService.createBroadcast(b).subscribe({
      next: () => {
        this.toastr.success('Broadcast sent!');
        this.isBroadcastModalOpen.set(false);
        this.newBroadcast.set({ title: '', message: '', type: 'info' });
      },
      error: () => this.toastr.error('Failed to send broadcast.')
    });
  }

  saveSettings() {
    this.adminService.updateFeatureFlags(this.featureFlags()).subscribe({
      next: () => this.toastr.success('Global settings saved.'),
      error: () => this.toastr.error('Failed to save settings.')
    });
  }
}
