import { ToastService } from '../../../core/services/toast';
import { ExportService } from '../../../core/services/export.service';
import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { provideCharts, withDefaultRegisterables, BaseChartDirective } from 'ng2-charts';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { FinancialService, ExpenseCreateInput } from '../../../core/services/financial.service';
import { Expense } from '../../../core/models/financial.models';
import { SettingsService } from '../../../core/services/settings.service';

@Component({
  selector: 'app-financial',
  standalone: true,
  imports: [CommonModule, FormsModule, BaseChartDirective, AppCurrencyPipe],
  providers: [provideCharts(withDefaultRegisterables())],
  templateUrl: './financial.html',
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
export class Financial implements OnInit {
  toastService = inject(ToastService);
  financialService = inject(FinancialService);
  settingsService = inject(SettingsService);
  activeTab = signal<'pnl' | 'expenses' | 'tax' | 'ledger' | 'journal'>('pnl');
  dateRange = signal<'this_month' | 'this_quarter' | 'this_year' | 'all_time'>('this_month');
  
  pnlData = signal<any>(null);
  taxReportData = signal<any>(null);

  // Financial Alerts
  financialAlerts = computed(() => {
    const pnl = this.pnlData();
    const alerts: { type: 'warning' | 'danger' | 'info'; message: string }[] = [];
    if (!pnl) return alerts;

    if (pnl.net_profit < 0) {
      alerts.push({ type: 'danger', message: `You are operating at a net loss of $${Math.abs(pnl.net_profit).toFixed(2)} for the selected period.` });
    } else if (pnl.gross_revenue > 0 && (pnl.net_profit / pnl.gross_revenue) < 0.1) {
      alerts.push({ type: 'warning', message: 'Your net margin is below 10%. Consider reviewing operating expenses.' });
    }
    return alerts;
  });

  // Drill-down State
  isExpenseDetailsOpen = signal(false);
  selectedExpense = signal<Expense | null>(null);

  // Add/Edit Expense State
  isExpenseModalOpen = signal(false);
  isExpenseCategoryModalOpen = signal(false);
  savingExpenseCategory = signal(false);
  newExpenseCategory = signal({ name: '', type: 'operating', description: '', monthly_budget: 0 });
  savingExpense = signal(false);
  editingExpenseId = signal<string | null>(null);
  expenseForm = signal<ExpenseCreateInput>({
    category_id: '',
    amount: 0,
    date: new Date().toISOString().split('T')[0],
    description: '',
    is_recurring: false,
    recurrence_interval: 'monthly',
    receipt_url: ''
  });

  // Delete State
  isDeleteExpenseOpen = signal(false);
  expenseToDelete = signal<any>(null);

  // Expense Filtering
  expenseSearch = signal('');
  expenseCategoryFilter = signal('');

  // Chart Configuration
  chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: true, position: 'bottom' as const },
      tooltip: {
        backgroundColor: 'rgba(24, 24, 27, 0.9)',
        padding: 12,
        titleFont: { size: 14, weight: 'bold' },
        bodyFont: { size: 13 },
        cornerRadius: 8,
      }
    },
    scales: {
      y: { border: { display: false }, grid: { color: 'rgba(0,0,0,0.05)' } },
      x: { border: { display: false }, grid: { display: false } }
    }
  };

  chartData = signal({
    labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],
    datasets: [
      {
        label: 'Gross Revenue',
        data: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
        borderColor: '#10b981', // emerald-500
        backgroundColor: 'rgba(16, 185, 129, 0.1)',
        tension: 0.4,
        fill: true
      },
      {
        label: 'Expenses',
        data: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
        borderColor: '#f43f5e', // rose-500
        backgroundColor: 'rgba(244, 63, 94, 0.1)',
        tension: 0.4,
        fill: true
      }
    ]
  });

  // Ledger State
  ledgerAccounts = signal<any[]>([]);
  journalEntries = signal<any[]>([]);

  // Create Ledger State
  isLedgerModalOpen = signal(false);
  ledgerForm = signal({ name: '', type: 'Asset', code: '', description: '' });

  isJournalModalOpen = signal(false);
  journalForm = signal({
    reference_id: '',
    reference_type: '',
    description: '',
    lines: [
      { account_id: '', amount: 0, is_debit: true },
      { account_id: '', amount: 0, is_debit: false }
    ]
  });

  ngOnInit() {
    this.financialService.getExpenseCategories().subscribe();
    this.financialService.getTaxConfig().subscribe();
    this.fetchData();
  }

  fetchLedgerData() {
    this.financialService.getLedgerAccounts().subscribe(res => {
      console.log('Ledger Accounts:', res);
      this.ledgerAccounts.set((res as any)?.data || res || []);
    });
    this.financialService.getJournalEntries().subscribe(res => {
      console.log('Journal Entries:', res);
      this.journalEntries.set((res as any)?.data || res || []);
    });
  }

  getDateRangeParams(): { start_date?: string, end_date?: string } {
    const range = this.dateRange();
    const now = new Date();
    let start_date: string | undefined;
    let end_date: string | undefined;

    if (range === 'this_month') {
      start_date = new Date(now.getFullYear(), now.getMonth(), 1).toISOString().split('T')[0];
    } else if (range === 'this_quarter') {
      const q = Math.floor(now.getMonth() / 3);
      start_date = new Date(now.getFullYear(), q * 3, 1).toISOString().split('T')[0];
    } else if (range === 'this_year') {
      start_date = new Date(now.getFullYear(), 0, 1).toISOString().split('T')[0];
    }

    return { start_date, end_date };
  }

  fetchData() {
    const { start_date, end_date } = this.getDateRangeParams();
    this.financialService.getExpenses(start_date, end_date).subscribe();
    this.fetchLedgerData();
    this.financialService.getProfitAndLoss(start_date, end_date).subscribe(res => {
      this.pnlData.set(res);
      if (res.monthly_data && res.monthly_data.length > 0) {
        this.chartData.set({
          labels: res.monthly_data.map((d: any) => d.month),
          datasets: [
            {
              label: 'Gross Revenue',
              data: res.monthly_data.map((d: any) => d.revenue),
              borderColor: '#10b981',
              backgroundColor: 'rgba(16, 185, 129, 0.1)',
              tension: 0.4,
              fill: true
            },
            {
              label: 'Expenses',
              data: res.monthly_data.map((d: any) => d.expense),
              borderColor: '#f43f5e',
              backgroundColor: 'rgba(244, 63, 94, 0.1)',
              tension: 0.4,
              fill: true
            }
          ]
        });
      }
    });
    this.financialService.getTaxReport(start_date, end_date).subscribe(res => {
      this.taxReportData.set(res.tax_summary);
    });
  }

  onDateRangeChange(range: any) {
    this.dateRange.set(range);
    this.fetchData();
  }

  exportService = inject(ExportService);

  exportActiveFinancialTab(format: 'excel' | 'pdf' | 'csv') {
    const tab = this.activeTab();
    switch (tab) {
      case 'pnl':
        this.exportPnL(format);
        break;
      case 'expenses':
        this.exportExpenses(format);
        break;
      case 'tax':
        this.exportTaxReport(format);
        break;
      case 'ledger':
        this.exportLedger(format);
        break;
      case 'journal':
        this.exportJournal(format);
        break;
    }
  }

  exportPnL(format: 'excel' | 'pdf' | 'csv') {
    const pnl = this.pnlData();
    if (!pnl) {
      this.toastService.showError('No P&L data to export.');
      return;
    }

    const headers = ['Line Item', 'Amount (GHS)', '% of Revenue'];
    const rev = Number(pnl.gross_revenue || pnl.revenue || 0);
    const getPct = (val: number) => rev > 0 ? ((val / rev) * 100).toFixed(1) + '%' : '0.0%';

    const rows: any[][] = [
      ['Gross Sales / Revenue', rev.toFixed(2), '100.0%'],
      ['Cost of Goods Sold (COGS)', Number(pnl.cogs || 0).toFixed(2), getPct(Number(pnl.cogs || 0))],
      ['Gross Profit', Number(pnl.gross_profit || (rev - (pnl.cogs || 0))).toFixed(2), getPct(Number(pnl.gross_profit || (rev - (pnl.cogs || 0))))],
      ['Operating Expenses', Number(pnl.total_expenses || pnl.expenses || 0).toFixed(2), getPct(Number(pnl.total_expenses || pnl.expenses || 0))],
      ['Tax / VAT Expense', Number(pnl.tax || 0).toFixed(2), getPct(Number(pnl.tax || 0))],
      ['Net Profit / (Loss)', Number(pnl.net_profit || 0).toFixed(2), getPct(Number(pnl.net_profit || 0))],
    ];

    // Add category breakdown if available
    if (pnl.expense_breakdown && pnl.expense_breakdown.length) {
      rows.push(['--- Expense Breakdown by Category ---', '', '']);
      for (const eb of pnl.expense_breakdown) {
        rows.push([`  - ${eb.category || eb.name}`, Number(eb.amount || 0).toFixed(2), getPct(Number(eb.amount || 0))]);
      }
    }

    const marginPct = rev > 0 ? ((pnl.net_profit / rev) * 100).toFixed(1) + '%' : '0.0%';

    const options = {
      filename: `pnl_statement_${this.dateRange()}`,
      title: 'Profit & Loss Statement (Income Statement)',
      subtitle: `Accounting Period: ${this.dateRange().replace('_', ' ').toUpperCase()}`,
      dateRange: `Period: ${this.dateRange().replace('_', ' ').toUpperCase()}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Gross Revenue', value: 'GHS ' + rev.toFixed(2), highlight: true },
        { label: 'Total Expenses', value: 'GHS ' + Number(pnl.total_expenses || pnl.expenses || 0).toFixed(2) },
        { label: 'Net Profit', value: 'GHS ' + Number(pnl.net_profit || 0).toFixed(2), highlight: pnl.net_profit >= 0 },
        { label: 'Net Margin', value: marginPct },
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastService.showSuccess('P&L Statement exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastService.showSuccess('P&L Statement exported to CSV!');
    }
  }

  exportExpenses(format: 'excel' | 'pdf' | 'csv') {
    const expenses = this.filteredExpenses();
    if (!expenses.length) {
      this.toastService.showError('No expenses to export.');
      return;
    }

    const headers = ['Date', 'Category', 'Description', 'Amount (GHS)', 'Recurring', 'Interval'];
    const rows = expenses.map(e => [
      e.date ? e.date.split('T')[0] : '',
      e.category?.name || e.category_id || 'General',
      e.description || '–',
      Number(e.amount || 0).toFixed(2),
      e.is_recurring ? 'Yes' : 'No',
      e.recurrence_interval || '–'
    ]);

    const totalAmt = expenses.reduce((sum, e) => sum + (e.amount || 0), 0);

    const options = {
      filename: `expenses_${this.dateRange()}`,
      title: 'Operating Expenses Ledger',
      subtitle: `Period: ${this.dateRange().replace('_', ' ').toUpperCase()}`,
      dateRange: `Period: ${this.dateRange().replace('_', ' ').toUpperCase()}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Total Expenses', value: 'GHS ' + totalAmt.toFixed(2), highlight: true },
        { label: 'Expense Entries', value: expenses.length },
        { label: 'Average Expense', value: 'GHS ' + (expenses.length ? (totalAmt / expenses.length).toFixed(2) : '0.00') },
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastService.showSuccess('Expenses exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastService.showSuccess('Expenses exported to CSV!');
    }
  }

  exportTaxReport(format: 'excel' | 'pdf' | 'csv') {
    const tr = this.taxReportData();
    if (!tr) {
      this.toastService.showError('No tax report data to export.');
      return;
    }

    const headers = ['Tax Metric', 'Amount (GHS)'];
    const rows: any[][] = [
      ['Total Gross Sales', Number(tr.total_sales || 0).toFixed(2)],
      ['Taxable Sales Amount', Number(tr.taxable_amount || 0).toFixed(2)],
      ['Total Tax Collected', Number(tr.total_tax_collected || 0).toFixed(2)],
      ['Total Orders Count', Number(tr.order_count || 0)],
      ['Effective Tax Rate', tr.taxable_amount > 0 ? ((tr.total_tax_collected / tr.taxable_amount) * 100).toFixed(2) + '%' : '0.0%'],
    ];

    const options = {
      filename: `tax_report_${this.dateRange()}`,
      title: 'Tax Summary & Filing Report',
      subtitle: `Period: ${this.dateRange().replace('_', ' ').toUpperCase()}`,
      dateRange: `Period: ${this.dateRange().replace('_', ' ').toUpperCase()}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Tax Collected', value: 'GHS ' + Number(tr.total_tax_collected || 0).toFixed(2), highlight: true },
        { label: 'Taxable Sales', value: 'GHS ' + Number(tr.taxable_amount || 0).toFixed(2) },
        { label: 'Total Orders', value: tr.order_count || 0 },
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastService.showSuccess('Tax report exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastService.showSuccess('Tax report exported to CSV!');
    }
  }

  exportLedger(format: 'excel' | 'pdf' | 'csv') {
    const accounts = this.ledgerAccounts();
    if (!accounts.length) {
      this.toastService.showError('No ledger accounts to export.');
      return;
    }

    const headers = ['Account Code', 'Account Name', 'Type', 'Balance (GHS)', 'Description'];
    const rows = accounts.map(a => [
      a.code || '–',
      a.name || '–',
      a.type || '–',
      Number(a.balance || 0).toFixed(2),
      a.description || '–'
    ]);

    const options = {
      filename: 'chart_of_accounts',
      title: 'Chart of Accounts (General Ledger)',
      dateRange: `As of ${new Date().toISOString().slice(0, 10)}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Total Accounts', value: accounts.length, highlight: true }
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastService.showSuccess('Chart of Accounts exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastService.showSuccess('Chart of Accounts exported to CSV!');
    }
  }

  exportJournal(format: 'excel' | 'pdf' | 'csv') {
    const entries = this.journalEntries();
    if (!entries.length) {
      this.toastService.showError('No journal entries to export.');
      return;
    }

    const headers = ['Date', 'Reference ID', 'Reference Type', 'Description', 'Lines Count'];
    const rows = entries.map(j => [
      j.created_at ? new Date(j.created_at).toLocaleDateString() : '–',
      j.reference_id || '–',
      j.reference_type || 'Manual',
      j.description || '–',
      j.lines?.length || 0
    ]);

    const options = {
      filename: 'journal_entries',
      title: 'General Journal Entries Record',
      dateRange: `As of ${new Date().toISOString().slice(0, 10)}`,
      headers,
      rows,
      summaryCards: [
        { label: 'Total Entries', value: entries.length, highlight: true }
      ]
    };

    if (format === 'excel') {
      this.exportService.exportToExcel(options);
      this.toastService.showSuccess('Journal entries exported to Excel!');
    } else if (format === 'pdf') {
      this.exportService.exportToPdf(options);
    } else {
      this.exportService.exportToCsv(options);
      this.toastService.showSuccess('Journal entries exported to CSV!');
    }
  }

  printPnL() {
    this.exportPnL('pdf');
  }

  saveTaxConfig() {
    const config = this.financialService.taxConfig();
    if (config) {
      this.financialService.updateTaxConfig(config).subscribe(() => {
        this.toastService.showSuccess('Tax Configuration Saved Successfully');
      });
    }
  }

  openExpenseModal() {
    this.editingExpenseId.set(null);
    this.expenseForm.set({
      category_id: '',
      amount: 0,
      date: new Date().toISOString().split('T')[0],
      description: '',
      is_recurring: false,
      recurrence_interval: 'monthly',
      receipt_url: ''
    });
    this.isExpenseModalOpen.set(true);
  }

  openEditExpenseModal(expense: any) {
    this.editingExpenseId.set(expense.id);
    const dateStr = expense.date ? new Date(expense.date).toISOString().split('T')[0] : new Date().toISOString().split('T')[0];
    this.expenseForm.set({
      category_id: expense.category_id || '',
      amount: expense.amount || 0,
      date: dateStr,
      description: expense.description || '',
      is_recurring: expense.is_recurring || false,
      recurrence_interval: expense.recurrence_interval || 'monthly',
      receipt_url: expense.receipt_url || ''
    });
    this.closeExpenseDetails();
    this.isExpenseModalOpen.set(true);
  }

  closeExpenseModal() {
    this.isExpenseModalOpen.set(false);
  }

  openExpenseCategoryModal() {
    this.newExpenseCategory.set({ name: '', type: 'operating', description: '', monthly_budget: 0 });
    this.isExpenseCategoryModalOpen.set(true);
  }

  closeExpenseCategoryModal() {
    this.isExpenseCategoryModalOpen.set(false);
  }

  saveExpenseCategory() {
    const cat = this.newExpenseCategory();
    if (!cat.name) return;
    this.savingExpenseCategory.set(true);
    this.financialService.createExpenseCategory(cat).subscribe({
      next: () => {
        this.savingExpenseCategory.set(false);
        this.closeExpenseCategoryModal();
      },
      error: () => this.savingExpenseCategory.set(false)
    });
  }

  openExpenseDetails(expense: Expense) {
    this.selectedExpense.set(expense);
    this.isExpenseDetailsOpen.set(true);
  }

  closeExpenseDetails() {
    this.isExpenseDetailsOpen.set(false);
    this.selectedExpense.set(null);
  }

  saveExpense() {
    const form = this.expenseForm();
    if (!form.category_id || form.amount <= 0 || !form.date) return;

    this.savingExpense.set(true);
    const editId = this.editingExpenseId();
    const request$ = editId
      ? this.financialService.updateExpense(editId, form)
      : this.financialService.createExpense(form);

    request$.subscribe({
      next: () => {
        this.savingExpense.set(false);
        this.closeExpenseModal();
        this.fetchData();
      },
      error: () => this.savingExpense.set(false)
    });
  }

  confirmDeleteExpense(expense: any) {
    this.expenseToDelete.set(expense);
    this.closeExpenseDetails();
    this.isDeleteExpenseOpen.set(true);
  }

  executeDeleteExpense() {
    const expense = this.expenseToDelete();
    if (!expense) return;
    this.savingExpense.set(true);
    this.financialService.deleteExpense(expense.id).subscribe({
      next: () => {
        this.savingExpense.set(false);
        this.isDeleteExpenseOpen.set(false);
        this.expenseToDelete.set(null);
        this.fetchData();
      },
      error: () => this.savingExpense.set(false)
    });
  }

  filteredExpenses() {
    const search = this.expenseSearch().toLowerCase();
    const cat = this.expenseCategoryFilter();
    return this.financialService.expenses().filter(e => {
      const matchSearch = !search || (e.description || '').toLowerCase().includes(search);
      const matchCat = !cat || e.category_id === cat;
      return matchSearch && matchCat;
    });
  }

  saveLedgerAccount() {
    const form = this.ledgerForm();
    if (!form.name || !form.code) {
      this.toastService.showError('Name and Code are required');
      return;
    }
    this.financialService.createLedgerAccount(form).subscribe({
      next: (res) => {
        this.ledgerAccounts.update(list => [...list, res]);
        this.isLedgerModalOpen.set(false);
        this.ledgerForm.set({ name: '', type: 'Asset', code: '', description: '' });
        this.toastService.showSuccess('Ledger account created successfully');
      },
      error: (err) => this.toastService.showError('Failed to create ledger account')
    });
  }

  addJournalLine() {
    this.journalForm.update(form => ({
      ...form,
      lines: [...form.lines, { account_id: '', amount: 0, is_debit: true }]
    }));
  }

  removeJournalLine(index: number) {
    this.journalForm.update(form => {
      const lines = [...form.lines];
      lines.splice(index, 1);
      return { ...form, lines };
    });
  }

  saveJournalEntry() {
    const form = this.journalForm();
    if (!form.description || form.lines.length < 2) {
      this.toastService.showError('Description and at least 2 lines are required');
      return;
    }

    // Basic validation
    let debits = 0;
    let credits = 0;
    for (const line of form.lines) {
      if (!line.account_id || line.amount <= 0) {
        this.toastService.showError('All lines must have an account and amount > 0');
        return;
      }
      if (line.is_debit) debits += line.amount;
      else credits += line.amount;
    }

    if (Math.abs(debits - credits) > 0.01) {
      this.toastService.showError(`Debits (${debits}) must equal Credits (${credits})`);
      return;
    }

    const payload = {
      description: form.description,
      reference_type: form.reference_type || 'Manual',
      reference_id: form.reference_id || undefined,
      lines: form.lines
    };

    this.financialService.createJournalEntry(payload).subscribe({
      next: (res) => {
        this.journalEntries.update(list => [res, ...list]);
        this.isJournalModalOpen.set(false);
        this.journalForm.set({
          reference_id: '',
          reference_type: '',
          description: '',
          lines: [
            { account_id: '', amount: 0, is_debit: true },
            { account_id: '', amount: 0, is_debit: false }
          ]
        });
        this.toastService.showSuccess('Journal entry recorded');
      },
      error: (err) => this.toastService.showError(err.error?.error || 'Failed to record journal entry')
    });
  }
}

