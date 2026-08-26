import { Routes } from '@angular/router';
import { LayoutComponent } from './layout/layout/layout';
import { authGuard } from './auth.guard';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () => import('./features/login/login').then(m => m.LoginComponent)
  },
  {
    path: '',
    component: LayoutComponent,
    canActivate: [authGuard],
    children: [
      {
        path: '',
        redirectTo: 'dashboard',
        pathMatch: 'full'
      },
      {
        path: 'dashboard',
        loadComponent: () => import('./features/dashboard/dashboard').then(m => m.DashboardComponent)
      },
      {
        path: 'tenants',
        loadComponent: () => import('./features/tenants/tenants').then(m => m.TenantsComponent)
      },
      {
        path: 'pricing-plans',
        loadComponent: () => import('./features/pricing-plans/pricing-plans').then(m => m.PricingPlansComponent)
      },
      {
        path: 'subscriptions',
        loadComponent: () => import('./features/subscriptions/subscriptions').then(m => m.SubscriptionsComponent)
      },
      {
        path: 'promo-codes',
        loadComponent: () => import('./features/promo-codes/promo-codes').then(m => m.PromoCodesComponent)
      },
      {
        path: 'broadcasts',
        loadComponent: () => import('./features/broadcasts/broadcasts').then(m => m.BroadcastsComponent)
      },
      {
        path: 'domains',
        loadComponent: () => import('./features/domains/domains').then(m => m.DomainsComponent)
      },
      {
        path: 'payments',
        loadComponent: () => import('./features/payments/payments').then(m => m.PaymentsComponent)
      },
      {
        path: 'audit-logs',
        loadComponent: () => import('./features/audit-logs/audit-logs').then(m => m.AuditLogsComponent)
      },
      {
        path: 'telemetry',
        loadComponent: () => import('./features/telemetry/telemetry').then(m => m.TelemetryComponent)
      },
      {
        path: 'faqs',
        loadComponent: () => import('./features/faqs/faqs').then(m => m.FaqsComponent)
      },
      {
        path: 'apps',
        loadComponent: () => import('./features/app-marketplace/app-marketplace').then(m => m.AppMarketplaceComponent)
      },
      {
        path: 'backups',
        loadComponent: () => import('./features/backups/backups').then(m => m.BackupsComponent)
      },
      {
        path: 'referrals',
        loadComponent: () => import('./features/referrals/referrals').then(m => m.ReferralsComponent)
      },
      {
        path: 'renewals',
        loadComponent: () => import('./features/renewals/renewals').then(m => m.RenewalsComponent)
      },
      {
        path: 'failed-payments',
        loadComponent: () => import('./features/failed-payments/failed-payments').then(m => m.FailedPaymentsComponent)
      },
      {
        path: 'webhook-logs',
        loadComponent: () => import('./features/webhook-logs/webhook-logs').then(m => m.WebhookLogsComponent)
      },
      {
        path: 'settings',
        loadComponent: () => import('./features/settings/settings').then(m => m.SettingsComponent)
      },
      {
        path: 'admin-roles',
        loadComponent: () => import('./features/admin-roles/admin-roles').then(m => m.AdminRolesComponent)
      },
      {
        path: 'admin-users',
        loadComponent: () => import('./features/admin-users/admin-users').then(m => m.AdminUsersComponent)
      },
      {
        path: 'api-keys',
        loadComponent: () => import('./features/api-keys/api-keys').then(m => m.ApiKeysComponent)
      },
      {
        path: 'legal-documents',
        loadComponent: () => import('./features/legal-documents/legal-documents').then(m => m.LegalDocumentsComponent)
      },
      {
        path: 'blog',
        loadComponent: () => import('./features/blog/blog-list').then(m => m.BlogListComponent)
      },
      {
        path: 'blog/:id',
        loadComponent: () => import('./features/blog/blog-edit').then(m => m.BlogEditComponent)
      },
      {
        path: 'gift-cards',
        loadComponent: () => import('./features/gift-cards/gift-cards').then(m => m.AdminGiftCardsComponent)
      }
    ]
  },
  {
    path: '**',
    redirectTo: 'dashboard'
  }
];
