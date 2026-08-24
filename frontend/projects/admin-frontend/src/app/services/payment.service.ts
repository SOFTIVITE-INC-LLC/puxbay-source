import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Payment {
  id: string;
  subscription_id: string;
  amount: number;
  paystack_reference?: string;
  status: string;
  date: string;
  subscription?: {
    tenant?: {
      name: string;
    }
  };
}

export interface PaymentResponse {
  data: Payment[];
  stats: {
    total_revenue: number;
    successful_count: number;
    failed_count: number;
  };
}

@Injectable({
  providedIn: 'root'
})
export class PaymentService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/payments';

  getPayments(): Observable<PaymentResponse> {
    return this.http.get<PaymentResponse>(this.apiUrl);
  }
}
