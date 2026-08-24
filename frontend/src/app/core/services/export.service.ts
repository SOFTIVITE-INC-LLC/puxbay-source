import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from './api.service';

import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ExportService {
  private api = inject(ApiService);
  private apiUrl = environment.apiUrl; // We need the base URL to construct download links

  // Using window.open is usually the simplest way to trigger a file download from a GET endpoint
  
  exportOrders(params?: { branch_id?: string, start_date?: string, end_date?: string }): void {
    const url = new URL(`${this.apiUrl}/export/orders`);
    if (params?.branch_id) url.searchParams.append('branch_id', params.branch_id);
    if (params?.start_date) url.searchParams.append('start_date', params.start_date);
    if (params?.end_date) url.searchParams.append('end_date', params.end_date);
    
    this.downloadFile(url.toString());
  }

  exportProducts(branchId?: string): void {
    const url = new URL(`${this.apiUrl}/export/products`);
    if (branchId) url.searchParams.append('branch_id', branchId);
    
    this.downloadFile(url.toString());
  }

  exportInventory(branchId?: string): void {
    const url = new URL(`${this.apiUrl}/export/inventory`);
    if (branchId) url.searchParams.append('branch_id', branchId);
    
    this.downloadFile(url.toString());
  }

  exportCustomers(): void {
    const url = new URL(`${this.apiUrl}/export/customers`);
    this.downloadFile(url.toString());
  }

  private downloadFile(url: string): void {
    // We can't easily attach the Authorization header if we use window.open directly
    // since the browser handles the GET request.
    // However, if the API expects an auth token via cookies, window.open works.
    // For standard token auth, we usually fetch the blob and create an object URL.
    
    this.api.get<Blob>(url.replace(this.apiUrl, ''), { responseType: 'blob' } as any).subscribe(blob => {
      const downloadUrl = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = downloadUrl;
      // Extract filename from headers would be ideal, but for simplicity we rely on the browser
      // or we could hardcode generic names here.
      a.download = 'export.csv'; 
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(downloadUrl);
      a.remove();
    });
  }

  exportSalesCSV(): Observable<any> { return this.api.get<any>('/export/sales'); }
  exportInventoryCSV(): Observable<any> { return this.api.get<any>('/export/inventory'); }
  exportCustomersCSV(): Observable<any> { return this.api.get<any>('/export/customers'); }
  exportOrderItemsCSV(): Observable<any> { return this.api.get<any>('/export/order-items'); }
}
