const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/catalog.service.ts', 'utf8');

code = code.replace("import { Observable, tap } from 'rxjs';", "import { Observable, tap, from, of } from 'rxjs';\nimport { OfflineService } from './offline.service';");

code = code.replace("private api = inject(ApiService);", "private api = inject(ApiService);\n  private offlineService = inject(OfflineService);");

code = code.replace(/getProducts\(params\?: any\): Observable<ProductListResponse> \{[\s\S]*?\}\);\n  \}/, `getProducts(params?: any): Observable<ProductListResponse> {
    this.loading.set(true);
    
    if (!navigator.onLine) {
      return from(this.offlineService.getCache('products').then(prods => {
        this.products.set(prods);
        this.totalProducts.set(prods.length);
        this.loading.set(false);
        return { data: prods, total: prods.length, page: 1, limit: prods.length };
      }));
    }

    return this.api.get<ProductListResponse>('/products', { params }).pipe(
      tap(res => {
        const data = res.data || [];
        this.products.set(data);
        this.totalProducts.set(res.total || 0);
        this.loading.set(false);
        this.offlineService.saveCache('products', data);
      })
    );
  }`);

// Update category logic
code = code.replace(/getCategories\(\): Observable<Category\[\]> \{[\s\S]*?\}\);\n  \}/, `getCategories(): Observable<Category[]> {
    if (!navigator.onLine) {
      return from(this.offlineService.getCache('categories').then(cats => {
        this.categories.set(cats);
        return cats;
      }));
    }

    return this.api.get<Category[]>('/categories').pipe(
      tap(res => {
        const data = res || [];
        this.categories.set(data);
        this.offlineService.saveCache('categories', data);
      })
    );
  }`);

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/catalog.service.ts', code);
