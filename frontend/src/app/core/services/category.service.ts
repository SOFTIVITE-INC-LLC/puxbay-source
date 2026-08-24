import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { Category } from '../models/product.models';

export interface CategoryCreateInput {
  name: string;
  description?: string;
  color?: string;
}

@Injectable({
  providedIn: 'root'
})
export class CategoryService {
  private api = inject(ApiService);
  private readonly baseUrl = '/categories';
  
  categories = signal<Category[]>([]);
  loading = signal<boolean>(false);

  getCategories(params?: any): Observable<Category[]> {
    this.loading.set(true);
    return this.api.get<Category[]>(this.baseUrl, { params }).pipe(
      tap(res => {
        this.categories.set(res || []);
        this.loading.set(false);
      })
    );
  }

  getCategory(id: string): Observable<Category> {
    return this.api.get<Category>(`${this.baseUrl}/${id}`);
  }

  createCategory(input: CategoryCreateInput): Observable<Category> {
    return this.api.post<Category>(this.baseUrl, input).pipe(
      tap(c => this.categories.update(list => [...list, c]))
    );
  }

  updateCategory(id: string, input: CategoryCreateInput): Observable<Category> {
    return this.api.put<Category>(`${this.baseUrl}/${id}`, input).pipe(
      tap(c => this.categories.update(list => list.map(item => item.id === c.id ? c : item)))
    );
  }

  deleteCategory(id: string): Observable<void> {
    return this.api.delete<void>(`${this.baseUrl}/${id}`).pipe(
      tap(() => this.categories.update(list => list.filter(item => item.id !== id)))
    );
  }
}
