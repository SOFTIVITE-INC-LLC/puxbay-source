import { Injectable, PLATFORM_ID, inject, signal } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import Dexie, { Table } from 'dexie';

export interface SyncRequest {
  id: string;
  url: string;
  method: string;
  body: any;
  headers: Record<string, string>;
  createdAt: number;
}

export class OfflineDB extends Dexie {
  syncRequests!: Table<SyncRequest, string>;
  apiCache!: Table<{ url: string; data: any; timestamp: number }, string>;

  constructor() {
    super('puxbay-offline-db');
    this.version(1).stores({
      syncRequests: 'id, createdAt',
      apiCache: 'url'
    });
  }
}

@Injectable({
  providedIn: 'root'
})
export class OfflineService {
  private db!: OfflineDB;
  private isBrowser: boolean;
  
  isOnline = signal(true);
  pendingSyncCount = signal(0);

  constructor() {
    this.isBrowser = isPlatformBrowser(inject(PLATFORM_ID));
    if (this.isBrowser) {
      this.db = new OfflineDB();
      this.isOnline.set(navigator.onLine);
      window.addEventListener('online', () => {
        this.isOnline.set(true);
        this.syncPendingRequests();
      });
      window.addEventListener('offline', () => this.isOnline.set(false));
      this.updatePendingCount();
      
      // Attempt sync on startup if online
      if (navigator.onLine) {
        setTimeout(() => this.syncPendingRequests(), 2000);
      }
    }
  }
  
  async syncPendingRequests() {
    if (!this.isBrowser || !this.isOnline()) return;
    
    const requests = await this.getSyncRequests();
    if (requests.length === 0) return;
    
    console.log(`[Offline Sync] Attempting to sync ${requests.length} pending requests...`);
    
    for (const req of requests) {
      try {
        const response = await fetch(req.url, {
          method: req.method,
          headers: {
            'Content-Type': 'application/json',
            ...req.headers
          },
          body: JSON.stringify(req.body)
        });
        
        if (response.ok) {
          await this.removeSyncRequest(req.id);
          console.log(`[Offline Sync] Successfully synced request ${req.id}`);
        } else {
          console.error(`[Offline Sync] Failed to sync request ${req.id}, status: ${response.status}`);
          // Stop syncing on first server error to preserve order if needed
          break;
        }
      } catch (err) {
        console.error(`[Offline Sync] Network error syncing request ${req.id}:`, err);
        break;
      }
    }
    
    this.updatePendingCount();
  }
  
  async updatePendingCount() {
    if (!this.isBrowser) return;
    const count = await this.db.syncRequests.count();
    this.pendingSyncCount.set(count);
  }

  // --- Sync Queue (Offline Writes) ---

  async addSyncRequest(url: string, method: string, body: any, headers: Record<string, string> = {}): Promise<void> {
    if (!this.isBrowser) return;

    const request: SyncRequest = {
      id: crypto.randomUUID(),
      url,
      method,
      body,
      headers,
      createdAt: Date.now()
    };
    
    await this.db.syncRequests.put(request);
    this.updatePendingCount();
    
    // Register background sync if supported
    if ('serviceWorker' in navigator && 'SyncManager' in window) {
      try {
        const registration = await navigator.serviceWorker.ready;
        await (registration as any).sync.register('sync-requests');
      } catch (err) {
        console.error('Background sync registration failed:', err);
      }
    }
  }

  async getSyncRequests(): Promise<SyncRequest[]> {
    if (!this.isBrowser) return [];
    return this.db.syncRequests.orderBy('createdAt').toArray();
  }

  async removeSyncRequest(id: string): Promise<void> {
    if (!this.isBrowser) return;
    await this.db.syncRequests.delete(id);
    this.updatePendingCount();
  }

  // --- API Cache (Offline Reads) ---

  async cacheResponse(url: string, data: any): Promise<void> {
    if (!this.isBrowser) return;
    await this.db.apiCache.put({ url, data, timestamp: Date.now() });
  }

  async getCachedResponse(url: string): Promise<any> {
    if (!this.isBrowser) return null;
    const record = await this.db.apiCache.get(url);
    return record ? record.data : null;
  }
}
