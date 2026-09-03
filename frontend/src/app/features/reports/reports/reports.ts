import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AnalyticsService, StaffPerformanceData, SalesGoalProgress, CustomReportResult } from '../../../core/services/analytics.service';
import { ToastService } from '../../../core/services/toast';
import { CrmService } from '../../../core/services/crm.service';
import { ExportService } from '../../../core/services/export.service';
import { RouterLink } from '@angular/router';

type TabId = 'overview' | 'heatmap' | 'categories' | 'cashflow' | 'staff' | 'credit' | 'builder';

export interface DimensionOption {
  id: string;
  label: string;
  icon: string;
  category: 'time' | 'sales' | 'people' | 'inventory';
}

export interface MetricOption {
  id: string;
  label: string;
  icon: string;
  category: 'financial' | 'activity' | 'deductions';
  isCurrency?: boolean;
}

@Component({
  selector: 'app-reports',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe, RouterLink],
  templateUrl: './reports.html',
})
export class Reports implements OnInit {
  analyticsService = inject(AnalyticsService);
  crmService = inject(CrmService);
  exportService = inject(ExportService);
  private toastr = inject(ToastService);

  activeTab = signal<TabId>('overview');
  dateRange = signal({ from: this.defaultFrom(), to: new Date().toISOString().split('T')[0] });

  salesTrends = signal<any>(null);
  revenueBreakdown = signal<any>(null);
  topProducts = signal<{by_revenue?: any[], by_quantity?: any[]}>({});
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

  // ── Custom Report Builder Config ──
  readonly dimensionOptions: DimensionOption[] = [
    // Time
    { id: 'date', label: 'Date (Daily)', icon: 'calendar_today', category: 'time' },
    { id: 'month', label: 'Month (Monthly)', icon: 'date_range', category: 'time' },
    { id: 'day_of_week', label: 'Day of Week', icon: 'view_week', category: 'time' },
    { id: 'hour', label: 'Hour of Day', icon: 'schedule', category: 'time' },
    // Sales Channels & Status
    { id: 'payment_method', label: 'Payment Method', icon: 'payments', category: 'sales' },
    { id: 'order_type', label: 'Sales Channel', icon: 'storefront', category: 'sales' },
    { id: 'payment_status', label: 'Payment Status', icon: 'check_circle', category: 'sales' },
    { id: 'status', label: 'Order Status', icon: 'sync', category: 'sales' },
    // People
    { id: 'staff', label: 'Staff / Cashier', icon: 'badge', category: 'people' },
    { id: 'customer', label: 'Customer Name', icon: 'person', category: 'people' },
    // Inventory
    { id: 'category', label: 'Product Category', icon: 'category', category: 'inventory' },
    { id: 'product', label: 'Product Name', icon: 'inventory_2', category: 'inventory' },
  ];

  readonly metricOptions: MetricOption[] = [
    // Financial
    { id: 'revenue', label: 'Gross Revenue', icon: 'attach_money', category: 'financial', isCurrency: true },
    { id: 'net_sales', label: 'Net Sales', icon: 'account_balance_wallet', category: 'financial', isCurrency: true },
    { id: 'subtotal', label: 'Subtotal', icon: 'receipt', category: 'financial', isCurrency: true },
    { id: 'amount_paid', label: 'Amount Paid', icon: 'paid', category: 'financial', isCurrency: true },
    // Activity & Volume
    { id: 'orders', label: 'Order Count', icon: 'receipt_long', category: 'activity', isCurrency: false },
    { id: 'items_sold', label: 'Items Sold', icon: 'shopping_bag', category: 'activity', isCurrency: false },
    { id: 'avg_order_value', label: 'Avg Order Value (AOV)', icon: 'trending_up', category: 'activity', isCurrency: true },
    // Deductions & Taxes
    { id: 'discounts', label: 'Discounts Given', icon: 'loyalty', category: 'deductions', isCurrency: true },
    { id: 'tax', label: 'Tax Collected', icon: 'account_balance', category: 'deductions', isCurrency: true },
  ];

  reportForm = signal({
    metrics: ['revenue', 'orders', 'discounts'] as string[],
    dimensions: ['date'] as string[],
    from: this.defaultFrom(),
    to: new Date().toISOString().split('T')[0],
    status: 'completed',
    order_type: 'all',
    payment_method: 'all'
  });

  // Custom Report UX
  customReportSearch = signal('');
  customReportSort = signal<{ column: string; asc: boolean }>({ column: '', asc: true });
  activeBuilderCategory = signal<'all' | 'time' | 'sales' | 'people' | 'inventory'>('all');

  filteredCustomReportRows = computed(() => {
    const report = this.customReport();
    if (!report?.rows?.length) return [];
    let rows = [...report.rows];

    // Search query
    const q = this.customReportSearch().toLowerCase().trim();
    if (q) {
      rows = rows.filter(row =>
        Object.values(row).some(v => String(v).toLowerCase().includes(q))
      );
    }

    // Sort column
    const sort = this.customReportSort();
    if (sort.column) {
      rows.sort((a, b) => {
        let valA = a[sort.column];
        let valB = b[sort.column];
        if (typeof valA === 'number' && typeof valB === 'number') {
          return sort.asc ? valA - valB : valB - valA;
        }
        valA = String(valA || '').toLowerCase();
        valB = String(valB || '').toLowerCase();
        return sort.asc ? valA.localeCompare(valB) : valB.localeCompare(valA);
      });
    }

    return rows;
  });

  customReportKpis = computed(() => {
    const report = this.customReport();
    if (!report?.rows?.length) return null;
    const rows = report.rows;
    let totalRevenue = 0;
    let totalOrders = 0;
    let totalDiscounts = 0;
    let totalTax = 0;
    let totalItems = 0;

    for (const r of rows) {
      totalRevenue += Number(r['revenue'] || 0);
      totalOrders += Number(r['orders'] || 0);
      totalDiscounts += Number(r['discounts'] || 0);
      totalTax += Number(r['tax'] || 0);
      totalItems += Number(r['items_sold'] || 0);
    }

    return {
      rowCount: rows.length,
      totalRevenue,
      totalOrders,
      totalDiscounts,
      totalTax,
      totalItems,
      aov: totalOrders > 0 ? totalRevenue / totalOrders : 0
    };
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

  // ── Dimension & Metric Selection ──
  toggleMetric(m: string) {
    const form = this.reportForm();
    if (form.metrics.includes(m)) {
      if (form.metrics.length > 1) {
        this.reportForm.update(f => ({ ...f, metrics: f.metrics.filter(x => x !== m) }));
      } else {
        this.toastr.showWarning('At least one metric must be selected.');
      }
    } else {
      this.reportForm.update(f => ({ ...f, metrics: [...f.metrics, m] }));
    }
  }

  toggleDimension(d: string) {
    const form = this.reportForm();
    if (form.dimensions.includes(d)) {
      if (form.dimensions.length > 1) {
        this.reportForm.update(f => ({ ...f, dimensions: f.dimensions.filter(x => x !== d) }));
      } else {
        this.toastr.showWarning('At least one dimension must be selected.');
      }
    } else {
      this.reportForm.update(f => ({ ...f, dimensions: [...f.dimensions, d] }));
    }
  }

  selectAllDimensions() {
    this.reportForm.update(f => ({ ...f, dimensions: ['date', 'payment_method', 'order_type', 'staff', 'category'] }));
  }

  selectAllMetrics() {
    this.reportForm.update(f => ({ ...f, metrics: ['revenue', 'orders', 'items_sold', 'discounts', 'tax', 'net_sales', 'avg_order_value'] }));
  }

  resetBuilder() {
    this.reportForm.set({
      metrics: ['revenue', 'orders'],
      dimensions: ['date'],
      from: this.dateRange().from,
      to: this.dateRange().to,
      status: 'completed',
      order_type: 'all',
      payment_method: 'all'
    });
  }

  sortCustomColumn(col: string) {
    const cur = this.customReportSort();
    if (cur.column === col) {
      this.customReportSort.set({ column: col, asc: !cur.asc });
    } else {
      this.customReportSort.set({ column: col, asc: true });
    }
  }

  generateCustomReport() {
    this.isBuildingReport.set(true);
    const payload = {
      ...this.reportForm(),
      from: this.dateRange().from,
      to: this.dateRange().to
    };

    this.analyticsService.reportBuilder(payload).subscribe({
      next: (res) => {
        this.customReport.set(res);
        this.isBuildingReport.set(false);
        this.toastr.showSuccess(`Custom report generated (${res.rows?.length || 0} rows)`);
      },
      error: () => {
        this.isBuildingReport.set(false);
        this.toastr.showError('Failed to generate report');
      }
    });
  }

  // ════════════════════════════════════════════════════════════════
  // ── UNIVERSAL EXPORT HANDLERS (EXCEL, PDF, CSV) ───────────────
  // ════════════════════════════════════════════════════════════════

  exportActiveTab(format: 'excel' | 'pdf' | 'csv') {
    const tab = this.activeTab();
    switch (tab) {
      case 'builder':
        this.exportCustomReport(format);
        break;
      case 'overview':
      case 'heatmap':
        this.exportOverviewReport(format);
        break;
      case 'categories':
        this.exportCategoryReport(format);
        break;
      case 'cashflow':
        this.exportCashflowReport(format);
        break;
      case 'staff':
        this.exportStaffReport(format);
        break;
      case 'credit':
        this.exportCreditReport(format);
        break;
    }
  }

  // ── Custom Report Export ──
  exportCustomReport(format: 'excel' | 'pdf' | 'csv') {
    const report = this.customReport();
    if (!report || !report.rows?.length) {
      this.toastr.showError('Please generate a report first before exporting.');
      return;
    }

    const headers = report.headers;
    const rows = report.rows.map(r => headers.map(h => r[h] !== undefined ? r[h] : ''));
    const kpis = this.customReportKpis();

    const summaryCards = kpis ? [
      { label: 'Total Revenue', value: 'GHS ' + kpis.totalRevenue.toFixed(2), highlight: true },
      { label: 'Total Orders', value: kpis.totalOrders },
      { label: 'Total Items Sold', value: kpis.totalItems },
      { label: 'Avg Order Value', value: 'GHS ' + kpis.aov.toFixed(2) },
      { label: 'Total Discounts', value: 'GHS ' + kpis.totalDiscounts.toFixed(2) },
      { label: 'Total Tax', value: 'GHS ' + kpis.totalTax.toFixed(2) },
    ] : [];

    const options = {
      filename: 'custom_report',
      title: 'Custom Analytics Report',
      subtitle: `Dimensions: ${this.reportForm().dimensions.join(', ')} | Metrics: ${this.reportForm().metrics.join(', ')}`,
      dateRange: `${this.dateRange().from} to ${this.dateRange().to}`,
      headers,
      rows,
      summaryCards,
      orientation: headers.length > 5 ? ('landscape' as const) : ('portrait' as const)
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastr.showSuccess('Report exported to Excel (.xls)!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastr.showSuccess('Report exported to CSV!');
    }
  }

  // ── Overview / Sales Trends Export ──
  exportOverviewReport(format: 'excel' | 'pdf' | 'csv') {
    const trends = this.salesTrends();
    const daily = trends?.daily_data || [];
    if (!daily.length) {
      this.toastr.showError('No sales trend data to export.');
      return;
    }

    const headers = ['Date', 'Revenue (GHS)', 'Orders', 'Avg Order Value (GHS)'];
    const rows = daily.map((d: any) => [
      d.date,
      Number(d.revenue || 0).toFixed(2),
      d.orders || 0,
      (d.orders > 0 ? (d.revenue / d.orders) : 0).toFixed(2)
    ]);

    const options = {
      filename: 'sales_overview_report',
      title: 'Sales & Revenue Overview',
      subtitle: `Period: ${this.dateRange().from} to ${this.dateRange().to}`,
      dateRange: `${this.dateRange().from} to ${this.dateRange().to}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Gross Revenue', value: 'GHS ' + (trends?.current_revenue || 0).toFixed(2), highlight: true },
        { label: 'Total Orders', value: trends?.current_orders || 0 },
        { label: 'Revenue Growth', value: (trends?.revenue_growth || 0).toFixed(1) + '%' },
        { label: 'Order Growth', value: (trends?.order_growth || 0).toFixed(1) + '%' },
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastr.showSuccess('Sales Overview exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastr.showSuccess('Sales Overview exported to CSV!');
    }
  }

  // ── Category Breakdown Export ──
  exportCategoryReport(format: 'excel' | 'pdf' | 'csv') {
    const catData = this.revenueBreakdown()?.by_category || [];
    if (!catData.length) {
      this.toastr.showError('No category data to export.');
      return;
    }

    const headers = ['Category', 'Revenue (GHS)', 'Contribution (%)'];
    const rows = catData.map((c: any) => [
      c.name,
      Number(c.revenue || 0).toFixed(2),
      this.catPct(c.revenue).toFixed(1) + '%'
    ]);

    const options = {
      filename: 'category_breakdown',
      title: 'Revenue by Product Category',
      dateRange: `${this.dateRange().from} to ${this.dateRange().to}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Total Category Sales', value: 'GHS ' + this.categoryTotal.toFixed(2), highlight: true },
        { label: 'Top Category', value: catData[0]?.name || 'N/A' },
        { label: 'Categories Count', value: catData.length },
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastr.showSuccess('Category report exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastr.showSuccess('Category report exported to CSV!');
    }
  }

  // ── Cash Flow Export ──
  exportCashflowReport(format: 'excel' | 'pdf' | 'csv') {
    const rowsData = this.cashflowRows;
    if (!rowsData.length) {
      this.toastr.showError('No cash flow data to export.');
      return;
    }

    const headers = ['Date', 'Daily Inflow (GHS)', 'Orders', 'Cumulative Revenue (GHS)'];
    const rows = rowsData.map((d: any) => [
      d.date,
      Number(d.revenue || 0).toFixed(2),
      d.orders || 0,
      Number(d.cumulative || 0).toFixed(2)
    ]);

    const options = {
      filename: 'cash_flow_report',
      title: 'Cumulative Cash Flow Statement',
      dateRange: `${this.dateRange().from} to ${this.dateRange().to}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Cumulative Inflow', value: 'GHS ' + this.cashflowMax.toFixed(2), highlight: true },
        { label: 'Total Days Tracked', value: rowsData.length },
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastr.showSuccess('Cash flow report exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastr.showSuccess('Cash flow report exported to CSV!');
    }
  }

  // ── Staff Performance Export ──
  exportStaffReport(format: 'excel' | 'pdf' | 'csv') {
    const staff = this.staffPerformance();
    if (!staff.length) {
      this.toastr.showError('No staff performance data to export.');
      return;
    }

    const headers = ['Staff Member', 'Orders Completed', 'Total Revenue (GHS)', 'Avg Sale / Order (GHS)'];
    const rows = staff.map((s: any) => [
      s.staff_name,
      s.orders_count || 0,
      Number(s.revenue || 0).toFixed(2),
      (s.orders_count > 0 ? (s.revenue / s.orders_count) : 0).toFixed(2)
    ]);

    const totalStaffRev = staff.reduce((acc, s) => acc + (s.revenue || 0), 0);
    const totalStaffOrders = staff.reduce((acc, s) => acc + (s.orders_count || 0), 0);

    const options = {
      filename: 'staff_performance',
      title: 'Staff Performance & Sales Commission Report',
      dateRange: `${this.dateRange().from} to ${this.dateRange().to}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Total Staff Revenue', value: 'GHS ' + totalStaffRev.toFixed(2), highlight: true },
        { label: 'Total Staff Orders', value: totalStaffOrders },
        { label: 'Active Staff Count', value: staff.length },
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastr.showSuccess('Staff report exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastr.showSuccess('Staff report exported to CSV!');
    }
  }

  // ── Credit & Debtors Export ──
  exportCreditReport(format: 'excel' | 'pdf' | 'csv') {
    const debtors = this.filteredDebtors();
    if (!debtors.length) {
      this.toastr.showError('No customer credit data to export.');
      return;
    }

    const summary = this.creditSummary();
    const headers = ['Customer Name', 'Phone', 'Email', 'Debt Balance (GHS)', 'Credit Limit (GHS)', 'Utilization (%)'];
    const rows = debtors.map((c: any) => {
      const debt = Number(c.debt_balance || 0);
      const limit = Number(c.credit_limit || 0);
      const util = limit > 0 ? ((debt / limit) * 100).toFixed(1) + '%' : 'N/A';
      return [c.name || 'Unknown', c.phone || '–', c.email || '–', debt.toFixed(2), limit.toFixed(2), util];
    });

    const options = {
      filename: 'customer_debt_ledger',
      title: 'Customer Credit & Outstanding Debt Ledger',
      dateRange: `As of ${new Date().toISOString().slice(0, 10)}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Total Outstanding Debt', value: 'GHS ' + summary.totalDebt.toFixed(2), highlight: true },
        { label: 'Debtors Count', value: summary.totalCustomersWithDebt },
        { label: 'Overdue Accounts', value: summary.totalOverdueAccounts },
        { label: 'Overdue Amount', value: 'GHS ' + summary.totalOverdueAmount.toFixed(2) },
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastr.showSuccess('Credit ledger exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastr.showSuccess('Credit ledger exported to CSV!');
    }
  }

  // ── Top Products Export ──
  exportTopProducts(format: 'excel' | 'pdf' | 'csv') {
    const prods = this.topProducts()?.by_revenue || [];
    if (!prods.length) {
      this.toastr.showError('No top products data to export.');
      return;
    }

    const headers = ['Product Name', 'SKU', 'Units Sold', 'Total Revenue (GHS)'];
    const rows = prods.map((p: any) => [
      p.name,
      p.sku || '–',
      p.quantity || 0,
      Number(p.revenue || 0).toFixed(2)
    ]);

    const options = {
      filename: 'top_products_report',
      title: 'Top Performing Products',
      dateRange: `${this.dateRange().from} to ${this.dateRange().to}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Top Product', value: prods[0]?.name || 'N/A', highlight: true },
        { label: 'Total Products', value: prods.length }
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastr.showSuccess('Top products exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastr.showSuccess('Top products exported to CSV!');
    }
  }

  // Backward-compatible template alias methods
  exportDebtorCSV() {
    this.exportCreditReport('csv');
  }

  exportTopProductsCSV() {
    this.exportTopProducts('csv');
  }
}
