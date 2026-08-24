import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { Expense, TaxConfiguration, ExpenseCategory } from '../models/financial.models';
import { HttpParams } from '@angular/common/http';

export interface ExpenseCreateInput {
  category_id: string;
  amount: number;
  date: string;
  description?: string;
  is_recurring?: boolean;
  recurrence_interval?: string;
  receipt_url?: string;
}

export interface MonthlyFinancialData {
  month: string;
  revenue: number;
  expense: number;
}

export interface ProfitAndLoss {
  gross_revenue: number;
  cogs: number;
  gross_profit: number;
  total_expenses: number;
  net_profit: number;
  tax_collected: number;
  operating_cash_flow?: number;
  monthly_data?: MonthlyFinancialData[];
}

export interface TaxSummary {
  total_sales: number;
  total_tax_collected: number;
  taxable_amount: number;
  order_count: number;
}

export interface LedgerAccount {
  id: string;
  name: string;
  type: string;
  code: string;
  description: string;
  created_at?: string;
  updated_at?: string;
}

export interface JournalEntryLine {
  id?: string;
  account_id: string;
  amount: number;
  is_debit: boolean;
  account?: LedgerAccount;
}

export interface JournalEntry {
  id: string;
  reference_id?: string;
  reference_type: string;
  description: string;
  created_at: string;
  lines: JournalEntryLine[];
}

export interface CreateJournalEntryInput {
  reference_id?: string;
  reference_type: string;
  description: string;
  lines: Omit<JournalEntryLine, 'id' | 'account'>[];
}

@Injectable({
  providedIn: 'root'
})
export class FinancialService {
  private api = inject(ApiService);
  
  expenses = signal<Expense[]>([]);
  expenseCategories = signal<ExpenseCategory[]>([]);
  taxConfig = signal<TaxConfiguration | null>(null);
  loading = signal<boolean>(false);

  updateTaxConfig(config: any): Observable<any> { return this.api.put('/financial/taxes/config', config).pipe(tap(res => this.taxConfig.set(config))); }

  getExpenses(startDate?: string, endDate?: string): Observable<Expense[]> {
    let params = new HttpParams();
    if (startDate) params = params.set('start_date', startDate);
    if (endDate) params = params.set('end_date', endDate);
    
    this.loading.set(true);
    return this.api.get<Expense[]>('/financial/expenses', { params }).pipe(
      tap(res => {
        this.expenses.set(res || []);
        this.loading.set(false);
      })
    );
  }

  getExpenseCategories(): Observable<{categories: ExpenseCategory[]}> {
    return this.api.get<{categories: ExpenseCategory[]}>('/financial/expense-categories').pipe(
      tap(res => this.expenseCategories.set(res?.categories || []))
    );
  }

  createExpenseCategory(category: { name: string, description?: string }): Observable<ExpenseCategory> {
    return this.api.post<ExpenseCategory>('/financial/expense-categories', category).pipe(
      tap(res => {
        this.expenseCategories.update(list => [res, ...list]);
      })
    );
  }

  createExpense(input: ExpenseCreateInput): Observable<Expense> {
    return this.api.post<Expense>('/financial/expenses', input).pipe(
      tap(res => {
        this.expenses.update(list => [res, ...list]);
      })
    );
  }

  updateExpense(id: string, input: ExpenseCreateInput): Observable<Expense> {
    return this.api.put<Expense>(`/financial/expenses/${id}`, input).pipe(
      tap(updated => {
        this.expenses.update(list => list.map(e => e.id === id ? updated : e));
      })
    );
  }

  deleteExpense(id: string): Observable<void> {
    return this.api.delete<void>(`/financial/expenses/${id}`).pipe(
      tap(() => {
        this.expenses.update(list => list.filter(e => e.id !== id));
      })
    );
  }

  getProfitAndLoss(startDate?: string, endDate?: string): Observable<ProfitAndLoss> {
    let params = new HttpParams();
    if (startDate) params = params.set('start_date', startDate);
    if (endDate) params = params.set('end_date', endDate);
    return this.api.get<ProfitAndLoss>('/financial/profit-and-loss', { params });
  }

  getTaxConfig(): Observable<TaxConfiguration> {
    return this.api.get<TaxConfiguration>('/financial/taxes/config').pipe(
      tap(res => this.taxConfig.set(res))
    );
  }

  getTaxReport(startDate?: string, endDate?: string): Observable<{tax_summary: TaxSummary}> {
    let params = new HttpParams();
    if (startDate) params = params.set('start_date', startDate);
    if (endDate) params = params.set('end_date', endDate);
    return this.api.get<{tax_summary: TaxSummary}>('/financial/taxes/report', { params });
  }

  // --- Ledger & Accounting ---
  
  getLedgerAccounts(): Observable<LedgerAccount[]> {
    return this.api.get<LedgerAccount[]>('/financial/ledger');
  }

  createLedgerAccount(input: { name: string, type: string, code: string, description: string }): Observable<LedgerAccount> {
    return this.api.post<LedgerAccount>('/financial/ledger', input);
  }

  getJournalEntries(): Observable<JournalEntry[]> {
    return this.api.get<JournalEntry[]>('/financial/journal-entries');
  }

  createJournalEntry(input: CreateJournalEntryInput): Observable<JournalEntry> {
    return this.api.post<JournalEntry>('/financial/journal-entries', input);
  }
}
