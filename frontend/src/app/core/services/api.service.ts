import { Injectable, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformServer } from '@angular/common';
import { HttpClient, HttpHeaders, HttpParams, HttpErrorResponse } from '@angular/common/http';
import { Observable, tap, catchError, of, throwError, from, switchMap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { OfflineService } from './offline/offline';

export interface ApiOptions {
  headers?: HttpHeaders | { [header: string]: string | string[] };
  params?: HttpParams | { [param: string]: string | string[] };
  responseType?: 'json' | 'text' | 'blob' | 'arraybuffer';
}

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private http = inject(HttpClient);
  private offlineService = inject(OfflineService);
  private platformId = inject(PLATFORM_ID);
  private apiUrl = isPlatformServer(this.platformId) && environment.production ? 'http://localhost:5000/api/v1' : environment.apiUrl;
  private cache = new Map<string, { data: any, timestamp: number }>();
  private CACHE_TTL = 5000;

  get<T>(path: string, options?: ApiOptions, bypassCache = false): Observable<T> {
    const url = `${this.apiUrl}${path}`;
    const cacheKey = url + JSON.stringify(options?.params || {});
    
    if (!bypassCache) {
      const cached = this.cache.get(cacheKey);
      if (cached && (Date.now() - cached.timestamp < this.CACHE_TTL)) {
        return of(cached.data as T);
      }
    }

    return this.http.get<T>(url, { ...(options as any), withCredentials: true }).pipe(
      tap((data: any) => {
        if (!bypassCache && options?.responseType !== 'blob' && options?.responseType !== 'arraybuffer') {
          this.cache.set(cacheKey, { data, timestamp: Date.now() });
          this.offlineService.cacheResponse(cacheKey, data);
        }
      }),
      catchError((error: HttpErrorResponse) => {
        if (error.status === 0 || !this.offlineService.isOnline()) {
          return from(this.offlineService.getCachedResponse(cacheKey)).pipe(
            switchMap(cachedData => {
              if (cachedData) return of(cachedData as T);
              return throwError(() => error);
            })
          );
        }
        return throwError(() => error);
      })
    );
  }

  post<T>(path: string, body: any, options?: ApiOptions): Observable<T> {
    const url = `${this.apiUrl}${path}`;
    return (this.http.post<T>(url, body, { ...(options as any), withCredentials: true }) as Observable<T>).pipe(
      tap(() => this.cache.clear()),
      catchError((error: HttpErrorResponse) => this.handleOfflineMutation<T>(error, url, 'POST', body, options))
    );
  }

  put<T>(path: string, body: any, options?: ApiOptions): Observable<T> {
    const url = `${this.apiUrl}${path}`;
    return (this.http.put<T>(url, body, { ...(options as any), withCredentials: true }) as Observable<T>).pipe(
      tap(() => this.cache.clear()),
      catchError((error: HttpErrorResponse) => this.handleOfflineMutation<T>(error, url, 'PUT', body, options))
    );
  }

  patch<T>(path: string, body: any, options?: ApiOptions): Observable<T> {
    const url = `${this.apiUrl}${path}`;
    return (this.http.patch<T>(url, body, { ...(options as any), withCredentials: true }) as Observable<T>).pipe(
      tap(() => this.cache.clear()),
      catchError((error: HttpErrorResponse) => this.handleOfflineMutation<T>(error, url, 'PATCH', body, options))
    );
  }

  delete<T>(path: string, options?: ApiOptions): Observable<T> {
    const url = `${this.apiUrl}${path}`;
    return (this.http.delete<T>(url, { ...(options as any), withCredentials: true }) as Observable<T>).pipe(
      tap(() => this.cache.clear()),
      catchError((error: HttpErrorResponse) => this.handleOfflineMutation<T>(error, url, 'DELETE', null, options))
    );
  }

  private handleOfflineMutation<T>(error: HttpErrorResponse, url: string, method: string, body: any, options?: ApiOptions): Observable<T> {
    if (error.status === 0 || !this.offlineService.isOnline()) {
      let headersRecord: Record<string, string> = {};
      
      if (options?.headers) {
        if (options.headers instanceof HttpHeaders) {
          options.headers.keys().forEach(key => {
            headersRecord[key] = (options.headers as HttpHeaders).get(key) || '';
          });
        } else {
          headersRecord = options.headers as Record<string, string>;
        }
      }

      this.offlineService.addSyncRequest(url, method, body, headersRecord);
      
      // Return optimistic success response
      const optimisticResponse = { success: true, offline: true, ...body } as unknown as T;
      return of(optimisticResponse);
    }
    return throwError(() => error);
  }
}
