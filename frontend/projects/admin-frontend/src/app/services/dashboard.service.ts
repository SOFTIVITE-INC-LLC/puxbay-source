import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class DashboardService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin';

  getSystemHealth(): Observable<{status: string, version: string, latency_ms: number}> {
    return this.http.get<{status: string, version: string, latency_ms: number}>(this.apiUrl + '/health');
  }

  getDashboardStats(): Observable<any> {
    return this.http.get('/api/v1/admin/dashboard');
  }
}
