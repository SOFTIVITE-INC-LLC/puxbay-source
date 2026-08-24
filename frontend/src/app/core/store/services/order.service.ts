import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface OrderTrackResponse {
  order: any; // Add specific order type if needed
}

@Injectable({
  providedIn: 'root'
})
export class OrderService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/storefront/orders';

  trackOrder(orderNumber: string): Observable<OrderTrackResponse> {
    return this.http.get<OrderTrackResponse>(`${this.apiUrl}/track?order_number=${encodeURIComponent(orderNumber)}`);
  }
}
