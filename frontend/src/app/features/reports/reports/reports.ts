import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AnalyticsService, StaffPerformanceData, SalesGoalProgress, CustomReportResult } from '../../../core/services/analytics.service';
import { ToastService } from '../../../core/services/toast';
import { CrmService } from '../../../core/services/crm.service';
import { RouterLink } from '@angular/router';

type TabId = 'overview' | 'heatmap' | 'categories' | 'cashflow' | 'staff' | 'credit' | 'builder';

@Component({
  selector: 'app-reports',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe, RouterLink],
  templateUrl: './reports.html',
})
export class Reports implements OnInit {
  analyticsService = inject(AnalyticsService);
  crmService = inject(CrmService);
  private toastr = inject(ToastService);

  activeTab = signal<TabId>('overview');
  dateRange = signal({ from: this.defaultFrom(), to: new Date().toISOString().split('T')[0] });

  salesTrends = signal<any>(null);
  revenueBreakdown = signal<any>(null);
  topProducts = signal<{by_revenue?: any[]}>({});
  heatmapData = signal<any[]>([]);

  staffPerformance = signal<StaffPerformanceData[]>([]);
  salesGoal = signal<SalesGoalProgress | null>(null);
  customReport = signal<CustomReportResult | null>(null);

  // Credit & Debt Report State
  overdueAccounts = signal<any[]>([]);
  allCustomersWithDebt = signal<any[]>([]);
  creditReportLoading = signal(false);
  creditReminderSending = signal<string | null>(null);

  creditSummary = computed(() => {
    const customers = this.allCustomersWithDebt();
    const overdue = this.overdueAccounts();
    const totalDebt = customers.reduce((sum: number, c: any) => sum + (c.debt_balance || 0), 0);
    const totalOverdueAmt = overdue.reduce((sum: number, a: any) => sum + Math.abs(a.balance || 0), 0);
    return {
      totalCustomersWithDebt: customers.length,
      totalDebt,
      totalOverdueAccounts: overdue.length,
      totalOverdueAmount: totalOverdueAmt,
    };
  });

  debtorSearch = signal('');
  filteredDebtors = computed(() => {
    const q = this.debtorSearch().toLowerCase().trim();
    const list = this.allCustomersWithDebt();
    if (!q) return list;
    return list.filter((c: any) =>
      c.name?.toLowerCase().includes(q) ||
      c.phone?.toLowerCase().includes(q) ||
      c.email?.toLowerCase().includes(q)
    );
  });

  isLoading = signal(false);
  isBuildingReport = signal(false);

  reportForm = signal({
    metrics: ['revenue'] as string[],
    dimensions: ['date'] as string[],
    from: this.defaultFrom(),
    to: new Date().toISOString().split('T')[0]
  });

  readonly periods = [
    { id: 'today',   label: 'Today' },
    { id: 'week',    label: '7D' },
    { id: 'month',   label: '30D' },
    { id: '3months', label: '90D' },
    { id: 'year',    label: '1Y' },
  ];
  activePeriod = signal('month');

  readonly tabs: { id: TabId; label: string; icon: string }[] = [
    { id: 'overview',   label: 'Overview',       icon: 'bar_chart' },
    { id: 'heatmap',    label: 'Sales Heatmap',  icon: 'grid_on' },
    { id: 'categories', label: 'Categories',     icon: 'pie_chart' },
    { id: 'cashflow',   label: 'Cash Flow',      icon: 'waterfall_chart' },
    { id: 'staff',      label: 'Staff Perf.',    icon: 'badge' },
    { id: 'credit',     label: 'Credit & Debt',  icon: 'credit_score' },
    { id: 'builder',    label: 'Custom Builder', icon: 'tune' },
  ];

  hours = Array.from({ length: 24 }, (_, i) => i);
  days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

  private defaultFrom() {
    const d = new Date();
    d.setMonth(d.getMonth() - 1);
    return d.toISOString().split('T')[0];
  }

  ngOnInit() { this.loadData(); }

  loadData() {
    this.isLoading.set(true);
    const params = { from: this.dateRange().from, to: this.dateRange().to };
    this.analyticsService.getSalesTrends(params).subscribe({ next: r => this.salesTrends.set(r) });
    this.analyticsService.getRevenueBreakdown(params).subscribe({ next: r => this.revenueBreakdown.set(r) });
    this.analyticsService.getTopProducts(params).subscribe({ next: r => this.topProducts.set(r) });
    this.analyticsService.salesHeatmap().subscribe({ next: r => this.heatmapData.set(r || []) });
    this.analyticsService.getStaffPerformance(params).subscribe({ next: r => this.staffPerformance.set(r || []) });
    this.analyticsService.getSalesGoalProgress().subscribe({
      next: r => { this.salesGoal.set(r); this.isLoading.set(false); },
      error: () => this.isLoading.set(false)
    });
  }

  loadCreditReport() {
    this.creditReportLoading.set(true);
    this.crmService.getOverdueAccounts().subscribe({
      next: r => this.overdueAccounts.set(r?.overdue_accounts || []),
      error: () => this.overdueAccounts.set([])
    });
    this.crmService.getCustomers({ limit: 500 }).subscribe({
      next: (res: any) => {
        const list = Array.isArray(res) ? res : (res?.data || []);
        const withDebt = list.filter((c: any) => (c.debt_balance || 0) > 0);
        this.allCustomersWithDebt.set(withDebt);
        this.creditReportLoading.set(false);
      },
      error: () => this.creditReportLoading.set(false)
    });
  }

  setActiveTab(tab: TabId) {
    this.activeTab.set(tab);
    if (tab === 'credit') {
      this.loadCreditReport();
    }
  }

  sendReminder(customerId: string) {
    this.creditReminderSending.set(customerId);
    this.crmService.sendCreditReminder(customerId).subscribe({
      next: () => {
        this.toastr.showSuccess('Repayment reminder sent via SMS');
        this.creditReminderSending.set(null);
      },
      error: () => {
        this.toastr.showError('Failed to send reminder');
        this.creditReminderSending.set(null);
      }
    });
  }

  exportDebtorCSV() {
    const data = this.filteredDebtors();
    if (!data.length) { this.toastr.showError('No debtor data to export.'); return; }
    const csv = [
      ['Customer Name', 'Phone', 'Email', 'Outstanding Debt (GHS)', 'Credit Limit (GHS)'].join(','),
      ...data.map((c: any) => [
        '"' + (c.name || '') + '"', c.phone || '', c.email || '', c.debt_balance || 0, c.credit_limit || 0
      ].join(','))
    ].join('\n');
    this.downloadCSV(csv, 'customer_debt_report');
  }

  setPeriod(pid: string) {
    this.activePeriod.set(pid);
    const to = new Date().toISOString().split('T')[0];
    const from = new Date();
    switch (pid) {
      case 'today':   break;
      case 'week':    from.setDate(from.getDate() - 7); break;
      case 'month':   from.setMonth(from.getMonth() - 1); break;
      case '3months': from.setMonth(from.getMonth() - 3); break;
      case 'year':    from.setFullYear(from.getFullYear() - 1); break;
    }
    this.dateRange.set({ from: from.toISOString().split('T')[0], to });
    this.loadData();
  }

  get dailyMax(): number {
    const data = this.salesTrends()?.daily_data;
    if (!data?.length) return 1;
    return Math.max(...data.map((d: any) => d.revenue || 0)) || 1;
  }
  barHeight(rev: number): number { return Math.max((rev / this.dailyMax) * 100, 2); }

  get heatmapMax(): number {
    if (!this.heatmapData().length) return 1;
    return Math.max(...this.heatmapData().map((d: any) => d.revenue || 0)) || 1;
  }
  getHeatmapDataAt(day: number, hour: number): { intensity: number, revenue: number } {
    const entry = this.heatmapData().find((d: any) => d.day === day && d.hour === hour);
    if (!entry?.revenue) return { intensity: 0, revenue: 0 };
    return { intensity: entry.revenue / this.heatmapMax, revenue: entry.revenue };
  }

  get categoryTotal(): number { return this.revenueBreakdown()?.by_category?.reduce((s: number, c: any) => s + (c.revenue || 0), 0) || 1; }
  get paymentMethodTotal(): number { return this.revenueBreakdown()?.by_payment_method?.reduce((s: number, p: any) => s + (p.revenue || 0), 0) || 1; }
  catPct(rev: number): number { return (rev / this.categoryTotal) * 100; }
  pmPct(rev: number): number  { return (rev / this.paymentMethodTotal) * 100; }

  get growthPositive(): boolean { return (this.salesTrends()?.revenue_growth || 0) >= 0; }
  get avgOrderValue(): number {
    const t = this.salesTrends();
    if (!t?.current_orders) return 0;
    return t.current_revenue / t.current_orders;
  }

  get cashflowRows(): any[] {
    const data = this.salesTrends()?.daily_data;
    if (!data?.length) return [];
    let running = 0;
    return data.map((d: any) => { running += d.revenue || 0; return { ...d, cumulative: running }; });
  }
  get cashflowMax(): number {
    const rows = this.cashflowRows;
    return rows.length ? Math.max(...rows.map((r: any) => r.cumulative)) || 1 : 1;
  }

  get insights(): string[] {
    const msgs = [];
    if (this.growthPositive) {
      msgs.push('Sales are up ' + (this.salesTrends()?.revenue_growth?.toFixed(1) || 0) + '% compared to the previous period.');
    } else {
      msgs.push('Sales are down ' + Math.abs(this.salesTrends()?.revenue_growth || 0).toFixed(1) + '% compared to the previous period.');
    }
    const cat = this.revenueBreakdown()?.by_category?.[0];
    if (cat) msgs.push('Your top category is ' + cat.name + ' with ' + this.catPct(cat.revenue).toFixed(1) + '% of total sales.');
    const staff = this.staffPerformance()?.[0];
    if (staff) msgs.push('Top performing staff member is ' + staff.staff_name + ' (' + staff.orders_count + ' orders).');
    const summary = this.creditSummary();
    if (summary.totalDebt > 0) {
      msgs.push('GHS ' + summary.totalDebt.toFixed(2) + ' outstanding across ' + summary.totalCustomersWithDebt + ' customer credit accounts.');
    }
    return msgs;
  }

  toggleMetric(m: string) {
    const form = this.reportForm();
    if (form.metrics.includes(m)) {
      if (form.metrics.length > 1) this.reportForm.update(f => ({ ...f, metrics: f.metrics.filter(x => x !== m) }));
    } else {
      this.reportForm.update(f => ({ ...f, metrics: [...f.metrics, m] }));
    }
  }

  toggleDimension(d: string) {
    const form = this.reportForm();
    if (form.dimensions.includes(d)) {
      if (form.dimensions.length > 1) this.reportForm.update(f => ({ ...f, dimensions: f.dimensions.filter(x => x !== d) }));
    } else {
      this.reportForm.update(f => ({ ...f, dimensions: [...f.dimensions, d] }));
    }
  }

  generateCustomReport() {
    this.isBuildingReport.set(true);
    this.analyticsService.reportBuilder(this.reportForm()).subscribe({
      next: (res) => {
        this.customReport.set(res);
        this.isBuildingReport.set(false);
        this.toastr.showSuccess('Report generated successfully');
      },
      error: () => {
        this.isBuildingReport.set(false);
        this.toastr.showError('Failed to generate report');
      }
    });
  }

  exportCustomCSV() {
    const report = this.customReport();
    if (!report || !report.rows.length) { this.toastr.showError('No report data to export.'); return; }
    const csv = [
      report.headers.join(','),
      ...report.rows.map(r => report.headers.map(h => r[h] || '').join(','))
    ].join('\n');
    this.downloadCSV(csv, 'custom_report');
  }

  exportTopProductsCSV() {
    const data = this.topProducts()?.by_revenue || [];
    if (!data.length) { this.toastr.showError('No data to export.'); return; }
    const csv = [
      ['Product', 'SKU', 'Quantity Sold', 'Revenue'].join(','),
      ...data.map((p: any) => [p.name, p.sku, p.quantity, p.revenue].join(','))
    ].join('\n');
    this.downloadCSV(csv, 'top_products');
  }

  private downloadCSV(csv: string, filenamePrefix: string) {
    const blob = new Blob([csv], { type: 'text/csv' });
    const url  = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filenamePrefix + '_' + new Date().toISOString().split('T')[0] + '.csv';
    link.click();
    URL.revokeObjectURL(url);
    this.toastr.showSuccess('CSV exported!');
  }
}
