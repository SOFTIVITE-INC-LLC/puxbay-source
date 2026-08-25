import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { StockTransfer, PurchaseOrder, StocktakeSession } from '../models/inventory.models';
import { ApiService } from './api.service';

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

  // ---------------------------------------------------------
  // Transfers
  // ---------------------------------------------------------
  listTransfers(): Observable<StockTransfer[]> {
    return this.api.get<StockTransfer[]>('/inventory/transfers');
  }

  createTransfer(input: TransferCreateInput): Observable<StockTransfer> {
    return this.api.post<StockTransfer>('/inventory/transfers', input);
  }

  getTransfer(id: string): Observable<StockTransfer> {
    return this.api.get<StockTransfer>(`/inventory/transfers/${id}`);
  }

  approveTransfer(id: string): Observable<any> {
    return this.api.post<any>(`/inventory/transfers/${id}/approve`, {});
  }

  shipTransfer(id: string): Observable<any> {
    return this.api.post<any>(`/inventory/transfers/${id}/ship`, {});
  }

  receiveTransfer(id: string): Observable<any> {
    return this.api.post<any>(`/inventory/transfers/${id}/receive`, {});
  }

  // ---------------------------------------------------------
  // Purchase Orders
  // ---------------------------------------------------------
  listPOs(): Observable<PurchaseOrder[]> {
    return this.api.get<PurchaseOrder[]>('/inventory/purchase-orders');
  }

  createPO(input: POCreateInput): Observable<PurchaseOrder> {
    return this.api.post<PurchaseOrder>('/inventory/purchase-orders', input);
  }

  getPO(id: string): Observable<PurchaseOrder> {
    return this.api.get<PurchaseOrder>(`/inventory/purchase-orders/${id}`);
  }

  receivePO(id: string, input: { items: { item_id: string; quantity_received: number }[] }): Observable<any> {
    return this.api.post<any>(`/inventory/purchase-orders/${id}/receive`, input);
  }

  // ---------------------------------------------------------
  // Stocktakes
  // ---------------------------------------------------------
  listStocktakes(): Observable<StocktakeSession[]> {
    return this.api.get<StocktakeSession[]>('/inventory/stocktakes');
  }

  getStocktake(id: string): Observable<any> {
    return this.api.get<any>(`/inventory/stocktakes/${id}`);
  }

  createStocktake(data: any): Observable<any> {
    return this.api.post<any>('/inventory/stocktakes', data);
  }

  finalizeStocktake(id: string): Observable<any> {
    return this.api.post<any>(`/inventory/stocktakes/${id}/finalize`, {});
  }

  // ---------------------------------------------------------
  // Movements & Stock
  // ---------------------------------------------------------
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
