const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/offline.service.ts', 'utf8');

// Upgrade version to 2 and create new object stores
code = code.replace(/await openDB\('puxbay-offline-db', 1, \{[\s\S]*?\}\);/, `await openDB('puxbay-offline-db', 2, {
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
    });`);

// Add generic sync methods
code += `
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
`;

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/offline.service.ts', code);
