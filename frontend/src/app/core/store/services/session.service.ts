import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class SessionService {
  private readonly SESSION_KEY = 'storefront_session';

  getSessionId(): string | null {
    if (typeof localStorage === 'undefined') return null;
    return localStorage.getItem(this.SESSION_KEY);
  }

  getOrCreateSessionId(): string {
    if (typeof localStorage === 'undefined') return '';
    let sessionId = localStorage.getItem(this.SESSION_KEY);
    if (!sessionId) {
      sessionId = 'guest_' + Math.random().toString(36).substring(2, 11);
      localStorage.setItem(this.SESSION_KEY, sessionId);
    }
    return sessionId;
  }

  clearSession(): void {
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem(this.SESSION_KEY);
    }
  }
}
