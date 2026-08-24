import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Broadcast {
  id?: string;
  title: string;
  message: string;
  type: string;
  target_audience?: string;
  created_at?: string;
  created_by?: string;
}

@Injectable({
  providedIn: 'root'
})
export class BroadcastService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/broadcasts';

  getBroadcasts(): Observable<Broadcast[]> {
    return this.http.get<Broadcast[]>(this.apiUrl);
  }

  createBroadcast(broadcast: Broadcast): Observable<Broadcast> {
    return this.http.post<Broadcast>(this.apiUrl, broadcast);
  }
}
