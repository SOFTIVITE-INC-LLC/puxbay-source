import { ToastService } from '../../../core/services/toast';
import { Component, inject, OnInit, computed } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { StorefrontService } from '../../../core/services/storefront.service';
import { SettingsService } from '../../../core/services/settings.service';
import { TenantStore } from '../../../core/services/tenant.store';

@Component({
  selector: 'app-storefront',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './storefront.html',
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
export class Storefront implements OnInit {
  toastService = inject(ToastService);
  storefrontService = inject(StorefrontService);
  settingsService = inject(SettingsService);
  tenantStore = inject(TenantStore);

  subdomain = computed(() => this.tenantStore.subdomain() || 'yourstore');

  ngOnInit() {
    this.storefrontService.getSettings().subscribe();
  }

  saveSettings() {
    const s = this.storefrontService.settings();
    if (s) {
      this.storefrontService.updateSettings(s).subscribe(() => {
        this.toastService.showSuccess('Storefront settings updated successfully.');
      });
    }
  }
}
