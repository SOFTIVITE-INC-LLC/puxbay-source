import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap, BehaviorSubject } from 'rxjs';
import { ApiService } from '../../../core/services/api.service';

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

export interface DemandForecast {
  product_id: string;
  product_name: string;
  sku: string;
  current_stock: number;
  daily_velocity: number;
  forecast_30d_qty: number;
  suggested_restock: number;
  estimated_po_value: number;
  urgency: 'low' | 'medium' | 'high';
}

export interface SupplierTeamMember {
  id?: string;
  created_at?: string;
  full_name: string;
  email: string;
  role: 'admin' | 'finance' | 'warehouse' | 'sales';
  phone?: string;
  is_active?: boolean;
}

export interface PayoutResult {
  success: boolean;
  invoice_number: string;
  payout_ref: string;
  amount_settled: number;
  status: string;
}

export interface QRScanResult {
  valid: boolean;
  po_number?: string;
  status?: string;
  total_amount?: number;
  item_count?: number;
  message: string;
}

@Injectable({
  providedIn: 'root'
})
export class SupplierPortalService {
  private api = inject(ApiService);
  private readonly base = '/supplier-portal';

  private currentSupplierSub = new BehaviorSubject<SupplierProfile | null>(null);
  currentSupplier$ = this.currentSupplierSub.asObservable();

  loading = signal<boolean>(false);

  login(credentials: { email: string; password: string }): Observable<SupplierLoginResponse> {
    return this.api.post<SupplierLoginResponse>(`${this.base}/login`, credentials).pipe(
      tap(res => {
        if (res?.token) {
          localStorage.setItem('supplier_token', res.token);
          this.currentSupplierSub.next(res.supplier);
        }
      })
    );
  }

  logout() {
    this.api.post(`${this.base}/logout`, {}).subscribe({ next: () => {}, error: () => {} });
    localStorage.removeItem('supplier_token');
    this.currentSupplierSub.next(null);
  }

  getMe(): Observable<SupplierProfile> {
    return this.api.get<SupplierProfile>(`${this.base}/me`).pipe(
      tap(res => this.currentSupplierSub.next(res))
    );
  }

  getDashboard(): Observable<DashboardStats> {
    return this.api.get<DashboardStats>(`${this.base}/dashboard`);
  }

  getPurchaseOrders(): Observable<PurchaseOrder[]> {
    this.loading.set(true);
    return this.api.get<PurchaseOrder[]>(`${this.base}/purchase-orders`, undefined, true).pipe(
      tap(() => this.loading.set(false))
    );
  }

  acknowledgePO(poId: string, payload: { status: string; expected_date?: string; notes?: string }): Observable<PurchaseOrder> {
    return this.api.post<PurchaseOrder>(`${this.base}/purchase-orders/${poId}/acknowledge`, payload);
  }

  flipPOToInvoice(poId: string, payload: { invoice_number?: string; due_date?: string }): Observable<SupplierInvoice> {
    return this.api.post<SupplierInvoice>(`${this.base}/purchase-orders/${poId}/invoice`, payload);
  }

  getShipments(): Observable<SupplierASN[]> {
    return this.api.get<SupplierASN[]>(`${this.base}/shipments`);
  }

  createShipment(asn: Partial<SupplierASN>): Observable<SupplierASN> {
    return this.api.post<SupplierASN>(`${this.base}/shipments`, asn);
  }

  getInvoices(): Observable<SupplierInvoice[]> {
    return this.api.get<SupplierInvoice[]>(`${this.base}/invoices`);
  }

  getCatalog(): Observable<SupplierProduct[]> {
    return this.api.get<SupplierProduct[]>(`${this.base}/catalog`);
  }

  getPriceRequests(): Observable<SupplierPriceChangeRequest[]> {
    return this.api.get<SupplierPriceChangeRequest[]>(`${this.base}/price-requests`);
  }

  createPriceRequest(req: Partial<SupplierPriceChangeRequest>): Observable<SupplierPriceChangeRequest> {
    return this.api.post<SupplierPriceChangeRequest>(`${this.base}/price-requests`, req);
  }

  getQuotes(): Observable<SupplierQuote[]> {
    return this.api.get<SupplierQuote[]>(`${this.base}/quotes`);
  }

  createQuote(quote: Partial<SupplierQuote>): Observable<SupplierQuote> {
    return this.api.post<SupplierQuote>(`${this.base}/quotes`, quote);
  }

  getPayoutAccount(): Observable<SupplierPayoutAccount> {
    return this.api.get<SupplierPayoutAccount>(`${this.base}/payout-account`);
  }

  savePayoutAccount(account: Partial<SupplierPayoutAccount>): Observable<SupplierPayoutAccount> {
    return this.api.post<SupplierPayoutAccount>(`${this.base}/payout-account`, account);
  }

  getMessages(refId?: string): Observable<SupplierMessage[]> {
    const params = refId ? { reference_id: refId } : undefined;
    return this.api.get<SupplierMessage[]>(`${this.base}/messages`, params ? { params } : undefined);
  }

  sendMessage(msg: Partial<SupplierMessage>): Observable<SupplierMessage> {
    return this.api.post<SupplierMessage>(`${this.base}/messages`, msg);
  }

  // Phase 2: Demand Forecasts
  getForecasts(): Observable<DemandForecast[]> {
    return this.api.get<DemandForecast[]>(`${this.base}/forecasts`);
  }

  // Phase 2: Team RBAC
  getTeam(): Observable<SupplierTeamMember[]> {
    return this.api.get<SupplierTeamMember[]>(`${this.base}/team`, undefined, true);
  }

  inviteTeamMember(member: Partial<SupplierTeamMember>): Observable<SupplierTeamMember> {
    return this.api.post<SupplierTeamMember>(`${this.base}/team/invite`, member);
  }

  // Phase 2: Early Payout
  initiateEarlyPayout(invoiceId: string): Observable<PayoutResult> {
    return this.api.post<PayoutResult>(`${this.base}/invoices/${invoiceId}/payout`, {});
  }

  // Phase 2: QR Receiving Scanner
  submitQRScan(qrPayload: string): Observable<QRScanResult> {
    return this.api.post<QRScanResult>(`${this.base}/receive-scan`, { qr_payload: qrPayload });
  }

  // Phase 3: RMAs & Defect Claims
  getRMAs(): Observable<SupplierRMA[]> {
    return this.api.get<SupplierRMA[]>(`${this.base}/rmas`, undefined, true);
  }

  dispatchRMAReplacement(rmaId: string, notes: string): Observable<SupplierRMA> {
    return this.api.post<SupplierRMA>(`${this.base}/rmas/${rmaId}/replace`, { notes });
  }

  // Phase 3: Dock Delivery Slots
  getDockSlots(): Observable<SupplierDeliverySlot[]> {
    return this.api.get<SupplierDeliverySlot[]>(`${this.base}/dock-slots`);
  }

  bookDockSlot(slot: Partial<SupplierDeliverySlot>): Observable<SupplierDeliverySlot> {
    return this.api.post<SupplierDeliverySlot>(`${this.base}/dock-slots`, slot);
  }

  // Phase 3: Compliance Documents Vault
  getDocuments(): Observable<SupplierDocument[]> {
    return this.api.get<SupplierDocument[]>(`${this.base}/documents`);
  }

  uploadDocument(doc: Partial<SupplierDocument>): Observable<SupplierDocument> {
    return this.api.post<SupplierDocument>(`${this.base}/documents`, doc);
  }

  // Phase 3: Announcements
  getAnnouncements(): Observable<SupplierAnnouncement[]> {
    return this.api.get<SupplierAnnouncement[]>(`${this.base}/announcements`);
  }
}

export interface SupplierRMA {
  id: string;
  created_at: string;
  rma_number: string;
  purchase_order_id?: string;
  purchase_order?: PurchaseOrder;
  product_id: string;
  product?: {
    id: string;
    name: string;
    sku: string;
  };
  quantity: number;
  reason: string;
  photo_url?: string;
  status: 'pending' | 'approved' | 'replacement_dispatched' | 'refunded' | 'rejected';
  credit_note_ref?: string;
  credit_amount: number;
  resolution_notes?: string;
}

export interface SupplierDeliverySlot {
  id?: string;
  created_at?: string;
  asn_id?: string;
  asn?: SupplierASN;
  slot_date: string;
  time_window: string;
  dock_number: string;
  vehicle_plate?: string;
  driver_phone?: string;
  status: string;
}

export interface SupplierDocument {
  id?: string;
  created_at?: string;
  document_type: string;
  document_name: string;
  file_url: string;
  expiry_date?: string;
  status: 'pending_review' | 'verified' | 'expired' | 'rejected';
  notes?: string;
}

export interface SupplierAnnouncement {
  id: string;
  created_at: string;
  title: string;
  content: string;
  priority: 'info' | 'warning' | 'urgent';
  expires_at?: string;
  is_active: boolean;
}

