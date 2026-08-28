import { Injectable, inject, signal } from '@angular/core';
import { Observable, tap, BehaviorSubject } from 'rxjs';
import { ApiService } from '../../../core/services/api.service';
import { SettingsService } from '../../../core/services/settings.service';

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
  currency?: string;
  tenant_currency?: string;
}

export interface SupplierLoginResponse {
  token: string;
  supplier: SupplierProfile;
  currency?: string;
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
  three_way_match_status?: 'matched' | 'mismatched' | 'pending_receipt';
  early_discount_percent?: number;
  early_discount_days?: number;
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
  currency?: string;
  tenant_currency?: string;
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
  private settingsService = inject(SettingsService);
  private readonly base = '/supplier-portal';

  private currentSupplierSub = new BehaviorSubject<SupplierProfile | null>(null);
  currentSupplier$ = this.currentSupplierSub.asObservable();

  loading = signal<boolean>(false);

  private syncCurrency(currency?: string) {
    if (currency) {
      if (typeof window !== 'undefined') {
        localStorage.setItem('currency_code', currency);
      }
      this.settingsService.settings.update(s => ({
        ...(s || {
          timezone: 'UTC',
          date_format: 'YYYY-MM-DD',
          enable_email_receipts: false,
          hardware_proxy_url: '',
          enable_hardware_proxy: false,
          auto_print_receipts: false,
          enable_sms_notifications: false,
          enable_push_notifications: false,
          admin_notification_email: '',
          promo_threshold: 0,
          promo_discount_percent: 0,
        }),
        currency: currency
      }));
    }
  }

  login(credentials: { email: string; password: string }): Observable<SupplierLoginResponse> {
    return this.api.post<SupplierLoginResponse>(`${this.base}/login`, credentials).pipe(
      tap(res => {
        if (res?.token) {
          localStorage.setItem('supplier_token', res.token);
          this.currentSupplierSub.next(res.supplier);
          this.syncCurrency(res.currency || res.supplier?.currency || res.supplier?.tenant_currency);
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
      tap(res => {
        this.currentSupplierSub.next(res);
        this.syncCurrency(res?.currency || res?.tenant_currency);
      })
    );
  }

  getDashboard(): Observable<DashboardStats> {
    return this.api.get<DashboardStats>(`${this.base}/dashboard`).pipe(
      tap(res => {
        this.syncCurrency(res?.currency || res?.tenant_currency);
      })
    );
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

  // Enterprise Suite: GS1-128 Shipping Label
  getShippingLabel(asnId: string): Observable<ShippingLabelData> {
    return this.api.get<ShippingLabelData>(`${this.base}/shipments/${asnId}/shipping-label`);
  }

  // Enterprise Suite: Bulk CSV Catalog Import
  bulkImportCatalog(items: Array<{ sku: string; product_name: string; unit_cost: number; min_order_qty: number }>): Observable<{ message: string; imported_count: number }> {
    return this.api.post<{ message: string; imported_count: number }>(`${this.base}/catalog/bulk`, { items });
  }

  // Enterprise Suite: 3-Way Matching Audit
  getThreeWayMatch(invoiceId: string): Observable<ThreeWayMatchAudit> {
    return this.api.get<ThreeWayMatchAudit>(`${this.base}/invoices/${invoiceId}/three-way-match`);
  }

  // Enterprise Suite: Scorecard & Vendor Tier
  getScorecard(): Observable<SupplierScorecard> {
    return this.api.get<SupplierScorecard>(`${this.base}/scorecard`);
  }

  // Enterprise Suite: Developer API Keys
  listApiKeys(): Observable<SupplierAPIKey[]> {
    return this.api.get<SupplierAPIKey[]>(`${this.base}/api-keys`);
  }

  createApiKey(name: string): Observable<{ api_key: SupplierAPIKey; plain_key: string; message: string }> {
    return this.api.post<{ api_key: SupplierAPIKey; plain_key: string; message: string }>(`${this.base}/api-keys`, { name });
  }

  revokeApiKey(id: string): Observable<{ message: string }> {
    return this.api.delete<{ message: string }>(`${this.base}/api-keys/${id}`);
  }

  // Enterprise Suite: Webhooks Management
  listWebhooks(): Observable<SupplierWebhook[]> {
    return this.api.get<SupplierWebhook[]>(`${this.base}/webhooks`);
  }

  createWebhook(url: string, events: string): Observable<SupplierWebhook> {
    return this.api.post<SupplierWebhook>(`${this.base}/webhooks`, { url, events });
  }

  deleteWebhook(id: string): Observable<{ message: string }> {
    return this.api.delete<{ message: string }>(`${this.base}/webhooks/${id}`);
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

export interface ShippingLabelData {
  asn_number: string;
  po_number: string;
  carrier: string;
  tracking_number?: string;
  dispatch_date: string;
  expected_date?: string;
  package_count: number;
  total_weight_kg: number;
  barcode: string;
  vendor_name: string;
  vendor_phone?: string;
  vendor_address?: string;
  vendor_tin?: string;
  items?: PurchaseOrderItem[];
}

export interface ThreeWayMatchAudit {
  invoice_id: string;
  invoice_number: string;
  invoice_total: number;
  po_number: string;
  po_total: number;
  match_status: 'matched' | 'mismatched' | 'pending_receipt';
  discrepancy_notes: string;
  items: Array<{
    product_name: string;
    sku: string;
    quantity_ordered: number;
    quantity_received: number;
    quantity_invoiced: number;
    unit_cost_agreed: number;
    unit_cost_invoiced: number;
    is_matched: boolean;
  }>;
  early_discount_summary?: {
    discount_percent: number;
    discount_days: number;
    eligible_until: string;
    is_eligible: boolean;
    discount_amount: number;
    net_payable: number;
  };
}

export interface TierLevel {
  tier: 'platinum' | 'gold' | 'silver' | 'standard';
  name: string;
  badge: string;
  min_otd: number;
  max_defect: number;
  min_pos: number;
  settlement: string;
  perks: string;
}

export interface QuarterlyReview {
  period: string;
  total_pos: number;
  on_time_pct: number;
  fill_rate: number;
  defect_rate: number;
  score: number;
  status: string;
}

export interface SupplierScorecard {
  supplier_id: string;
  name: string;
  tier: 'platinum' | 'gold' | 'silver' | 'standard';
  otd_rate: number;
  defect_rate: number;
  fill_rate?: number;
  avg_lead_time_days?: number;
  match_accuracy_rate?: number;
  total_pos: number;
  total_rmas: number;
  benefits: string[];
  tier_progress?: {
    current_tier: string;
    next_tier: string;
    progress_pct: number;
    points_needed: number;
    target_otd: number;
    target_defect: number;
  };
  tier_levels?: TierLevel[];
  quarterly_history?: QuarterlyReview[];
  sla_benchmarks?: {
    order_acknowledgment_pct: number;
    asn_dispatch_pct: number;
    match_accuracy_pct: number;
    packaging_compliance_pct: number;
  };
}

export interface SupplierAPIKey {
  id: string;
  created_at: string;
  name: string;
  key_preview: string;
  last_used_at?: string;
  expires_at?: string;
  is_active: boolean;
}

export interface SupplierWebhook {
  id: string;
  created_at: string;
  url: string;
  events: string;
  secret: string;
  is_active: boolean;
  last_triggered_at?: string;
  last_status?: number;
}

