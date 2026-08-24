import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, RouterLinkActive, Router, NavigationEnd } from '@angular/router';
import { filter } from 'rxjs/operators';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, RouterModule, RouterLinkActive],
  templateUrl: './sidebar.html',
})
export class SidebarComponent {
  private authService = inject(AuthService);
  private router = inject(Router);
  
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
        { label: 'Dashboard', route: '/dashboard', icon: 'dashboard' }
      ]
    },
    {
      category: 'Tenant Management',
      items: [
        { label: 'Tenants', route: '/tenants', icon: 'storefront' },
        { label: 'Domains', route: '/domains', icon: 'public' }
      ]
    },
    {
      category: 'Billing & Revenue',
      items: [
        { label: 'Pricing Plans', route: '/pricing-plans', icon: 'payments' },
        { label: 'Subscriptions', route: '/subscriptions', icon: 'card_membership' },
        { label: 'Upcoming Renewals', route: '/renewals', icon: 'event_repeat' },
        { label: 'Invoices', route: '/payments', icon: 'receipt_long' },
        { label: 'Failed Payments', route: '/failed-payments', icon: 'money_off' },
        { label: 'Promo Codes', route: '/promo-codes', icon: 'local_activity' }
      ]
    },
    {
      category: 'Growth & Content',
      items: [
        { label: 'Blog Posts', route: '/blog', icon: 'article' },
        { label: 'Referrals', route: '/referrals', icon: 'group_add' },
        { label: 'Broadcasts', route: '/broadcasts', icon: 'campaign' },
        { label: 'FAQs', route: '/faqs', icon: 'help_center' },
        { label: 'Legal Documents', route: '/legal-documents', icon: 'gavel' }
      ]
    },
    {
      category: 'Operations & Integrations',
      items: [
        { label: 'App Marketplace', route: '/apps', icon: 'extension' },
        { label: 'Webhook Logs', route: '/webhook-logs', icon: 'webhook' },
        { label: 'System Backups', route: '/backups', icon: 'save' },
        { label: 'API Keys', route: '/api-keys', icon: 'key' }
      ]
    },
    {
      category: 'Security & Settings',
      items: [
        { label: 'Audit Logs', route: '/audit-logs', icon: 'policy' },
        { label: 'Telemetry Logs', route: '/telemetry', icon: 'timeline' },
        { label: 'Admin Roles', route: '/admin-roles', icon: 'admin_panel_settings' },
        { label: 'Settings', route: '/settings', icon: 'settings' }
      ]
    }
  ];

  logout() {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
