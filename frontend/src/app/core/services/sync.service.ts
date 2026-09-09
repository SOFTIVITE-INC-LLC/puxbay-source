import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class SyncService {
  private http = inject(HttpClient);

  constructor() {
    if (typeof window !== 'undefined') {
      this.startBackgroundSync();
    }
  }

  private startBackgroundSync() {
    if (typeof window === 'undefined') return;
    // Poll every 60 seconds to check for offline queue items and sync them
    setInterval(() => {
      if (typeof navigator !== 'undefined' && navigator.onLine) {
        this.syncNow();
      }
    }, 60000);
  }

  async syncNow() {
    if (typeof localStorage === 'undefined') return;
    const queueData = localStorage.getItem('offline_request_queue');
    if (!queueData) return;
    
    try {
      const actions = JSON.parse(queueData);
      
      // Connect to real bulk sync endpoint
      await this.http.post('/api/v1/offline/bulk-sync', { actions }).toPromise();
      
      // Clear queue on success
      localStorage.removeItem('offline_request_queue');
      
      console.log(`Successfully synced ${actions.length} actions.`);
    } catch (error) {
      console.error('Failed to sync offline data:', error);
    }
  }
}
