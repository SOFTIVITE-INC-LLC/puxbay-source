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
  isUploadingLogo = signal(false);

  get config(): any {
    return this.settingsService.settings() || {};
  }

  get storefrontConfig(): any {
    return this.storefrontService.settings() || {};
  }

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

  onLogoFileSelected(event: any) {
    const file = event.target?.files?.[0];
    if (!file) return;

    this.isUploadingLogo.set(true);
    this.settingsService.uploadImage(file, 'logo').subscribe({
      next: (res) => {
        this.isUploadingLogo.set(false);
        const url = res.url;
        this.config.logo_url = url;

        // Also update storefront settings signal
        this.storefrontService.settings.update(s => s ? { ...s, logo_image: url } : { logo_image: url });

        // Save immediately
        this.saveAllStoreSettings();
        this.toastService.showSuccess('Logo uploaded and applied across the system!');
      },
      error: (err) => {
        this.isUploadingLogo.set(false);
        this.alertService.alert(err.error?.error || 'Failed to upload logo', 'Upload Error', 'danger');
      }
    });
  }

  removeLogo() {
    this.config.logo_url = '';
    this.storefrontService.settings.update(s => s ? { ...s, logo_image: '' } : { logo_image: '' });
    this.saveAllStoreSettings();
    this.toastService.showInfo('Logo removed.');
  }

  saveAllStoreSettings() {
    const s = this.config;
    if (s) {
      if (s.company_name) {
        s.store_name = s.company_name;
        this.storefrontService.settings.update(sf => sf ? { ...sf, store_name: s.company_name, logo_image: s.logo_url } : { store_name: s.company_name, logo_image: s.logo_url });
      }
      this.settingsService.updateSettings(s).subscribe(() => {
        const sf = this.storefrontService.settings();
        if (sf) {
          this.storefrontService.updateSettings(sf).subscribe();
        }
        this.toastService.showSuccess('Store Details Saved Successfully!');
      });
    }
  }

  saveSettings() {
    this.saveAllStoreSettings();
  }

  saveStorefrontSettings() {
    const s = this.storefrontService.settings();
    if (s) {
      this.storefrontService.updateSettings(s).subscribe(() => this.toastService.showSuccess('Storefront Settings Saved!'));
    }
  }
}
