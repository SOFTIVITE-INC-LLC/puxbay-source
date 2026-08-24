import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface DatabaseBackup {
  id: number;
  filename: string;
  file_path: string;
  size_bytes: number;
  created_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class BackupService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/backups';

  getBackups(): Observable<DatabaseBackup[]> {
    return this.http.get<DatabaseBackup[]>(this.apiUrl);
  }

  triggerBackup(): Observable<DatabaseBackup> {
    return this.http.post<DatabaseBackup>(`${this.apiUrl}/trigger`, {});
  }

  downloadBackup(id: number): Observable<Blob> {
    return this.http.get(`${this.apiUrl}/${id}/download`, {
      responseType: 'blob'
    });
  }
}
