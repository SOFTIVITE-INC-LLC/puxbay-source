import { ToastService } from '../../../core/services/toast';
import { Component, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { CommonModule, DatePipe } from '@angular/common';
import { SettingsService } from '../../../core/services/settings.service';
import { PaymentMethodService } from '../../../core/services/payment-method.service';
import { AdminService } from '../../../core/services/admin.service';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';
import { AlertService } from '../../../core/services/alert.service';
import { RolesComponent } from '../roles/roles.component';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [CommonModule, FormsModule, RolesComponent, DatePipe],
  templateUrl: './settings.html',
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
export class Settings implements OnInit {
  toastService = inject(ToastService);
  private route = inject(ActivatedRoute);
  settingsService = inject(SettingsService);
  paymentMethodService = inject(PaymentMethodService);
  adminService = inject(AdminService);
  storefrontService = inject(StorefrontSettingsService);
  alertService = inject(AlertService);
  
  activeTab = signal('store');
  domains = signal<any[]>([]);
  newDomainName = signal('');

  ngOnInit() {
    this.route.paramMap.subscribe(params => {
      const tab = params.get('tab');
      if (tab && ['store', 'hardware', 'payments', 'notifications', 'domains', 'roles'].includes(tab)) {
        this.activeTab.set(tab);
      }
    });

    this.settingsService.getSettings().subscribe();
    this.paymentMethodService.getMethods().subscribe();
    this.storefrontService.loadSettings().subscribe();
    this.loadDomains();
  }

  loadDomains() {
    this.adminService.listDomains().subscribe(res => this.domains.set(res || []));
  }

  addDomain() {
    if (!this.newDomainName().trim()) return;
    this.adminService.createDomain({ domain_name: this.newDomainName() }).subscribe(() => {
      this.newDomainName.set('');
      this.loadDomains();
    });
  }

  verifyDomain(id: string) {
    this.adminService.verifyDomain(id).subscribe(() => this.loadDomains());
  }

  setPrimary(id: string) {
    this.adminService.setPrimaryDomain(id).subscribe(() => this.loadDomains());
  }

  async removeDomain(id: string) {
    if (await this.alertService.confirm('Are you sure you want to remove this domain?', 'Remove Domain')) {
      this.adminService.deleteDomain(id).subscribe(() => this.loadDomains());
    }
  }

  saveSettings() {
    const s = this.settingsService.settings();
    if (s) {
      this.settingsService.updateSettings(s).subscribe(() => this.toastService.showSuccess('Settings Saved!'));
    }
  }

  saveStorefrontSettings() {
    const s = this.storefrontService.settings();
    if (s) {
      this.storefrontService.updateSettings(s).subscribe(() => this.toastService.showSuccess('Storefront Settings Saved!'));
    }
  }
}
