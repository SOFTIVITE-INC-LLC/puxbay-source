import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, RouterLinkActive, Router, NavigationEnd } from '@angular/router';
import { filter } from 'rxjs/operators';
import { AuthService } from '../../services/auth.service';
import { LayoutService } from '../layout.service';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, RouterModule, RouterLinkActive],
  templateUrl: './sidebar.html',
})
export class SidebarComponent {
  private authService = inject(AuthService);
  private router = inject(Router);
  public layout = inject(LayoutService);
  
  currentRoute = signal<string>('');

  constructor() {
    this.router.events.pipe(
      filter(event => event instanceof NavigationEnd)
    ).subscribe((event: any) => {
      this.currentRoute.set(event.urlAfterRedirects.split('?')[0]);
    });
  }

  menuCategories = [
    {
      category: 'Overview',
      items: [
        { label: 'Dashboard', route: '/dashboard', icon: 'dashboard', permission: 'dashboard:read' }
      ]
    },
    {
      category: 'Tenant Management',
      items: [
        { label: 'Tenants', route: '/tenants', icon: 'storefront', permission: 'tenants:read' },
        { label: 'Domains', route: '/domains', icon: 'public', permission: 'domains:read' }
      ]
    },
    {
      category: 'Billing & Revenue',
      items: [
        { label: 'Pricing Plans', route: '/pricing-plans', icon: 'payments', permission: 'pricing_plans:read' },
        { label: 'Subscriptions', route: '/subscriptions', icon: 'card_membership', permission: 'billing:read' },
        { label: 'Upcoming Renewals', route: '/renewals', icon: 'event_repeat', permission: 'billing:read' },
        { label: 'Invoices', route: '/payments', icon: 'receipt_long', permission: 'billing:read' },
        { label: 'Failed Payments', route: '/failed-payments', icon: 'money_off', permission: 'billing:read' },
        { label: 'Promo Codes', route: '/promo-codes', icon: 'local_activity', permission: 'promo_codes:read' }
      ]
    },
    {
      category: 'Growth & Content',
      items: [
        { label: 'Gift Cards', route: '/gift-cards', icon: 'card_giftcard', permission: 'billing:read' },
        { label: 'Blog Posts', route: '/blog', icon: 'article', permission: 'content:read' },
        { label: 'Referrals', route: '/referrals', icon: 'group_add', permission: 'referrals:read' },
        { label: 'Broadcasts', route: '/broadcasts', icon: 'campaign', permission: 'broadcasts:read' },
        { label: 'FAQs', route: '/faqs', icon: 'help_center', permission: 'content:read' },
        { label: 'Legal Documents', route: '/legal-documents', icon: 'gavel', permission: 'content:read' }
      ]
    },
    {
      category: 'Operations & Integrations',
      items: [
        { label: 'App Marketplace', route: '/apps', icon: 'extension', permission: 'apps:read' },
        { label: 'Webhook Logs', route: '/webhook-logs', icon: 'webhook', permission: 'webhooks:read' },
        { label: 'System Backups', route: '/backups', icon: 'save', permission: 'backups:read' },
        { label: 'API Keys', route: '/api-keys', icon: 'key', permission: 'api_keys:read' }
      ]
    },
    {
      category: 'Security & Settings',
      items: [
        { label: 'Audit Logs', route: '/audit-logs', icon: 'policy', permission: 'security:read' },
        { label: 'Telemetry Logs', route: '/telemetry', icon: 'timeline', permission: 'security:read' },
        { label: 'Admin Roles', route: '/admin-roles', icon: 'admin_panel_settings', permission: 'admin_users:read' },
        { label: 'Admin Users', route: '/admin-users', icon: 'manage_accounts', permission: 'admin_users:read' },
        { label: 'Settings', route: '/settings', icon: 'settings', permission: 'settings:write' }
      ]
    }
  ];

  get filteredCategories() {
    const user = this.authService.currentUser();

    // 1. Superusers, superadmins, or while user is loading: show all menus
    if (!user || user.is_superuser || user.role === 'superadmin') {
      return this.menuCategories;
    }

    // 2. Wildcard admin role: show all menus
    if (user.permissions && user.permissions.includes('*')) {
      return this.menuCategories;
    }

    // 3. Restricted staff: filter by assigned permissions
    return this.menuCategories.map(cat => ({
      ...cat,
      items: cat.items.filter(item => {
        if (!item.permission) return true;
        return this.authService.hasPermission(item.permission);
      })
    })).filter(cat => cat.items.length > 0);
  }

  logout() {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
