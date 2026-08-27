import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ProductResponse, ProductsResponse } from '../models/product.model';

@Injectable({
  providedIn: 'root'
})
export class ProductService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/storefront';

  getCategories(branchId?: string): Observable<any[]> {
    const params: any = {};
    if (branchId) params.branch_id = branchId;
    return this.http.get<any[]>(`${this.apiUrl}/categories`, { params });
  }

  getProducts(params?: any): Observable<ProductsResponse> {
    return this.http.get<ProductsResponse>(`${this.apiUrl}/products`, { params });
  }

  searchProducts(query: string): Observable<ProductsResponse> {
    return this.http.get<ProductsResponse>(`${this.apiUrl}/products?${query}`);
  }

  getProduct(id: string): Observable<ProductResponse> {
    return this.http.get<ProductResponse>(`${this.apiUrl}/products/${id}`);
  }

  notifyRestock(id: string, email: string): Observable<any> {
    return this.http.post(`${this.apiUrl}/products/${id}/notify`, { email });
  }

  submitReview(id: string, review: { customer_id: string; rating: number; comment: string }): Observable<any> {
    return this.http.post(`${this.apiUrl}/products/${id}/reviews`, review);
  }
}
