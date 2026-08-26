import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';

/**
 * Global HTTP Error Interceptor.
 * Handles 401 (session expired), 403 (insufficient permissions), and 5xx errors.
 */
export const httpErrorInterceptor: HttpInterceptorFn = (req, next) => {
  const router = inject(Router);

  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 401 && !req.url.includes('/auth/login') && !req.url.includes('/auth/session') && !req.url.includes('/auth/refresh')) {
        // Session expired or invalid — redirect to login
        router.navigate(['/login'], { queryParams: { reason: 'session_expired' } });
      } else if (error.status === 403 && !req.url.includes('/auth/login')) {
        // Show an in-page notification without disrupting the user
        showToast('⛔ Insufficient permissions for this action', 'error');
      } else if (error.status >= 500) {
        showToast('🔴 Server error — please try again in a moment', 'error');
      }
      return throwError(() => error);
    })
  );
};

function showToast(message: string, type: 'error' | 'success' | 'info') {
  const existing = document.getElementById('__pux_toast');
  if (existing) existing.remove();

  const toast = document.createElement('div');
  toast.id = '__pux_toast';
  toast.style.cssText = `
    position: fixed;
    bottom: 24px;
    right: 24px;
    z-index: 99999;
    background: ${type === 'error' ? '#ef4444' : type === 'success' ? '#22c55e' : '#6366f1'};
    color: white;
    font-family: system-ui, sans-serif;
    font-size: 14px;
    font-weight: 600;
    padding: 12px 20px;
    border-radius: 12px;
    box-shadow: 0 8px 30px rgba(0,0,0,0.2);
    transform: translateY(0);
    transition: opacity 0.3s ease;
    max-width: 360px;
    line-height: 1.4;
  `;
  toast.textContent = message;
  document.body.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = '0';
    setTimeout(() => toast.remove(), 300);
  }, 4000);
}
