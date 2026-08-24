import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SettingsService, FeatureFlags } from '../../services/settings.service';
import { SecurityService, IPAllowlist } from '../../services/security.service';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './settings.html',
})
export class SettingsComponent implements OnInit {
  private service = inject(SettingsService);
  private securityService = inject(SecurityService);

  flags = signal<FeatureFlags>({
    enable_new_dashboard: false,
    enable_pos_beta: false,
    maintenance_mode: false,
    enable_api_keys: false
  });
  
  isLoading = signal(true);
  isSaving = signal(false);
  activeTab = signal<string>('general');

  // IP Allowlist
  ipList = signal<IPAllowlist[]>([]);
  newIpAddress = signal('');
  newIpDescription = signal('');
  isAddingIp = signal(false);

  ngOnInit() {
    this.loadFlags();
    this.loadIPs();
  }

  loadFlags() {
    this.isLoading.set(true);
    this.service.getFeatureFlags().subscribe({
      next: (data) => {
        this.flags.set(data);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load settings', err);
        this.isLoading.set(false);
      }
    });
  }

  loadIPs() {
    this.securityService.getIPAllowlist().subscribe({
      next: (res) => this.ipList.set(res.data || [])
    });
  }

  setTab(tab: string) {
    this.activeTab.set(tab);
  }

  saveFlags() {
    this.isSaving.set(true);
    this.service.updateFeatureFlags(this.flags()).subscribe({
      next: () => {
        this.isSaving.set(false);
      },
      error: (err) => {
        console.error('Failed to update settings', err);
        this.isSaving.set(false);
      }
    });
  }

  addIP() {
    if (!this.newIpAddress()) return;
    this.isAddingIp.set(true);
    this.securityService.addIP({ ip_address: this.newIpAddress(), description: this.newIpDescription() }).subscribe({
      next: () => {
        this.newIpAddress.set('');
        this.newIpDescription.set('');
        this.isAddingIp.set(false);
        this.loadIPs();
      },
      error: () => this.isAddingIp.set(false)
    });
  }

  removeIP(id: string) {
    this.securityService.removeIP(id).subscribe({
      next: () => this.loadIPs()
    });
  }
}
