const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/customer.service.ts', 'utf8');

code = code.replace("import { Observable, tap } from 'rxjs';", "import { Observable, tap, from } from 'rxjs';\nimport { OfflineService } from './offline.service';");

code = code.replace("private api = inject(ApiService);", "private api = inject(ApiService);\n  private offlineService = inject(OfflineService);");

code = code.replace(/getCustomers\(params\?: any\): Observable<\{data: Customer\[\], total: number, page: number, limit: number\}> \{[\s\S]*?\}\);\n  \}/, `getCustomers(params?: any): Observable<{data: Customer[], total: number, page: number, limit: number}> {
    this.loading.set(true);
    
    if (!navigator.onLine) {
      return from(this.offlineService.getCache('customers').then(custs => {
        this.customers.set(custs);
        this.total.set(custs.length);
        this.loading.set(false);
        return { data: custs, total: custs.length, page: 1, limit: custs.length };
      }));
    }

    return this.api.get<{data: Customer[], total: number, page: number, limit: number}>('/customers', { params }).pipe(
      tap(res => {
        const data = res.data || [];
        this.customers.set(data);
        this.total.set(res.total || 0);
        this.loading.set(false);
        this.offlineService.saveCache('customers', data);
      })
    );
  }`);

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/customer.service.ts', code);
