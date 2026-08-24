import { Injectable, signal } from '@angular/core';
import Dexie, { Table } from 'dexie';

/**
 * Cached product entry for offline POS usage.
 */
export interface CachedProduct {
  id: string;
  name: string;
  sku: string;
  selling_price: number;
  cost_price: number;
  current_stock: number;
  stock_unit?: string;
  image_url?: string;
  category?: { id: string; name: string };
  track_inventory: boolean;
  is_active: boolean;
  branch_id?: string;
  cached_at: number; // Unix timestamp
}

/**
 * An order that was queued while offline and needs to be synced.
 */
export interface QueuedOrder {
  id?: number;           // Auto-incremented local ID
  payload: any;          // The full POST body for /api/v1/pos/checkout
  queuedAt: number;
  retries: number;
  lastError?: string;
}

/**
 * IndexedDB database via Dexie for offline POS support.
 */
class PuxbayOfflineDB extends Dexie {
  products!: Table<CachedProduct, string>;
  queuedOrders!: Table<QueuedOrder, number>;

  constructor() {
    super('puxbay_offline');
    this.version(1).stores({
      products: 'id, name, sku, branch_id, is_active',
      queuedOrders: '++id, queuedAt, retries',
    });
  }
}

@Injectable({ providedIn: 'root' })
export class OfflineDbService {
  private db: PuxbayOfflineDB | null = null;

  /** Reactive signal: number of orders waiting to sync */
  readonly pendingCount = signal(0);

  constructor() {
    if (typeof window !== 'undefined') {
      this.db = new PuxbayOfflineDB();
      this._refreshPendingCount();
      // Listen for online events to trigger sync
      window.addEventListener('online', () => this.syncPendingOrders());
    }
  }

  // ─── Product Cache ────────────────────────────────────────────

  /** Save products to local cache (replaces existing) */
  async cacheProducts(products: CachedProduct[]): Promise<void> {
    if (!this.db) return;
    const withTimestamp = products.map(p => ({ ...p, cached_at: Date.now() }));
    await this.db.products.bulkPut(withTimestamp);
  }

  /** Get all cached products (optionally filter by branch) */
  async getCachedProducts(branchId?: string): Promise<CachedProduct[]> {
    if (!this.db) return [];
    let query = this.db.products.where('is_active').equals(1 as any);
    if (branchId) {
      return this.db.products.where('branch_id').equals(branchId).toArray();
    }
    return query.toArray();
  }

  /** Search cached products by name or SKU */
  async searchCachedProducts(query: string, branchId?: string): Promise<CachedProduct[]> {
    const lower = query.toLowerCase();
    const products = await this.getCachedProducts(branchId);
    return products.filter(p =>
      p.name.toLowerCase().includes(lower) || p.sku.toLowerCase().includes(lower)
    );
  }

  /** Clear the product cache */
  async clearProductCache(): Promise<void> {
    if (!this.db) return;
    await this.db.products.clear();
  }

  /** Returns how old (in minutes) the cache is */
  async getCacheAge(): Promise<number | null> {
    if (!this.db) return null;
    const first = await this.db.products.orderBy('cached_at').first();
    if (!first) return null;
    return Math.floor((Date.now() - first.cached_at) / 60000);
  }

  // ─── Order Queue ──────────────────────────────────────────────

  /** Queue an order to be synced when the connection is restored */
  async queueOrder(payload: any): Promise<number> {
    if (!this.db) return -1;
    const id = await this.db.queuedOrders.add({
      payload,
      queuedAt: Date.now(),
      retries: 0,
    });
    await this._refreshPendingCount();
    return id as number;
  }

  /** Get all queued orders */
  async getPendingOrders(): Promise<QueuedOrder[]> {
    if (!this.db) return [];
    return this.db.queuedOrders.orderBy('queuedAt').toArray();
  }

  /** Remove a successfully synced order from the queue */
  async removeQueuedOrder(id: number): Promise<void> {
    if (!this.db) return;
    await this.db.queuedOrders.delete(id);
    await this._refreshPendingCount();
  }

  /**
   * Attempt to sync all pending orders to the server.
   * Calls the provided syncFn for each order.
   * Returns the number of successfully synced orders.
   */
  async syncPendingOrders(
    syncFn?: (payload: any) => Promise<any>
  ): Promise<{ synced: number; failed: number }> {
    if (typeof navigator === 'undefined' || !navigator.onLine || !syncFn || !this.db) {
      return { synced: 0, failed: 0 };
    }

    const orders = await this.getPendingOrders();
    let synced = 0;
    let failed = 0;

    for (const order of orders) {
      try {
        await syncFn(order.payload);
        await this.removeQueuedOrder(order.id!);
        synced++;
      } catch (err: any) {
        // Increment retry count, save error message
        if (this.db) {
          await this.db.queuedOrders.update(order.id!, {
            retries: order.retries + 1,
            lastError: err?.message ?? 'Sync failed',
          });
        }
        failed++;
      }
    }

    await this._refreshPendingCount();
    return { synced, failed };
  }

  private async _refreshPendingCount(): Promise<void> {
    if (!this.db) return;
    const count = await this.db.queuedOrders.count();
    this.pendingCount.set(count);
  }
}
