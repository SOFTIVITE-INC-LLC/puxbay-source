import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AlertService } from './services/alert.service';

/**
 * Global HTTP Error Interceptor.
 * Handles 401 (session expired), 403 (insufficient permissions), and 5xx errors.
 */
export const httpErrorInterceptor: HttpInterceptorFn = (req, next) => {
  const router = inject(Router);
  const alert = inject(AlertService);

  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 401 && !req.url.includes('/auth/login') && !req.url.includes('/auth/session') && !req.url.includes('/auth/refresh')) {
        // Session expired or invalid — redirect to login
        router.navigate(['/login'], { queryParams: { reason: 'session_expired' } });
      } else if (error.status === 403 && !req.url.includes('/auth/login')) {
        alert.error('Insufficient permissions for this action', 'Access Denied');
      } else if (error.status >= 500) {
        alert.error('Server error — please try again in a moment', 'System Error');
      }
      return throwError(() => error);
    })
  );
};
