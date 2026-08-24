import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { TenantStore } from '../services/tenant.store';

export const tenantInterceptor: HttpInterceptorFn = (req, next) => {
  const tenantStore = inject(TenantStore);
  
  // Try from store first (which handles dynamic subdomains or dev fallbacks if configured)
  let subdomain = tenantStore.subdomain();

  if (!subdomain) {
    // Extract subdomain from the current host
    const hostname = typeof window !== 'undefined' ? window.location.hostname : '';
    const parts = hostname.split('.');
    
    // If it's something like tenant.puxbay.com or tenant.localhost
    if (parts.length >= 2 && parts[0] !== 'www' && parts[0] !== 'api') {
      subdomain = parts[0];
    } else if (hostname === 'localhost' || hostname === '127.0.0.1') {
      subdomain = typeof localStorage !== 'undefined' ? (localStorage.getItem('dev_tenant') || 'thinkce') : 'thinkce'; 
    }
    
    if (subdomain) {
      tenantStore.setSubdomain(subdomain);
    }
  }

  // Clone the request and add the custom header if a subdomain is found
  if (subdomain) {
    const cloned = req.clone({
      setHeaders: {
        'X-Tenant-Subdomain': subdomain
      }
    });
    return next(cloned);
  }
  
  return next(req);
};
