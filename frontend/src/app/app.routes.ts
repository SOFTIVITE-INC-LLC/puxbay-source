import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';
import { permissionGuard } from './core/guards/permission.guard';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () => import('./core/layout/public-layout/public-layout').then(m => m.PublicLayout),
    children: [
      {
        path: '',
        pathMatch: 'full',
        loadComponent: () => import('./features/main/landing/landing').then(m => m.Landing)
      },
      {
        path: 'features',
        loadComponent: () => import('./features/main/features/features').then(m => m.Features)
      },
      {
        path: 'solutions',
        loadComponent: () => import('./features/main/solutions/solutions').then(m => m.Solutions)
      },
      {
        path: 'pricing',
        loadComponent: () => import('./features/main/pricing/pricing').then(m => m.Pricing)
      },
      {
        path: 'product/pos',
        loadComponent: () => import('./features/main/pos-product/pos-product').then(m => m.PosProduct)
      },
      {
        path: 'product/inventory',
        loadComponent: () => import('./features/main/inventory-product/inventory-product').then(m => m.InventoryProduct)
      },
      {
        path: 'product/storefront',
        loadComponent: () => import('./features/main/storefront-product/storefront-product').then(m => m.StorefrontProduct)
      },
      {
        path: 'product/analytics',
        loadComponent: () => import('./features/main/analytics-product/analytics-product').then(m => m.AnalyticsProduct)
      },
      {
        path: 'about',
        loadComponent: () => import('./features/main/about/about').then(m => m.About)
      },
      {
        path: 'careers',
        loadComponent: () => import('./features/main/careers/careers').then(m => m.Careers)
      },
      {
        path: 'blog',
        loadComponent: () => import('./features/main/blog/blog').then(m => m.Blog)
      },
      {
        path: 'contact',
        loadComponent: () => import('./features/main/contact/contact').then(m => m.Contact)
      },
      {
        path: 'privacy-policy',
        loadComponent: () => import('./features/main/privacy/privacy').then(m => m.Privacy)
      },
      {
        path: 'terms',
        loadComponent: () => import('./features/main/terms/terms').then(m => m.Terms)
      },
      {
        path: 'cookie-policy',
        loadComponent: () => import('./features/main/cookies/cookies').then(m => m.Cookies)
      },
      {
        path: 'public/receipts/:token',
        loadComponent: () => import('./features/public-receipt/public-receipt').then(m => m.PublicReceiptComponent)
      },
      {
        path: 'public/feedback/:tenant_id',
        loadComponent: () => import('./features/public-feedback/public-feedback').then(m => m.PublicFeedback)
      }
    ]
  },
  {
    path: 'login',
    loadComponent: () => import('./features/auth/login/login').then(m => m.Login)
  },
  {
    path: 'auth/login',
    loadComponent: () => import('./features/auth/login/login').then(m => m.Login)
  },
  {
    path: 'force-change-password',
    loadComponent: () => import('./features/auth/force-change-password/force-change-password').then(m => m.ForceChangePassword)
  },
  {
    path: 'register',
    loadComponent: () => import('./features/auth/register/register').then(m => m.Register)
  },
  {
    path: 'auth/register',
    loadComponent: () => import('./features/auth/register/register').then(m => m.Register)
  },
  {
    path: 'verify-email',
    loadComponent: () => import('./features/auth/verify-email/verify-email').then(m => m.VerifyEmailComponent)
  },
  {
    path: 'auth/verify-email',
    loadComponent: () => import('./features/auth/verify-email/verify-email').then(m => m.VerifyEmailComponent)
  },
  {
    path: 'pos',
    canActivate: [authGuard],
    loadChildren: () => import('./features/pos/pos.routes').then(m => m.posRoutes)
  },
  {
    path: 'stocktake/portal/:token',
    loadComponent: () => import('./features/inventory/stocktake-portal/stocktake-portal').then(m => m.StocktakePortalComponent)
  },
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () => import('./core/layout/main-layout/main-layout').then(m => m.MainLayout),
    children: [
      {
        path: 'dashboard',
        loadComponent: () => import('./features/dashboard/dashboard/dashboard').then(m => m.Dashboard)
      },
      {
        path: 'orders',
        loadComponent: () => import('./features/orders/orders/orders').then(m => m.Orders)
      },
      {
        path: 'orders/:id',
        loadComponent: () => import('./features/orders/order-detail/order-detail').then(m => m.OrderDetail)
      },
      {
        path: 'customers',
        loadComponent: () => import('./features/customers/customers/customers').then(m => m.Customers)
      },
      {
        path: 'customers/:id',
        loadComponent: () => import('./features/customers/customer-detail/customer-detail').then(m => m.CustomerDetail)
      },
      {
        path: 'crm/helpdesk',
        loadComponent: () => import('./features/crm/helpdesk/helpdesk').then(m => m.Helpdesk)
      },
      {
        path: 'delivery',
        loadComponent: () => import('./features/delivery/delivery-dashboard/delivery-dashboard').then(m => m.DeliveryDashboard)
      },
      {
        path: 'inventory',
        loadComponent: () => import('./features/inventory/inventory/inventory').then(m => m.Inventory)
      },
      {
        path: 'inventory/batches',
        loadComponent: () => import('./features/inventory/batch-tracker/batch-tracker').then(m => m.BatchTracker)
      },
      {
        path: 'inventory/products/:id',
        loadComponent: () => import('./features/inventory/product-detail/product-detail').then(m => m.ProductDetail)
      },
      {
        path: 'inventory/add',
        loadComponent: () => import('./features/inventory/product-form/product-form').then(m => m.ProductForm)
      },
      {
        path: 'inventory/edit/:id',
        loadComponent: () => import('./features/inventory/product-form/product-form').then(m => m.ProductForm)
      },
      {
        path: 'inventory/supply-chain',
        loadComponent: () => import('./features/inventory/supply-chain/supply-chain').then(m => m.SupplyChain)
      },
      {
        path: 'inventory/stocktake/:id',
        loadComponent: () => import('./features/inventory/stocktake-detail/stocktake-detail').then(m => m.StocktakeDetailComponent)
      },
      {
        path: 'inventory/purchase-orders',
        loadComponent: () => import('./features/inventory/purchase-orders/purchase-orders').then(m => m.PurchaseOrders)
      },
      {
        path: 'financial',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'manager', 'superadmin'] },
        loadComponent: () => import('./features/financial/financial/financial').then(m => m.Financial)
      },
      {
        path: 'reports',
        loadComponent: () => import('./features/reports/reports/reports').then(m => m.Reports)
      },
      {
        path: 'branches',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'manager', 'superadmin'] },
        loadComponent: () => import('./features/branches/branches/branches').then(m => m.Branches)
      },
      {
        path: 'suppliers',
        loadComponent: () => import('./features/suppliers/suppliers/suppliers').then(m => m.Suppliers)
      },
      {
        path: 'billing',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'superadmin'] },
        loadComponent: () => import('./features/billing/billing/billing').then(m => m.Billing)
      },
      {
        path: 'checkout/:planId',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'superadmin'] },
        loadComponent: () => import('./features/billing/checkout/checkout').then(m => m.Checkout)
      },
      {
        path: 'staff',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'manager', 'superadmin'] },
        loadComponent: () => import('./features/staff/staff/staff').then(m => m.Staff)
      },
      {
        path: 'fb',
        loadComponent: () => import('./features/fb/fb/fb').then(m => m.Fb)
      },
      {
        path: 'hr',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'manager', 'superadmin'] },
        loadComponent: () => import('./features/hr/hr/hr').then(m => m.Hr)
      },
      {
        path: 'hr/:tab',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'manager', 'superadmin'] },
        loadComponent: () => import('./features/hr/hr/hr').then(m => m.Hr)
      },
      {
        path: 'services',
        loadComponent: () => import('./features/services/services/services').then(m => m.Services)
      },
      {
        path: 'feedback',
        loadComponent: () => import('./features/feedback/feedback/feedback').then(m => m.Feedback)
      },
      {
        path: 'b2b',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'manager', 'superadmin'] },
        loadComponent: () => import('./features/b2b/b2b/b2b').then(m => m.B2b)
      },
      {
        path: 'storefront',
        loadComponent: () => import('./features/storefront/storefront/storefront').then(m => m.Storefront)
      },
      {
        path: 'shop/:slug',
        loadComponent: () => import('./features/storefront/public-storefront/public-storefront').then(m => m.PublicStorefront)
      },
      {
        path: 'intelligence',
        loadComponent: () => import('./features/intelligence/intelligence/intelligence').then(m => m.Intelligence)
      },
      {
        path: 'marketing',
        loadComponent: () => import('./features/marketing/marketing/marketing').then(m => m.Marketing)
      },
      {
        path: 'integration',
        loadComponent: () => import('./features/integration/integration/integration').then(m => m.Integration)
      },
      {
        path: 'settings',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'superadmin'] },
        loadComponent: () => import('./features/settings/settings/settings').then(m => m.Settings)
      },
      {
        path: 'settings/:tab',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'superadmin'] },
        loadComponent: () => import('./features/settings/settings/settings').then(m => m.Settings)
      },
      {
        path: 'enterprise/:tab',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'superadmin'] },
        loadComponent: () => import('./features/enterprise/enterprise/enterprise').then(m => m.Enterprise)
      },
      {
        path: 'enterprise',
        redirectTo: 'enterprise/command-center',
        pathMatch: 'full'
      },
      {
        path: 'branch-settings',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'manager', 'superadmin'] },
        loadComponent: () => import('./features/branch-settings/branch-settings/branch-settings').then(m => m.BranchSettings)
      },
      {
        path: 'terminal',
        loadComponent: () => import('./features/terminal/terminal/terminal').then(m => m.Terminal)
      },
      {
        path: 'wallet',
        loadComponent: () => import('./features/wallet/wallet/wallet').then(m => m.Wallet)
      },
      {
        path: 'returns',
        loadComponent: () => import('./features/returns/returns/returns').then(m => m.Returns)
      },
      {
        path: 'notifications',
        loadComponent: () => import('./features/notifications/notifications/notifications').then(m => m.Notifications)
      },
      {
        path: 'security',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'superadmin'] },
        loadComponent: () => import('./features/security/security/security').then(m => m.Security)
      },
      {
        path: 'webhooks',
        canActivate: [permissionGuard],
        data: { roles: ['admin', 'superadmin'] },
        loadComponent: () => import('./features/webhooks/webhooks/webhooks').then(m => m.Webhooks)
      },
      {
        path: 'content',
        loadComponent: () => import('./features/content/content/content').then(m => m.Content)
      },
      {
        path: 'profile',
        loadComponent: () => import('./features/profile/profile/profile').then(m => m.Profile)
      },
      {
        path: 'categories',
        loadComponent: () => import('./features/categories/categories/categories').then(m => m.Categories)
      },
      {
        path: 'privacy',
        loadComponent: () => import('./features/privacy/privacy/privacy').then(m => m.Privacy)
      },
      {
        path: 'superadmin',
        canActivate: [permissionGuard],
        data: { roles: ['superadmin'] },
        loadComponent: () => import('./features/superadmin/superadmin/superadmin').then(m => m.Superadmin)
      },
      {
        path: 'cash-drawers',
        loadComponent: () => import('./features/cash-drawers/cash-drawers/cash-drawers').then(m => m.CashDrawers)
      },
      {
        path: 'gift-cards',
        loadComponent: () => import('./features/gift-cards/gift-cards/gift-cards').then(m => m.GiftCards)
      },
      {
        path: 'payment-methods',
        loadComponent: () => import('./features/payment-methods/payment-methods/payment-methods').then(m => m.PaymentMethods)
      },
      {
        path: '',
        redirectTo: 'dashboard',
        pathMatch: 'full'
      }
    ]
  },
  {
    path: 'portal/:slug',
    loadComponent: () => import('./features/portal/portal/portal').then(m => m.Portal)
  },
  {
    path: 'supplier-portal',
    children: [
      {
        path: '',
        redirectTo: 'login',
        pathMatch: 'full'
      },
      {
        path: 'login',
        loadComponent: () => import('./features/supplier-portal/pages/login/login.component').then(m => m.SupplierPortalLoginComponent)
      },
      {
        path: 'orders',
        loadComponent: () => import('./features/supplier-portal/pages/orders/orders.component').then(m => m.SupplierPortalOrdersComponent)
      }
    ]
  },

  {
    path: 'kiosk',
    loadComponent: () => import('./features/kiosk/kiosk/kiosk').then(m => m.Kiosk)
  },
  {
    path: 'kiosk/:branchId',
    loadComponent: () => import('./features/kiosk/kiosk/kiosk').then(m => m.Kiosk)
  },
  {
    path: 'cds',
    loadComponent: () => import('./features/pos/cds/cds').then(m => m.Cds)
  },
  {
    path: 'store',
    loadChildren: () => import('./features/store/store.routes').then(m => m.storeRoutes)
  },
  {
    path: '**',
    loadComponent: () => import('./core/components/not-found.component').then(m => m.NotFoundComponent)
  }
];
