import { Injectable, inject, signal } from '@angular/core';
import { tap } from 'rxjs/operators';
import { Observable } from 'rxjs';
import { ApiService } from './api.service';

export interface SMSWallet {
  id: string;
  credits_balance: number;
  credits_total: number;
  credits_used: number;
  balance_amount: number;
  price_per_sms: number;
  currency: string;
}

export interface SMSSenderID {
  id: string;
  sender_id: string;
  purpose: string;
  status: 'pending' | 'approved' | 'rejected';
  rejection_reason?: string;
  approved_at?: string;
  created_at: string;
}

export interface SMSTransaction {
  id: string;
  type: 'topup' | 'deduction';
  amount: number;
  credits_added: number;
  credits_used: number;
  price_per_sms: number;
  reference: string;
  status: string;
  description: string;
  created_at: string;
}

export interface SMSGatewayConfig {
  id?: string;
  provider: string;
  default_sender_id: string;
  price_per_sms: number;
  price_currency: string;
  is_active: boolean;
}

export interface AdminSenderIDEntry {
  id: string;
  tenant_id: string;
  tenant_name: string;
  sender_id: string;
  purpose: string;
  status: 'pending' | 'approved' | 'rejected';
  rejection_reason?: string;
  approved_at?: string;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class SMSService {
  private api = inject(ApiService);

  wallet = signal<SMSWallet | null>(null);
  senderIDs = signal<SMSSenderID[]>([]);
  transactions = signal<SMSTransaction[]>([]);
  activeSenderID = signal<SMSSenderID | null>(null);
  rate = signal<number>(0.20);
  currency = signal<string>('GHS');

  // ─── Tenant Endpoints ───────────────────────────────────────────────

  getWallet(): Observable<any> {
    return this.api.get<any>('/sms/wallet').pipe(
      tap(res => {
        this.wallet.set(res.wallet);
        this.activeSenderID.set(res.sender_id || null);
        if (res.rate) this.rate.set(res.rate);
        if (res.currency) this.currency.set(res.currency);
      })
    );
  }

  getTransactions(): Observable<SMSTransaction[]> {
    return this.api.get<SMSTransaction[]>('/sms/transactions').pipe(
      tap(list => this.transactions.set(list || []))
    );
  }

  getSenderIDs(): Observable<SMSSenderID[]> {
    return this.api.get<SMSSenderID[]>('/sms/sender-ids').pipe(
      tap(list => this.senderIDs.set(list || []))
    );
  }

  submitSenderID(senderID: string, purpose: string): Observable<SMSSenderID> {
    return this.api.post<SMSSenderID>('/sms/sender-ids', { sender_id: senderID, purpose }).pipe(
      tap(s => this.senderIDs.update(ids => [s, ...ids]))
    );
  }

  initiateTopup(amount: number, email: string): Observable<any> {
    return this.api.post<any>('/sms/wallet/topup/initiate', { amount, email });
  }

  verifyTopup(reference: string): Observable<any> {
    return this.api.post<any>('/sms/wallet/topup/verify', { reference }).pipe(
      tap(() => this.getWallet().subscribe())
    );
  }
}
