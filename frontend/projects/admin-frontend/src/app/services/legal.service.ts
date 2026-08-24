import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export type LegalDocType = 'terms' | 'privacy' | 'cookie';

export interface LegalDocument {
  id?: string;
  type: LegalDocType;
  title: string;
  content: string;
  version: string;
  effective_date?: string;
  created_at?: string;
  updated_at?: string;
}

export interface LegalDocumentsResponse {
  documents: LegalDocument[];
}

@Injectable({ providedIn: 'root' })
export class LegalService {
  private http = inject(HttpClient);
  private base = '/api/v1/admin/legal';

  getLegalDocuments(): Observable<LegalDocumentsResponse> {
    return this.http.get<LegalDocumentsResponse>(this.base);
  }

  upsertLegalDocument(type: LegalDocType, payload: Partial<LegalDocument>): Observable<LegalDocument> {
    return this.http.put<LegalDocument>(`${this.base}/${type}`, payload);
  }
}
