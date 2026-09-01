import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { Order } from '../models/order.models';

export interface OrderListResponse {
  data: Order[];
  total: number;
  page: number;
  limit: number;
}

export interface OrderItemInput {
  product_id: string;
  variant_id?: string;
  quantity: number;
  unit_price: number;
  discount: number;
  total: number;
}

export interface OrderCreateInput {
  branch_id?: string;
  customer_id?: string;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  amount_paid: number;
  payment_method: string;
  order_type: string;
  notes?: string;
  items: OrderItemInput[];
}

@Injectable({
  providedIn: 'root'
})
export class OrderService {
  private api = inject(ApiService);
  
  orders = signal<Order[]>([]);
  loading = signal<boolean>(false);
  totalOrders = signal<number>(0);

  getOrders(params?: any): Observable<OrderListResponse> {
    this.loading.set(true);
    return this.api.get<OrderListResponse>('/orders', { params }).pipe(
      tap(res => {
        if (params?.page > 1) {
          this.orders.update(prev => [...prev, ...(res.data || [])]);
        } else {
          this.orders.set(res.data || []);
        }
        this.totalOrders.set(res.total || 0);
        this.loading.set(false);
      })
    );
  }

  getOrder(id: string): Observable<Order> {
    return this.api.get<Order>(`/orders/${id}`);
  }

  createOrder(order: OrderCreateInput): Observable<Order> {
    return this.api.post<Order>('/orders', order).pipe(
      tap(o => this.orders.update(list => [o, ...list]))
    );
  }

  processPOSCheckout(order: OrderCreateInput): Observable<Order> {
    return this.api.post<Order>('/orders/pos', order).pipe(
      tap(o => this.orders.update(list => [o, ...list]))
    );
  }

  voidOrder(id: string, overridePin?: string): Observable<any> { 
    let headers: any = {};
    if (overridePin) {
      headers['X-Manager-Override-PIN'] = overridePin;
    }
    return this.api.post<any>(`/orders/${id}/void`, {}, { headers }); 
  }

  completeOrder(id: string, overridePin?: string): Observable<any> {
    return this.api.post<any>(`/orders/${id}/complete`, { override_pin: overridePin });
  }

  sendPickupOTP(id: string): Observable<{ status: string; phone: string; masked: string; message: string }> {
    return this.api.post<{ status: string; phone: string; masked: string; message: string }>(`/orders/${id}/send-pickup-otp`, {});
  }

  verifyPickupOTP(id: string, otp: string): Observable<{ status: string; message: string }> {
    return this.api.post<{ status: string; message: string }>(`/orders/${id}/verify-pickup-otp`, { otp });
  }

  getReceipt(id: string): Observable<any> { return this.api.get<any>(`/orders/${id}/receipt`); }
}
