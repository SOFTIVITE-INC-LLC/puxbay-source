import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap, BehaviorSubject } from 'rxjs';
import { environment } from '../../../../environments/environment';

export interface SupplierProfile {
  id: string;
  name: string;
  contact_person?: string;
  email?: string;
  phone?: string;
  address?: string;
  website?: string;
  tax_number?: string;
  credit_balance?: number;
  portal_email?: string;
}

export interface SupplierLoginResponse {
  token: string;
  supplier: SupplierProfile;
}

export interface PurchaseOrderItem {
  id: string;
  product_id: string;
  product?: {
    id: string;
    name: string;
    sku: string;
    image_url?: string;
  };
  quantity_ordered: number;
  quantity_received: number;
  unit_cost: number;
}

export interface PurchaseOrder {
  id: string;
  created_at: string;
  po_number: string;
  status: string; // draft, issued, confirmed, partially_received, received, cancelled, rejected
  total_amount: number;
  expected_date?: string;
  notes?: string;
  items?: PurchaseOrderItem[];
}

export interface SupplierASN {
  id?: string;
  created_at?: string;
  asn_number?: string;
  purchase_order_id?: string;
  purchase_order?: PurchaseOrder;
  carrier: string;
  tracking_number?: string;
  dispatch_date: string;
  expected_arrival?: string;
  package_count: number;
  total_weight_kg: number;
  status: string; // dispatched, in_transit, delivered, rejected
  notes?: string;
}

export interface SupplierInvoice {
  id?: string;
  created_at?: string;
  purchase_order_id?: string;
  purchase_order?: PurchaseOrder;
  invoice_number: string;
  issue_date: string;
  due_date: string;
  subtotal: number;
  tax: number;
  total: number;
  amount_paid: number;
  status: string; // pending, approved, partially_paid, paid, rejected
  payment_ref?: string;
  notes?: string;
}

export interface SupplierProduct {
  id: string;
  supplier_id: string;
  product_id: string;
  supplier_sku?: string;
  unit_cost: number;
  min_order_qty: number;
  product?: {
    id: string;
    name: string;
    sku: string;
    current_stock: number;
    image_url?: string;
  };
}

export interface SupplierPriceChangeRequest {
  id?: string;
  created_at?: string;
  product_id: string;
  product?: {
    id: string;
    name: string;
    sku: string;
  };
  current_cost: number;
  proposed_cost: number;
  effective_date: string;
  reason: string;
  status: string; // pending, approved, rejected
  review_notes?: string;
}

export interface SupplierQuote {
  id?: string;
  created_at?: string;
  quote_number?: string;
  title: string;
  total_amount: number;
  currency: string;
  valid_until: string;
  lead_time_days: number;
  payment_terms: string;
  status: string; // draft, submitted, accepted, rejected, expired
  notes?: string;
}

export interface SupplierPayoutAccount {
  id?: string;
  account_type: string; // bank, momo, stripe
  bank_name?: string;
  account_number?: string;
  account_name?: string;
  routing_code?: string;
  momo_network?: string;
  momo_number?: string;
  is_default: boolean;
}

export interface SupplierMessage {
  id?: string;
  created_at?: string;
  sender_type: string; // supplier, merchant
  sender_name: string;
  reference_id?: string;
  message: string;
  is_read?: boolean;
}

export interface DashboardStats {
  total_pos: number;
  pending_deliveries: number;
  total_invoiced: number;
  open_quotes: number;
  otd_score: number;
}

@Injectable({
  providedIn: 'root'
})
export class SupplierPortalService {
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl + '/supplier-portal';

  private currentSupplierSub = new BehaviorSubject<SupplierProfile | null>(null);
  currentSupplier$ = this.currentSupplierSub.asObservable();
  
  loading = signal<boolean>(false);

  login(credentials: { email: string; password: string }): Observable<SupplierLoginResponse> {
    return this.http.post<SupplierLoginResponse>(`${this.apiUrl}/login`, credentials).pipe(
      tap(res => {
        if (res && res.token) {
          localStorage.setItem('supplier_token', res.token);
          this.currentSupplierSub.next(res.supplier);
        }
      })
    );
  }

  logout() {
    this.http.post(`${this.apiUrl}/logout`, {}).subscribe({
      next: () => {},
      error: () => {}
    });
    localStorage.removeItem('supplier_token');
    this.currentSupplierSub.next(null);
  }

  getMe(): Observable<SupplierProfile> {
    return this.http.get<SupplierProfile>(`${this.apiUrl}/me`).pipe(
      tap(res => this.currentSupplierSub.next(res))
    );
  }

  getDashboard(): Observable<DashboardStats> {
    return this.http.get<DashboardStats>(`${this.apiUrl}/dashboard`);
  }

  getPurchaseOrders(): Observable<PurchaseOrder[]> {
    this.loading.set(true);
    return this.http.get<PurchaseOrder[]>(`${this.apiUrl}/purchase-orders`).pipe(
      tap(() => this.loading.set(false))
    );
  }

  acknowledgePO(poId: string, payload: { status: string; expected_date?: string; notes?: string }): Observable<PurchaseOrder> {
    return this.http.post<PurchaseOrder>(`${this.apiUrl}/purchase-orders/${poId}/acknowledge`, payload);
  }

  flipPOToInvoice(poId: string, payload: { invoice_number?: string; due_date?: string }): Observable<SupplierInvoice> {
    return this.http.post<SupplierInvoice>(`${this.apiUrl}/purchase-orders/${poId}/invoice`, payload);
  }

  getShipments(): Observable<SupplierASN[]> {
    return this.http.get<SupplierASN[]>(`${this.apiUrl}/shipments`);
  }

  createShipment(asn: Partial<SupplierASN>): Observable<SupplierASN> {
    return this.http.post<SupplierASN>(`${this.apiUrl}/shipments`, asn);
  }

  getInvoices(): Observable<SupplierInvoice[]> {
    return this.http.get<SupplierInvoice[]>(`${this.apiUrl}/invoices`);
  }

  getCatalog(): Observable<SupplierProduct[]> {
    return this.http.get<SupplierProduct[]>(`${this.apiUrl}/catalog`);
  }

  getPriceRequests(): Observable<SupplierPriceChangeRequest[]> {
    return this.http.get<SupplierPriceChangeRequest[]>(`${this.apiUrl}/price-requests`);
  }

  createPriceRequest(req: Partial<SupplierPriceChangeRequest>): Observable<SupplierPriceChangeRequest> {
    return this.http.post<SupplierPriceChangeRequest>(`${this.apiUrl}/price-requests`, req);
  }

  getQuotes(): Observable<SupplierQuote[]> {
    return this.http.get<SupplierQuote[]>(`${this.apiUrl}/quotes`);
  }

  createQuote(quote: Partial<SupplierQuote>): Observable<SupplierQuote> {
    return this.http.post<SupplierQuote>(`${this.apiUrl}/quotes`, quote);
  }

  getPayoutAccount(): Observable<SupplierPayoutAccount> {
    return this.http.get<SupplierPayoutAccount>(`${this.apiUrl}/payout-account`);
  }

  savePayoutAccount(account: Partial<SupplierPayoutAccount>): Observable<SupplierPayoutAccount> {
    return this.http.post<SupplierPayoutAccount>(`${this.apiUrl}/payout-account`, account);
  }

  getMessages(refId?: string): Observable<SupplierMessage[]> {
    const params = refId ? `?reference_id=${encodeURIComponent(refId)}` : '';
    return this.http.get<SupplierMessage[]>(`${this.apiUrl}/messages${params}`);
  }

  sendMessage(msg: Partial<SupplierMessage>): Observable<SupplierMessage> {
    return this.http.post<SupplierMessage>(`${this.apiUrl}/messages`, msg);
  }
}
