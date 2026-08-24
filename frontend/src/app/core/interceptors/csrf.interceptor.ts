import { HttpInterceptorFn } from '@angular/common/http';
import { tap } from 'rxjs/operators';

let csrfToken: string | null = null;

/**
 * CSRF Interceptor — Double Submit Cookie pattern.
 *
 * 1. On every response, capture the X-CSRF-Token header sent by the server.
 * 2. On every mutating request (POST/PUT/PATCH/DELETE), attach the captured
 *    token as an X-CSRF-Token request header so the server can validate it.
 *
 * If no token is stored yet (very first mutating request), we fall back to
 * reading the csrf_token cookie directly from document.cookie.
 */
export const csrfInterceptor: HttpInterceptorFn = (req, next) => {
  const mutatingMethods = ['POST', 'PUT', 'PATCH', 'DELETE'];

  // Read token from cookie as a fallback (works before first response arrives)
  function getTokenFromCookie(): string | null {
    if (typeof document === 'undefined') return null;
    const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]+)/);
    return match ? decodeURIComponent(match[1]) : null;
  }

  const token = csrfToken || getTokenFromCookie();

  // Attach token on mutating requests
  let outgoingReq = req;
  if (token && mutatingMethods.includes(req.method.toUpperCase())) {
    outgoingReq = req.clone({
      setHeaders: { 'X-CSRF-Token': token }
    });
  }

  return next(outgoingReq).pipe(
    tap({
      next: (event: any) => {
        if (event?.headers) {
          const newToken = event.headers.get('X-CSRF-Token');
          if (newToken) {
            csrfToken = newToken;
          }
        }
      },
      error: (err: any) => {
        if (err?.headers) {
          const newToken = err.headers.get('X-CSRF-Token');
          if (newToken) {
            csrfToken = newToken;
          }
        }
      }
    })
  );
};
