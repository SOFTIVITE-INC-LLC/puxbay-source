import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface FAQ {
  id?: number;
  question: string;
  answer: string;
  order_index: number;
  is_published?: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class FaqService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/faqs';

  getFAQs(): Observable<FAQ[]> {
    return this.http.get<FAQ[]>(this.apiUrl);
  }

  createFAQ(faq: FAQ): Observable<FAQ> {
    return this.http.post<FAQ>(this.apiUrl, faq);
  }

  toggleFAQ(id: number): Observable<any> {
    return this.http.post(`${this.apiUrl}/${id}/toggle`, {});
  }
}
