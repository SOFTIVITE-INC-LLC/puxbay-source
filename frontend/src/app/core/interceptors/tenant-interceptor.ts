import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { TenantStore } from '../services/tenant.store';

export const tenantInterceptor: HttpInterceptorFn = (req, next) => {
  const tenantStore = inject(TenantStore);
  
  let subdomain = tenantStore.subdomain();

  if (typeof window !== 'undefined') {
    // 1. Check query parameter ?tenant=xxx or ?subdomain=xxx in current URL
    const urlParams = new URLSearchParams(window.location.search);
    const querySubdomain = urlParams.get('tenant') || urlParams.get('subdomain');
    if (querySubdomain) {
      subdomain = querySubdomain;
      tenantStore.setSubdomain(subdomain);
      try {
        localStorage.setItem('dev_tenant', subdomain);
        localStorage.setItem('tenant_subdomain', subdomain);
      } catch (_) {}
    }

    // 2. If not found, check localStorage
    if (!subdomain) {
      subdomain = localStorage.getItem('tenant_subdomain') || localStorage.getItem('dev_tenant');
    }

    // 3. If still not found, check hostname
    if (!subdomain) {
      const hostname = window.location.hostname;
      const isIP = /^(\d{1,3}\.){3}\d{1,3}$/.test(hostname);
      
      if (hostname === 'localhost' || isIP) {
        subdomain = 'thinkce'; // default fallback for dev
      } else {
        const parts = hostname.split('.');
        if (parts.length >= 2 && parts[0] !== 'www' && parts[0] !== 'api') {
          subdomain = parts[0];
        }
      }
    }

    if (subdomain && tenantStore.subdomain() !== subdomain) {
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
