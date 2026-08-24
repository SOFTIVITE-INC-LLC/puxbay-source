import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { environment } from '../../environments/environment';

export interface TelemetryLog {
  id: string;
  trace_id: string;
  span_id: string;
  name: string;
  start_time: string;
  end_time: string;
  duration_ms: number;
  status: string;
  attributes: any;
  events: any;
  created_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class TelemetryService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin';

  logs = signal<TelemetryLog[]>([]);
  totalLogs = signal<number>(0);
  loading = signal<boolean>(false);

  private ws: WebSocket | null = null;
  liveMode = signal<boolean>(false);

  getTelemetryLogs(page: number = 1, limit: number = 50): Observable<{ data: TelemetryLog[], total: number }> {
    this.loading.set(true);
    const offset = (page - 1) * limit;
    return this.http.get<{ data: TelemetryLog[], total: number }>(`${this.apiUrl}/system-traces?limit=${limit}&offset=${offset}`).pipe(
      tap(res => {
        this.logs.set(res.data || []);
        this.totalLogs.set(res.total || 0);
        this.loading.set(false);
      })
    );
  }

  toggleLiveMode(token: string) {
    if (this.liveMode()) {
      this.disconnectWebSocket();
      this.liveMode.set(false);
    } else {
      this.connectWebSocket(token);
      this.liveMode.set(true);
    }
  }

  private connectWebSocket(token: string) {
    const wsUrl = environment.apiUrl.replace('http', 'ws') + '/ws?token=' + token;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      this.ws?.send(JSON.stringify({ action: 'join', room: 'admin:telemetry' }));
    };

    this.ws.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.type === 'telemetry') {
          // Prepend new log to the list (keep up to 500 logs max to avoid memory bloat)
          this.logs.update(current => {
            const updated = [payload.data, ...current];
            if (updated.length > 500) return updated.slice(0, 500);
            return updated;
          });
          this.totalLogs.update(t => t + 1);
        }
      } catch (e) {
        console.error('Failed to parse telemetry event', e);
      }
    };

    this.ws.onclose = () => {
      this.liveMode.set(false);
    };
  }

  private disconnectWebSocket() {
    if (this.ws) {
      this.ws.send(JSON.stringify({ action: 'leave', room: 'admin:telemetry' }));
      this.ws.close();
      this.ws = null;
    }
  }
}
