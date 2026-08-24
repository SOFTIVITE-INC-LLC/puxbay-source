import { Injectable, inject, DestroyRef } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { OfflineService, SyncRequest } from '../offline/offline';
import { fromEvent, merge, of } from 'rxjs';
import { map } from 'rxjs/operators';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

@Injectable({
  providedIn: 'root'
})
export class SyncService {
  private http = inject(HttpClient);
  private offlineService = inject(OfflineService);
  private destroyRef = inject(DestroyRef);
  private isSyncing = false;

  constructor() {
    this.initNetworkMonitoring();
  }

  private initNetworkMonitoring() {
    if (typeof window !== 'undefined') {
      const online$ = fromEvent(window, 'online').pipe(map(() => true));
      const offline$ = fromEvent(window, 'offline').pipe(map(() => false));
      
      merge(online$, offline$, of(navigator.onLine)).pipe(
        takeUntilDestroyed(this.destroyRef)
      ).subscribe(isOnline => {
        if (isOnline) {
          console.log('Network online. Attempting to sync background requests.');
          this.syncPendingRequests();
        } else {
          console.log('Network offline. Requests will be queued.');
        }
      });
    }
  }

  async syncPendingRequests() {
    if (this.isSyncing) return;
    this.isSyncing = true;

    try {
      const requests = await this.offlineService.getSyncRequests();
      if (requests.length === 0) {
        this.isSyncing = false;
        return;
      }

      console.log(`Starting sync for ${requests.length} pending requests.`);

      for (const req of requests) {
        try {
          await this.processRequest(req);
          await this.offlineService.removeSyncRequest(req.id);
          console.log(`Successfully synced request ${req.id}`);
        } catch (error) {
          console.error(`Failed to sync request ${req.id}:`, error);
          // Stop syncing on first failure to maintain order, or handle retries specifically
          break;
        }
      }
    } finally {
      this.isSyncing = false;
    }
  }

  private processRequest(req: SyncRequest): Promise<any> {
    let headers = new HttpHeaders();
    if (req.headers) {
      Object.keys(req.headers).forEach(key => {
        headers = headers.append(key, req.headers[key]);
      });
    }

    return new Promise((resolve, reject) => {
      this.http.request(req.method, req.url, {
        body: req.body,
        headers: headers
      }).subscribe({
        next: (res) => resolve(res),
        error: (err) => reject(err)
      });
    });
  }
}
