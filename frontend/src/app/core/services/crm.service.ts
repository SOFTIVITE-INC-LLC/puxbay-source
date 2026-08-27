import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { CustomerTier, GiftCard, Customer } from '../models/models';
import { Observable, tap } from 'rxjs';

export interface CRMSettings {
  points_per_currency: number;
  redemption_rate: number;
}

export interface LoyaltyTransaction {
  id: string;
  customer_id: string;
  points: number;
  transaction_type: string;
  description: string;
  created_at: string;
}

export interface CustomerCreditTransaction {
  id: string;
  customer_id: string;
  amount: number;
  transaction_type: string;
  reference: string;
  notes: string;
  created_at: string;
}

export interface CustomerFeedback {
  id: string;
  customer_id: string;
  rating: number;
  comment: string;
  created_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class CrmService {
  customers = signal<any[]>([]);
  getCustomers() { return this.api.get<any[]>('/crm/customers').pipe(tap(res => this.customers.set(res || []))); }
  createCustomer(c: any) { return this.api.post('/crm/customers', c); }
  updateCustomer(id: string, c: any) { return this.api.put('/crm/customers/'+id, c); }

  private api = inject(ApiService);
  
  tiers = signal<CustomerTier[]>([]);
  giftCards = signal<GiftCard[]>([]);
  loading = signal<boolean>(false);

  // --- Loyalty Transactions ---

  getLoyaltyTransactions(customerId?: string): Observable<{transactions: LoyaltyTransaction[]}> {
    const params: any = customerId ? { customer_id: customerId } : {};
    return this.api.get<{transactions: LoyaltyTransaction[]}>('/crm/loyalty', { params });
  }

  // --- Gift Cards ---

  getGiftCards(status?: string): Observable<{gift_cards: GiftCard[]}> {
    const params = status ? { status } : undefined;
    return this.api.get<{gift_cards: GiftCard[]}>('/crm/gift-cards', { params }).pipe(
      tap(res => this.giftCards.set(res.gift_cards || []))
    );
  }

  createGiftCard(giftCard: Partial<GiftCard>): Observable<GiftCard> {
    return this.api.post<GiftCard>('/crm/gift-cards', giftCard).pipe(
      tap(newCard => this.giftCards.update(cards => [newCard, ...cards]))
    );
  }

  // --- CRM Settings ---

  getSettings(): Observable<CRMSettings> {
    return this.api.get<CRMSettings>('/crm/settings');
  }

  updateSettings(settings: CRMSettings): Observable<CRMSettings> {
    return this.api.put<CRMSettings>('/crm/settings', settings);
  }

  // --- Tiers ---

  getTiers(): Observable<{tiers: CustomerTier[]}> {
    return this.api.get<{tiers: CustomerTier[]}>('/crm/tiers').pipe(
      tap(res => this.tiers.set(res.tiers || []))
    );
  }

  createTier(tier: Partial<CustomerTier>): Observable<CustomerTier> {
    return this.api.post<CustomerTier>('/crm/tiers', tier).pipe(
      tap(newTier => this.tiers.update(ts => [...ts, newTier]))
    );
  }

  updateTier(id: string, tier: Partial<CustomerTier>): Observable<CustomerTier> {
    return this.api.put<CustomerTier>(`/crm/tiers/${id}`, tier).pipe(
      tap(updatedTier => this.tiers.update(ts => ts.map(t => t.id === id ? updatedTier : t)))
    );
  }

  deleteTier(id: string): Observable<void> {
    return this.api.delete<void>(`/crm/tiers/${id}`).pipe(
      tap(() => this.tiers.update(ts => ts.filter(t => t.id !== id)))
    );
  }

  // --- Customer Store Credit & BNPL Debt ---

  getCreditAccount(customerId: string): Observable<{
    account: {
      id: string;
      customer_id: string;
      credit_limit: number;
      balance: number;
      status: string;
      days_to_repay: number;
      last_drawdown_at?: string;
      last_repayment_at?: string;
      notes?: string;
    };
    available_credit: number;
    transactions: {
      id: string;
      amount: number;
      balance_after: number;
      transaction_type: string;
      payment_method?: string;
      reference?: string;
      notes?: string;
      paid_at?: string;
      created_at: string;
    }[];
    instalments: {
      id: string;
      order_id?: string;
      instalment_number: number;
      total_instalments: number;
      amount: number;
      amount_paid: number;
      due_date: string;
      status: string;
      paid_at?: string;
    }[];
  }> {
    return this.api.get(`/customers/${customerId}/credit-account`);
  }

  recordCustomerPayment(customerId: string, amount: number, payment_method: string = 'cash', notes: string = '', reference?: string): Observable<any> {
    return this.api.post(`/customers/${customerId}/credit/repay`, {
      amount,
      payment_method,
      reference,
      notes
    });
  }

  setCreditLimit(customerId: string, credit_limit: number, days_to_repay: number = 30, notes: string = ''): Observable<any> {
    return this.api.post(`/customers/${customerId}/credit-limit`, {
      credit_limit,
      days_to_repay,
      notes
    });
  }

  sendCreditReminder(customerId: string): Observable<any> {
    return this.api.post(`/customers/${customerId}/credit/send-reminder`, {});
  }

  getCustomerCreditTransactions(customerId: string): Observable<{transactions: CustomerCreditTransaction[], outstanding_debt: number}> {
    return this.api.get<{transactions: CustomerCreditTransaction[], outstanding_debt: number}>(`/crm/customers/${customerId}/credit`);
  }

  getOverdueAccounts(): Observable<{overdue_accounts: any[], total_overdue: number}> {
    return this.api.get<{overdue_accounts: any[], total_overdue: number}>('/credit/overdue');
  }

  getFeedbackList(): Observable<{feedback: CustomerFeedback[]}> {
    return this.api.get<{feedback: CustomerFeedback[]}>('/crm/feedback');
  }

  createFeedback(payload: { customer_id: string; rating: number; comment: string }): Observable<CustomerFeedback> {
    return this.api.post<CustomerFeedback>('/feedback', payload);
  }

  deleteFeedback(id: string): Observable<any> {
    return this.api.delete<any>(`/feedback/${id}`);
  }

  // --- Helpdesk Tickets ---
  tickets = signal<any[]>([]);
  
  getTickets(): Observable<any[]> {
    return this.api.get<any[]>('/crm/tickets').pipe(
      tap(res => this.tickets.set(res || []))
    );
  }

  createTicket(payload: { customer_id?: string; subject: string; description: string; priority: string }): Observable<any> {
    return this.api.post<any>('/crm/tickets', payload).pipe(
      tap(res => this.tickets.update(list => [res, ...list]))
    );
  }

  getTicketMessages(ticketId: string): Observable<any[]> {
    return this.api.get<any[]>(`/crm/tickets/${ticketId}/messages`);
  }

  replyTicket(ticketId: string, message: string): Observable<any> {
    return this.api.post<any>(`/crm/tickets/${ticketId}/reply`, { message });
  }
}
