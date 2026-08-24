import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface PromoCode {
  id?: string;
  code: string;
  discount_type: 'percentage' | 'fixed';
  discount_value: number;
  max_uses: number;
  current_uses?: number;
  is_active?: boolean;
  valid_from?: string;
  valid_until?: string | null;
}

export interface PromoCodeResponse {
  data: PromoCode[];
  stats: {
    active_codes: number;
    total_redemptions: number;
    top_code: string;
  };
}

@Injectable({
  providedIn: 'root'
})
export class PromoCodeService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/promo-codes';

  getPromoCodes(): Observable<PromoCodeResponse> {
    return this.http.get<PromoCodeResponse>(this.apiUrl);
  }

  createPromoCode(code: PromoCode): Observable<PromoCode> {
    return this.http.post<PromoCode>(this.apiUrl, code);
  }

  togglePromoCode(id: string): Observable<any> {
    return this.http.post(`${this.apiUrl}/${id}/toggle`, {});
  }
}
