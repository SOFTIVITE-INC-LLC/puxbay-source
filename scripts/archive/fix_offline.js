const fs = require('fs');

const code = `import { Injectable, signal } from '@angular/core';
import { openDB, IDBPDatabase } from 'idb';

@Injectable({
  providedIn: 'root'
})
export class OfflineService {
  isOnline = signal<boolean>(navigator.onLine);
  pendingSyncCount = signal<number>(0);
  private db!: IDBPDatabase;

  constructor() {
    this.initDB();
    this.setupListeners();
  }

  private async initDB() {
    this.db = await openDB('puxbay-offline-db', 2, {
      upgrade(db) {
        if (!db.objectStoreNames.contains('orders')) {
          db.createObjectStore('orders', { keyPath: 'id', autoIncrement: true });
        }
        if (!db.objectStoreNames.contains('products')) {
          db.createObjectStore('products', { keyPath: 'id' });
        }
        if (!db.objectStoreNames.contains('categories')) {
          db.createObjectStore('categories', { keyPath: 'id' });
        }
        if (!db.objectStoreNames.contains('customers')) {
          db.createObjectStore('customers', { keyPath: 'id' });
        }
      },
    });
    this.updatePendingCount();
  }

  private setupListeners() {
    window.addEventListener('online', () => {
      this.isOnline.set(true);
      this.syncQueuedOrders();
    });
    
    window.addEventListener('offline', () => {
      this.isOnline.set(false);
    });
  }

  async queueOrder(order: any) {
    if (!this.db) await this.initDB();
    await this.db.add('orders', { ...order, queuedAt: new Date().toISOString() });
    this.updatePendingCount();
  }

  private async updatePendingCount() {
    if (!this.db) return;
    const count = await this.db.count('orders');
    this.pendingSyncCount.set(count);
  }

  async getQueuedOrders() {
    if (!this.db) await this.initDB();
    return this.db.getAll('orders');
  }

  async removeQueuedOrder(id: number) {
    if (!this.db) await this.initDB();
    await this.db.delete('orders', id);
    this.updatePendingCount();
  }

  async syncQueuedOrders() {
    if (!this.isOnline()) return;
    
    const orders = await this.getQueuedOrders();
    if (orders.length === 0) return;

    console.log(\`Syncing \${orders.length} offline orders...\`);
    for (const order of orders) {
      try {
        // Logic to replay queued HTTP requests would go here
        
        await this.removeQueuedOrder(order.id);
      } catch (err) {
        console.error('Failed to sync order', order, err);
      }
    }
  }

  // --- CACHING METHODS ---
  async saveCache(store: 'products' | 'categories' | 'customers', items: any[]) {
    if (!this.db) await this.initDB();
    const tx = this.db.transaction(store, 'readwrite');
    for (const item of items) {
      if (item && item.id) {
        tx.store.put(item);
      }
    }
    await tx.done;
  }

  async getCache(store: 'products' | 'categories' | 'customers') {
    if (!this.db) await this.initDB();
    return this.db.getAll(store);
  }
}
`;

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/offline.service.ts', code);
