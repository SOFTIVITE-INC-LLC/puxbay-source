import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap, BehaviorSubject } from 'rxjs';
import { environment } from '../../../../environments/environment';

export interface SupplierProfile {
  id: string;
  name: string;
}

export interface SupplierLoginResponse {
  token: string;
  supplier: SupplierProfile;
}

export interface PurchaseOrder {
  id: string;
  created_at: string;
  po_number: string;
  status: string;
  total_amount: number;
  expected_date?: string;
  items?: any[]; // We can refine this later if needed
  notes?: string;
}

@Injectable({
  providedIn: 'root'
})
export class SupplierPortalService {
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl + '/supplier-portal';

  private currentSupplierSub = new BehaviorSubject<SupplierProfile | null>(null);
  currentSupplier$ = this.currentSupplierSub.asObservable();
  
  loading = signal<boolean>(false);

  login(credentials: { email: string; password: string }): Observable<SupplierLoginResponse> {
    return this.http.post<SupplierLoginResponse>(`${this.apiUrl}/login`, credentials).pipe(
      tap(res => {
        if (res && res.token) {
          localStorage.setItem('supplier_token', res.token);
          this.currentSupplierSub.next(res.supplier);
        }
      })
    );
  }

  logout() {
    localStorage.removeItem('supplier_token');
    this.currentSupplierSub.next(null);
  }

  getMe(): Observable<SupplierProfile> {
    return this.http.get<SupplierProfile>(`${this.apiUrl}/me`).pipe(
      tap(res => this.currentSupplierSub.next(res))
    );
  }

  getPurchaseOrders(): Observable<PurchaseOrder[]> {
    this.loading.set(true);
    return this.http.get<PurchaseOrder[]>(`${this.apiUrl}/purchase-orders`).pipe(
      tap(() => this.loading.set(false))
    );
  }
}
