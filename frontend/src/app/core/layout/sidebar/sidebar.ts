import { Component, Input, Output, EventEmitter, inject, computed, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { BranchService } from '../../../core/services/branch.service';
import { StorefrontSettingsService } from '../../store/services/storefront-settings.service';
import { SettingsService } from '../../services/settings.service';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './sidebar.html',
  styleUrl: './sidebar.css',
})
export class Sidebar implements OnInit {
  authService = inject(AuthService);
  router = inject(Router);
  branchService = inject(BranchService);
  storefrontSettings = inject(StorefrontSettingsService);
  settingsService = inject(SettingsService);

  @Input() isCollapsed = false;
  @Input() storeName = 'Puxbay OS';

  tenantLogo = computed(() => {
    return this.settingsService.settings()?.logo_url || this.storefrontSettings.settings()?.logo_image || null;
  });

  companyName = computed(() => {
    return this.settingsService.settings()?.company_name || this.storefrontSettings.settings()?.store_name || this.storeName || 'Puxbay';
  });

  userRole = computed(() => {
    return (this.authService.currentUser()?.role || 'admin').toLowerCase();
  });

  ngOnInit() {
    this.storefrontSettings.loadSettings().subscribe();
    this.settingsService.getSettings().subscribe();
  }

  @Output() toggleCollapse = new EventEmitter<void>();

  onToggle() {
    this.toggleCollapse.emit();
  }

  logout() {
    this.authService.logout();
    this.router.navigate(['/login']);
  }

  exitBranch() {
    this.branchService.setActiveBranch(null);
    this.router.navigate(['/branches']);
  }
}
