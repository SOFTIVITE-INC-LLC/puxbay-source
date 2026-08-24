import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface AuditLog {
  id: number;
  tenant_id?: string;
  actor_id?: string;
  action_type: string;
  action?: string;
  severity?: string;
  target_model?: string;
  target_object_id?: string;
  description: string;
  ip_address?: string;
  created_at: string;
  tenant?: {
    name: string;
    subdomain: string;
  };
  actor?: {
    first_name: string;
    last_name: string;
    email: string;
  };
}

export interface AuditLogResponse {
  data: AuditLog[];
  stats: {
    total_events: number;
    critical_errors: number;
    high_risk_actions: number;
  };
}

@Injectable({
  providedIn: 'root'
})
export class AuditService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/audit-logs';

  getAuditLogs(): Observable<AuditLogResponse> {
    return this.http.get<AuditLogResponse>(this.apiUrl);
  }
}
