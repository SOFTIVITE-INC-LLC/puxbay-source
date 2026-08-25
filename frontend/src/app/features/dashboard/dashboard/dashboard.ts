import { Component, inject, OnInit, signal, ChangeDetectionStrategy } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { DashboardService } from '../../../core/services/dashboard.service';
import { BranchService } from '../../../core/services/branch.service';
import { CatalogService } from '../../../core/services/catalog.service';
import { AuthService } from '../../../core/services/auth.service';
import { Router } from '@angular/router';
import { SettingsService } from '../../../core/services/settings.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule, AppCurrencyPipe],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './dashboard.html',
  styles: `
    @keyframes gradientShift {
      0% { background-position: 0% 50%; }
      50% { background-position: 100% 50%; }
      100% { background-position: 0% 50%; }
    }
    .animated-gradient-bg {
      background: linear-gradient(135deg, #6366f1, #a855f7, #ec4899, #f59e0b);
      background-size: 300% 300%;
      animation: gradientShift 8s ease infinite;
    }
    .branch-gradient-bg {
      background: linear-gradient(135deg, #0ea5e9, #6366f1, #8b5cf6);
      background-size: 300% 300%;
      animation: gradientShift 8s ease infinite;
    }
    @keyframes float {
      0%, 100% { transform: translateY(0px); }
      50% { transform: translateY(-6px); }
    }
    .float-anim { animation: float 4s ease-in-out infinite; }
  `,
})
export class Dashboard implements OnInit {
  dashboardService = inject(DashboardService);
  branchService = inject(BranchService);
  catalogService = inject(CatalogService);
  authService = inject(AuthService);
  router = inject(Router);
  settingsService = inject(SettingsService);

  readonly Math = Math;

  get tenantQuickActions() {
    const role = this.authService.currentUser()?.role?.toLowerCase();
    const actions = [
      { label: 'New Sale', icon: 'point_of_sale', route: '/pos', color: 'from-indigo-500 to-purple-600', roles: ['admin', 'manager', 'cashier'] },
      { label: 'Branches', icon: 'store', route: '/branches', color: 'from-sky-500 to-blue-600', roles: ['admin', 'manager'] },
      { label: 'Staff', icon: 'badge', route: '/staff', color: 'from-emerald-500 to-teal-600', roles: ['admin', 'manager'] },
      { label: 'Financial', icon: 'account_balance', route: '/financial', color: 'from-orange-500 to-amber-600', roles: ['admin'] },
      { label: 'Reports', icon: 'analytics', route: '/reports', color: 'from-pink-500 to-rose-600', roles: ['admin', 'manager'] },
      { label: 'Customers', icon: 'group', route: '/customers', color: 'from-violet-500 to-purple-600', roles: ['admin', 'manager', 'cashier'] },
    ];
    return actions.filter(a => a.roles.includes(role || ''));
  }

  get branchQuickActions() {
    const role = this.authService.currentUser()?.role?.toLowerCase();
    const actions = [
      { label: 'New Sale', icon: 'point_of_sale', route: '/pos', color: 'from-indigo-500 to-purple-600', roles: ['admin', 'manager', 'cashier'] },
      { label: 'Orders', icon: 'receipt_long', route: '/orders', color: 'from-sky-500 to-blue-600', roles: ['admin', 'manager', 'cashier'] },
      { label: 'Products', icon: 'inventory_2', route: '/inventory', color: 'from-emerald-500 to-teal-600', roles: ['admin', 'manager'] },
      { label: 'Customers', icon: 'group', route: '/customers', color: 'from-pink-500 to-rose-600', roles: ['admin', 'manager', 'cashier'] },
      { label: 'Purchase Orders', icon: 'contract', route: '/purchase-orders', color: 'from-orange-500 to-amber-600', roles: ['admin', 'manager'] },
      { label: 'Reports', icon: 'insert_chart', route: '/reports', color: 'from-violet-500 to-purple-600', roles: ['admin', 'manager'] },
    ];
    return actions.filter(a => a.roles.includes(role || ''));
  }

  ngOnInit() {
    const user = this.authService.currentUser();
    const userRole = user?.role?.toLowerCase();

    if (userRole === 'sales') {
      this.router.navigate(['/pos']);
      return;
    } else if (userRole === 'supplier') {
      this.router.navigate(['/b2b']);
      return;
    }

    const branch = this.branchService.activeBranch();
    if (branch) {
      this.dashboardService.getBranchMetrics(branch.id).subscribe();
    } else {
      if (userRole !== 'admin' && userRole !== 'superadmin') {
        if (user?.branch_id) {
          // If they have a hardcoded branch, fetch it and stay on dashboard
          this.branchService.getBranch(user.branch_id).subscribe({
            next: (b) => {
              this.branchService.setActiveBranch(b);
              this.dashboardService.getBranchMetrics(b.id).subscribe();
            },
            error: () => {
              this.router.navigate(['/branches']);
            }
          });
          return;
        }
        this.router.navigate(['/branches']);
        return;
      }
      
      this.dashboardService.getMetrics().subscribe();
    }
    this.catalogService.getProducts().subscribe();
  }

  get lowStockProducts() {
    return this.catalogService.products()
      .filter(p => p.track_inventory && (p.current_stock || 0) <= (p.reorder_level || 5))
      .slice(0, 5);
  }

  get topProducts() {
    return this.catalogService.products()
      .filter(p => p.is_active)
      .sort((a, b) => (b.current_stock || 0) - (a.current_stock || 0))
      .slice(0, 5);
  }

  get revenueChartData(): number[] {
    const isGlobal = !this.branchService.activeBranch();
    const data = isGlobal 
      ? this.dashboardService.metrics()?.revenue_chart 
      : this.dashboardService.branchMetrics()?.revenue_chart;
      
    return data || [0, 0, 0, 0, 0, 0, 0];
  }

  get maxRevenue(): number {
    const data = this.revenueChartData;
    const max = Math.max(...data);
    return max > 0 ? max : 100; // prevent division by zero
  }

  get chartLabels(): string[] {
    const labels = [];
    for (let i = 6; i >= 0; i--) {
      const d = new Date();
      d.setDate(d.getDate() - i);
      labels.push(d.toLocaleDateString('en-US', { weekday: 'short' }));
    }
    return labels;
  }

  getBarHeight(value: number): number {
    return (value / this.maxRevenue) * 100;
  }

  // Global branch health gauge bars
  readonly storeHealthMetrics = [
    { label: 'Inventory Health', icon: 'inventory_2', color: 'bg-indigo-500' },
    { label: 'Sales Velocity', icon: 'trending_up', color: 'bg-purple-500' },
    { label: 'Customer Retention', icon: 'favorite', color: 'bg-emerald-500' },
    { label: 'Fulfillment Rate', icon: 'local_shipping', color: 'bg-amber-500' },
  ];
}
