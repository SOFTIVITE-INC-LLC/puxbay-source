import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { throwError, timer, EMPTY } from 'rxjs';
import { catchError, retry } from 'rxjs/operators';
import { ToastrService } from 'ngx-toastr';

export const globalErrorInterceptor: HttpInterceptorFn = (req, next) => {
  const toastr = inject(ToastrService);
  const router = inject(Router);

  return next(req).pipe(
    retry({
      count: 2,
      delay: (error, retryCount) => {
        if (error.status === 503 || error.status === 504 || error.status === 0) {
          // Retry on network errors or timeouts
          return timer(1000 * Math.pow(2, retryCount));
        }
        throw error;
      }
    }),
    catchError((error: HttpErrorResponse) => {
      let errorMessage = 'An unknown error occurred!';

      if (typeof ErrorEvent !== 'undefined' && error.error instanceof ErrorEvent) {
        // Client-side or network error
        errorMessage = `Network error: ${error.error.message}`;
      } else {
        // Backend error
        if (error.status === 401) {
          if (typeof window === 'undefined') {
            return EMPTY;
          }
          // Do not show session expired toast for public storefront/kiosk/auth pages
          const isPublicReq = req.url.includes('/storefront') ||
                              req.url.includes('/kiosk') ||
                              req.url.includes('/public') ||
                              req.url.includes('/auth/login') ||
                              req.url.includes('/auth/register');
          if (isPublicReq) {
            return throwError(() => error);
          }
          errorMessage = 'Session expired. Please log in again.';
        } else if (error.status === 402) {
          errorMessage = error.error?.error || 'Payment Required. Please update your billing details.';
          router.navigate(['/billing']);
        } else if (error.status === 403) {
          errorMessage = 'You do not have permission to perform this action.';
        } else if (error.status === 404) {
          const isPublicReq = req.url.includes('/storefront') || req.url.includes('/kiosk') || req.url.includes('/public');
          if (isPublicReq) {
            return throwError(() => error);
          }
          errorMessage = 'Requested resource was not found.';
        } else if (error.status === 422) {
          errorMessage = 'Validation failed. Please check your inputs.';
        } else if (error.status >= 500) {
          errorMessage = 'Server error. Our engineers have been notified.';
        } else if (error.error && error.error.message) {
          errorMessage = error.error.message;
        }
      }

      // Display global toast
      if (typeof window !== 'undefined') {
        const isPublicReq = req.url.includes('/storefront') || req.url.includes('/kiosk') || req.url.includes('/public');
        if (!isPublicReq) {
          toastr.error(errorMessage, '');
        }
      }

      // Continue throwing to allow component-level handling if needed
      return throwError(() => error);
    })
  );
};
