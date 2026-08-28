import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { Supplier } from '../models/financial.models';

export interface SupplierInput {
  name: string;
  contact_person?: string;
  email?: string;
  phone?: string;
  address?: string;
  website?: string;
  tax_number?: string;
  payment_terms?: string;
  notes?: string;
}

@Injectable({
  providedIn: 'root'
})
export class SupplierService {
  private api = inject(ApiService);

  suppliers = signal<Supplier[]>([]);
  loading = signal<boolean>(false);

  getSuppliers(params?: any): Observable<Supplier[]> {
    this.loading.set(true);
    return this.api.get<Supplier[]>('/suppliers', { params }).pipe(
      tap(res => {
        this.suppliers.set(res || []);
        this.loading.set(false);
      })
    );
  }

  getSupplier(id: string): Observable<Supplier> {
    return this.api.get<Supplier>(`/suppliers/${id}`);
  }

  createSupplier(supplier: SupplierInput): Observable<Supplier> {
    return this.api.post<Supplier>('/suppliers', supplier).pipe(
      tap(s => this.suppliers.update(list => [s, ...list]))
    );
  }

  updateSupplier(id: string, supplier: SupplierInput): Observable<Supplier> {
    return this.api.put<Supplier>(`/suppliers/${id}`, supplier).pipe(
      tap(s => this.suppliers.update(list => list.map(item => item.id === id ? s : item)))
    );
  }

  deleteSupplier(id: string): Observable<void> {
    return this.api.delete<void>(`/suppliers/${id}`).pipe(
      tap(() => this.suppliers.update(list => list.filter(item => item.id !== id)))
    );
  }

  getSupplierProducts(id: string): Observable<any[]> {
    return this.api.get<any[]>(`/suppliers/${id}/products`);
  }

  addSupplierProduct(id: string, product: { product_id: string; supplier_sku: string; unit_cost: number; min_order_qty?: number }): Observable<any> {
    return this.api.post<any>(`/suppliers/${id}/products`, product);
  }

  removeSupplierProduct(supplierID: string, productID: string): Observable<any> {
    return this.api.delete<any>(`/suppliers/${supplierID}/products/${productID}`);
  }

  getSupplierLedger(id: string): Observable<any[]> {
    return this.api.get<any[]>(`/suppliers/${id}/ledger`);
  }

  addSupplierLedger(id: string, entry: {
    entry_type: string;
    amount: number;
    reference_id?: string;
    notes?: string;
    transaction_date?: string;
  }): Observable<any> {
    return this.api.post<any>(`/suppliers/${id}/ledger`, entry);
  }

  inviteToPortal(id: string, data: { email: string; password: string }): Observable<any> {
    return this.api.post<any>(`/suppliers/${id}/invite`, data);
  }

  // Synergy Features
  disburseInvoicePayout(invoiceId: string): Observable<any> {
    return this.api.post<any>(`/suppliers/invoices/${invoiceId}/disburse`, {});
  }

  getAllPriceRequests(status?: string): Observable<any[]> {
    const params = status ? { status } : undefined;
    return this.api.get<any[]>('/suppliers/price-requests', params ? { params } : undefined);
  }

  getAllSupplierInvoices(status?: string): Observable<any[]> {
    const params = status ? { status } : undefined;
    return this.api.get<any[]>('/suppliers/invoices', params ? { params } : undefined);
  }

  approvePriceProposal(requestId: string, notes?: string): Observable<any> {
    return this.api.post<any>(`/suppliers/price-requests/${requestId}/approve`, { notes: notes || '' });
  }

  rejectPriceProposal(requestId: string, notes?: string): Observable<any> {
    return this.api.post<any>(`/suppliers/price-requests/${requestId}/reject`, { notes: notes || '' });
  }

  getRMAs(branchId?: string): Observable<any[]> {
    const params = branchId ? { branch_id: branchId } : undefined;
    return this.api.get<any[]>('/suppliers/rmas', params ? { params } : undefined);
  }

  createRMA(rma: {
    supplier_id: string;
    product_id: string;
    quantity: number;
    reason: string;
    photo_url?: string;
    purchase_order_id?: string;
    branch_id?: string;
  }): Observable<any> {
    return this.api.post<any>('/suppliers/rmas', rma);
  }

  resolveRMA(rmaId: string, payload: { status: string; resolution_notes?: string; credit_amount?: number }): Observable<any> {
    return this.api.post<any>(`/suppliers/rmas/${rmaId}/resolve`, payload);
  }

  createAnnouncement(announcement: { title: string; content: string; priority?: string }): Observable<any> {
    return this.api.post<any>('/suppliers/announcements', announcement);
  }

  getBranchDockSlots(branchId?: string, date?: string): Observable<any[]> {
    const params: any = {};
    if (branchId) params.branch_id = branchId;
    if (date) params.date = date;
    return this.api.get<any[]>('/suppliers/dock-slots', { params });
  }
}

