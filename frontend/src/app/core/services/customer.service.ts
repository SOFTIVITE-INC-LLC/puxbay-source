import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Customer } from '../models/models';
import { Observable, tap, from } from 'rxjs';
import { OfflineService } from './offline.service';

@Injectable({
  providedIn: 'root'
})
export class CustomerService {
  private api = inject(ApiService);
  private offlineService = inject(OfflineService);
  
  customers = signal<Customer[]>([]);
  total = signal<number>(0);
  loading = signal<boolean>(false);

  getCustomers(params?: any): Observable<{data: Customer[], total: number, page: number, limit: number}> {
    this.loading.set(true);
    return this.api.get<{data: Customer[], total: number, page: number, limit: number}>('/customers', { params }).pipe(
      tap(res => {
        this.customers.set(res.data || []);
        this.total.set(res.total || 0);
        this.loading.set(false);
      })
    );
  }

  getCustomer(id: string): Observable<Customer> {
    return this.api.get<Customer>(`/customers/${id}`);
  }

  createCustomer(customer: Partial<Customer>): Observable<Customer> {
    return this.api.post<Customer>('/customers', customer).pipe(
      tap(newCust => this.customers.update(custs => [newCust, ...custs]))
    );
  }

  updateCustomer(id: string, customer: Partial<Customer>): Observable<Customer> {
    return this.api.put<Customer>(`/customers/${id}`, customer).pipe(
      tap(updatedCust => this.customers.update(custs => custs.map(c => c.id === id ? updatedCust : c)))
    );
  }

  deleteCustomer(id: string): Observable<void> {
    return this.api.delete<void>(`/customers/${id}`).pipe(
      tap(() => this.customers.update(custs => custs.filter(c => c.id !== id)))
    );
  }
}
