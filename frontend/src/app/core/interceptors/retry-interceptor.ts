import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { retry, timer, throwError } from 'rxjs';

export const retryInterceptor: HttpInterceptorFn = (req, next) => {
  // We only want to retry idempotent requests (GET, HEAD, OPTIONS)
  if (req.method !== 'GET' && req.method !== 'HEAD' && req.method !== 'OPTIONS') {
    return next(req);
  }

  return next(req).pipe(
    retry({
      count: 3,
      delay: (error: HttpErrorResponse, retryCount: number) => {
        // Don't retry on client errors (4xx) except 408 (Request Timeout) or 429 (Too Many Requests)
        if (error.status >= 400 && error.status < 500 && error.status !== 408 && error.status !== 429) {
          return throwError(() => error);
        }

        // Exponential backoff: 1s, 2s, 4s
        const delayMs = Math.pow(2, retryCount - 1) * 1000;
        return timer(delayMs);
      }
    })
  );
};
