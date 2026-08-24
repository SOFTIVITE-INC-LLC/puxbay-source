import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

export interface PaymentMethod {
  id: string;
  name: string;
  provider: string;
  is_active: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class PaymentMethodService {
  private api = inject(ApiService);
  
  methods = signal<PaymentMethod[]>([]);
  loading = signal<boolean>(false);

  getMethods(): Observable<{payment_methods: PaymentMethod[]}> {
    this.loading.set(true);
    return this.api.get<{payment_methods: PaymentMethod[]}>('/payment-methods').pipe(
      tap(res => {
        this.methods.set(res.payment_methods || []);
        this.loading.set(false);
      })
    );
  }
}
