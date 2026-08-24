import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class DeliveryService {
  private api = inject(ApiService);
  
  drivers = signal<any[]>([]);

  // Get all drivers
  getDrivers(): Observable<any[]> {
    return this.api.get<any[]>('/delivery/drivers').pipe(
      tap(res => this.drivers.set(res || []))
    );
  }

  // Create driver
  addDriver(driver: { name: string; phone: string; vehicle_info: string }): Observable<any> {
    return this.api.post<any>('/delivery/drivers', driver).pipe(
      tap(res => this.drivers.update(d => [...d, res]))
    );
  }

  // Dispatch an order
  dispatchOrder(data: { order_id: string; driver_id: string; delivery_fee: number; delivery_notes: string }): Observable<any> {
    return this.api.post<any>('/delivery/dispatch', data);
  }

  // Get active dispatched orders
  getActiveDeliveries(): Observable<any[]> {
    return this.api.get<any[]>('/delivery/orders');
  }

  // Mark a delivery as completed
  completeDelivery(id: string): Observable<any> {
    return this.api.post<any>(`/delivery/orders/${id}/complete`, {});
  }
}
