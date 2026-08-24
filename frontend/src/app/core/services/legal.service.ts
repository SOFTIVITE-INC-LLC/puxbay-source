import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export type LegalDocType = 'terms' | 'privacy' | 'cookie';

export interface LegalDocument {
  id: string;
  type: LegalDocType;
  title: string;
  content: string;
  version: string;
  effective_date?: string;
  updated_at?: string;
}

@Injectable({ providedIn: 'root' })
export class LegalService {
  private http = inject(HttpClient);

  getLegalDocument(type: LegalDocType): Observable<LegalDocument> {
    return this.http.get<LegalDocument>(`${environment.apiUrl}/public/legal/${type}`);
  }
}
