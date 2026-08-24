import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface ExternalSystem {
  id: string;
  developer_id: string;
  name: string;
  description?: string;
  client_id: string;
  icon: string;
  is_public: boolean;
  is_active: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class MarketplaceService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/apps';

  getApps(): Observable<ExternalSystem[]> {
    return this.http.get<ExternalSystem[]>(this.apiUrl);
  }

  toggleApp(id: string, field: 'active' | 'public'): Observable<any> {
    return this.http.post(`${this.apiUrl}/${id}/toggle?field=${field}`, {});
  }
}
