import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface WebSocketMessage {
  type: string;
  payload: any;
}

@Injectable({
  providedIn: 'root'
})
export class WebsocketService {
  private socket: WebSocket | null = null;
  private reconnectInterval = 3000;
  private reconnectAttempts = 0;
  
  public messages$ = new Subject<WebSocketMessage>();

  constructor() {
    this.connect();
  }

  private connect() {
    if (!window.localStorage.getItem('token')) {
      return; // Only connect if authenticated
    }

    let wsUrl: string;
    const apiUrl = environment.apiUrl;

    if (apiUrl.startsWith('http')) {
      // Absolute URL: convert http(s) to ws(s) and append /ws
      wsUrl = apiUrl.replace(/^https:/, 'wss:').replace(/^http:/, 'ws:') + '/ws';
    } else {
      // Relative URL: use current host
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      wsUrl = `${wsProtocol}//${window.location.host}${apiUrl}/ws`;
    }

    this.socket = new WebSocket(wsUrl);

    this.socket.onopen = () => {
      console.log('[WebSocket] Connected');
      this.reconnectAttempts = 0; // Reset on success
      // Send auth token to upgrade connection if required by custom hub logic
      const token = localStorage.getItem('token');
      this.socket?.send(JSON.stringify({ type: 'auth', token }));
    };

    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.messages$.next(data);

        // Native Browser Push Notification
        if (data.type === 'notification' && 'Notification' in window && Notification.permission === 'granted') {
          try {
            new Notification(data.title || 'Puxbay Notification', {
              body: data.message,
              icon: '/assets/icons/icon-192x192.png',
            });
          } catch (_) {}
        }
      } catch (e) {
        console.error('[WebSocket] Failed to parse message', e);
      }
    };

    this.socket.onclose = () => {
      // Exponential backoff with jitter: 3s, 6s, 12s, max 30s
      let delay = this.reconnectInterval * Math.pow(2, this.reconnectAttempts);
      if (delay > 30000) delay = 30000;
      
      // Add jitter (-500ms to +500ms)
      delay += (Math.random() * 1000) - 500;
      
      console.log(`[WebSocket] Disconnected. Reconnecting in ${Math.round(delay)}ms... (Attempt ${this.reconnectAttempts + 1})`);
      
      this.reconnectAttempts++;
      setTimeout(() => this.connect(), delay);
    };

    this.socket.onerror = (error) => {
      console.error('[WebSocket] Error', error);
      this.socket?.close();
    };
  }

  public sendMessage(msg: WebSocketMessage) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(msg));
    } else {
      console.warn('[WebSocket] Cannot send message, socket not open');
    }
  }
}
