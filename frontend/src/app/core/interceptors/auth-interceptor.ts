import { HttpInterceptorFn } from '@angular/common/http';

/**
 * Auth interceptor — this is now a passthrough.
 *
 * Authentication is handled via HttpOnly `pux_session` cookies that the browser
 * sends automatically with every cross-origin request (because withCredentials: true
 * is set in ApiService). There is no need to manually attach any Authorization header.
 *
 * CSRF protection is handled by csrfInterceptor.
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  // Supplier Portal uses JWT Bearer token authentication
  if (req.url.includes('/supplier-portal') && !req.url.includes('/supplier-portal/login')) {
    if (typeof window !== 'undefined') {
      const supplierToken = localStorage.getItem('supplier_token');
      if (supplierToken) {
        req = req.clone({
          setHeaders: {
            Authorization: `Bearer ${supplierToken}`
          }
        });
      }
    }
  }
  return next(req);
};
