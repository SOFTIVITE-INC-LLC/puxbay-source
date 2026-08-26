import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface GiftCard {
  id: string;
  code: string;
  initial_balance: number;
  current_balance: number;
  status: string;
  is_active: boolean;
  expires_at?: string;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class GiftCardService {
  private http = inject(HttpClient);
  private base = '/api/v1/admin/gift-cards';

  getGiftCards(page: number = 1): Observable<{ data: GiftCard[], total: number }> {
    return this.http.get<{ data: GiftCard[], total: number }>(`${this.base}?page=${page}`);
  }

  createGiftCard(data: { initial_balance: number, custom_code?: string, expires_at?: string }): Observable<GiftCard> {
    return this.http.post<GiftCard>(this.base, data);
  }

  disableGiftCard(id: string): Observable<GiftCard> {
    return this.http.post<GiftCard>(`${this.base}/${id}/disable`, {});
  }
}
