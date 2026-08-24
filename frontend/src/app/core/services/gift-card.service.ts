import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

export interface GiftCard {
  id: string;
  code: string;
  initial_balance: number;
  current_balance: number;
  status: 'active' | 'depleted' | 'expired' | 'disabled';
  purchaser_id?: string;
  expires_at?: string;
  created_at: string;
  used_amount?: number;
}

export interface GiftCardCreateInput {
  initial_balance: number;
  purchaser_id?: string;
  expires_at?: string;
  custom_code?: string;
}

@Injectable({
  providedIn: 'root'
})
export class GiftCardService {
  private api = inject(ApiService);

  cards = signal<GiftCard[]>([]);
  loading = signal<boolean>(false);

  getCards(): Observable<GiftCard[]> {
    this.loading.set(true);
    return this.api.get<GiftCard[]>('/gift-cards').pipe(
      tap(res => {
        this.cards.set(res || []);
        this.loading.set(false);
      })
    );
  }

  createCard(input: GiftCardCreateInput): Observable<GiftCard> {
    return this.api.post<GiftCard>('/gift-cards', input).pipe(
      tap(c => this.cards.update(list => [c, ...list]))
    );
  }

  getCard(id: string): Observable<GiftCard> {
    return this.api.get<GiftCard>(`/gift-cards/${id}`);
  }

  disableCard(id: string): Observable<GiftCard> {
    return this.api.post<GiftCard>(`/gift-cards/${id}/disable`, {}).pipe(
      tap(c => this.cards.update(list => list.map(card => card.id === id ? c : card)))
    );
  }

  checkBalance(code: string): Observable<{gift_card: GiftCard}> {
    return this.api.get<{gift_card: GiftCard}>(`/gift-cards/check?code=${code}`);
  }
}
