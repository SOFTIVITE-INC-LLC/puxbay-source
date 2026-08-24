import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { SessionService } from './session.service';

@Injectable({
  providedIn: 'root'
})
export class CheckoutService {
  private http = inject(HttpClient);
  private sessionService = inject(SessionService);
  private apiUrl = '/api/v1/storefront/checkout/verify';

  processCheckout(payload: any): Observable<any> {
    const sessionId = this.sessionService.getSessionId();
    return this.http.post(this.apiUrl, payload, {
      headers: { 'X-Session-ID': sessionId || '' }
    });
  }

  updateCartEmail(email: string): Observable<any> {
    const sessionId = this.sessionService.getSessionId();
    return this.http.put('/api/v1/storefront/cart/email', { email }, {
      headers: { 'X-Session-ID': sessionId || '' }
    });
  }
}
