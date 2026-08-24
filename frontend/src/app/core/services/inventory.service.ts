import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../../environments/environment';
import { Observable } from 'rxjs';
import { StockTransfer, PurchaseOrder, StocktakeSession } from '../models/inventory.models';
import { ApiService } from './api.service';
import { inject } from '@angular/core';

export interface TransferCreateInput {
  reference_no: string;
  from_branch_id: string;
  to_branch_id: string;
  notes: string;
  items: { product_id: string; quantity: number }[];
}

export interface POCreateInput {
  po_number: string;
  supplier_id: string;
  branch_id: string;
  notes: string;
  items: { product_id: string; quantity_ordered: number; unit_cost: number }[];
}

@Injectable({
  providedIn: 'root'
})
export class InventoryService {
  private api = inject(ApiService);
  private readonly apiUrl = `${environment.apiUrl}/inventory`;

  constructor(private http: HttpClient) {}

  // ---------------------------------------------------------
  // Transfers
  // ---------------------------------------------------------
  listTransfers(): Observable<StockTransfer[]> {
    return this.http.get<StockTransfer[]>(`${this.apiUrl}/transfers`);
  }

  createTransfer(input: TransferCreateInput): Observable<StockTransfer> {
    return this.http.post<StockTransfer>(`${this.apiUrl}/transfers`, input);
  }

  getTransfer(id: string): Observable<StockTransfer> {
    return this.http.get<StockTransfer>(`${this.apiUrl}/transfers/${id}`);
  }

  approveTransfer(id: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/transfers/${id}/approve`, {});
  }

  shipTransfer(id: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/transfers/${id}/ship`, {});
  }

  receiveTransfer(id: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/transfers/${id}/receive`, {});
  }

  // ---------------------------------------------------------
  // Purchase Orders
  // ---------------------------------------------------------
  listPOs(): Observable<PurchaseOrder[]> {
    return this.http.get<PurchaseOrder[]>(`${this.apiUrl}/purchase-orders`);
  }

  createPO(input: POCreateInput): Observable<PurchaseOrder> {
    return this.http.post<PurchaseOrder>(`${this.apiUrl}/purchase-orders`, input);
  }

  getPO(id: string): Observable<PurchaseOrder> {
    return this.http.get<PurchaseOrder>(`${this.apiUrl}/purchase-orders/${id}`);
  }

  receivePO(id: string, input: { items: { item_id: string; quantity_received: number }[] }): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/purchase-orders/${id}/receive`, input);
  }

  // ---------------------------------------------------------
  // Stocktakes
  // ---------------------------------------------------------
  listStocktakes(): Observable<StocktakeSession[]> {
    return this.http.get<StocktakeSession[]>(`${this.apiUrl}/stocktakes`);
  }

  createStocktake(data: any): Observable<any> { return this.api.post<any>('/inventory/stocktakes', data); }
  
  finalizeStocktake(id: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/stocktakes/${id}/finalize`, {});
  }
  listMovements(): Observable<any[]> { return this.api.get<any[]>('/inventory/movements'); }
  receiveStock(data: any): Observable<any> { return this.api.post<any>('/inventory/receive', data); }
  lowStockAlerts(): Observable<any[]> { return this.api.get<any[]>('/inventory/low-stock'); }

  // ---------------------------------------------------------
  // Batches & Expiry
  // ---------------------------------------------------------
  listBatches(productId: string): Observable<any> {
    return this.api.get<any>(`/inventory/products/${productId}/batches`);
  }

  createBatch(productId: string, data: any): Observable<any> {
    return this.api.post<any>(`/inventory/products/${productId}/batches`, data);
  }

  updateBatch(batchId: string, data: any): Observable<any> {
    return this.api.put<any>(`/inventory/batches/${batchId}`, data);
  }

  deleteBatch(batchId: string): Observable<any> {
    return this.api.delete<any>(`/inventory/batches/${batchId}`);
  }

  listExpiringBatches(branchId?: string, days: number = 30): Observable<any> {
    const params: any = { days };
    if (branchId) params.branch_id = branchId;
    return this.api.get<any>('/inventory/expiring-batches', { params });
  }
}
