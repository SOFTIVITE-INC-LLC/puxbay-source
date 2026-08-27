import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface SMSGatewayConfig {
  id?: string;
  provider: string;
  default_sender_id: string;
  price_per_sms: number;
  price_currency: string;
  is_active: boolean;
}

export interface AdminSenderIDEntry {
  id: string;
  tenant_id: string;
  tenant_name: string;
  sender_id: string;
  purpose: string;
  status: 'pending' | 'approved' | 'rejected';
  rejection_reason?: string;
  approved_at?: string;
  created_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class AdminSMSService {
  private http = inject(HttpClient);
  private base = '/api/v1/admin/sms';

  getConfig(): Observable<SMSGatewayConfig> {
    return this.http.get<SMSGatewayConfig>(`${this.base}/config`);
  }

  updateConfig(cfg: Partial<SMSGatewayConfig>): Observable<any> {
    return this.http.put<any>(`${this.base}/config`, cfg);
  }

  getSenderIDs(status?: string): Observable<AdminSenderIDEntry[]> {
    const params: any = {};
    if (status) params['status'] = status;
    return this.http.get<AdminSenderIDEntry[]>(`${this.base}/sender-ids`, { params });
  }

  approveSenderID(id: string, schema: string): Observable<any> {
    return this.http.post<any>(`${this.base}/sender-ids/${id}/approve?schema=${schema}`, {});
  }

  rejectSenderID(id: string, schema: string, reason: string): Observable<any> {
    return this.http.post<any>(`${this.base}/sender-ids/${id}/reject?schema=${schema}`, { reason });
  }
}
