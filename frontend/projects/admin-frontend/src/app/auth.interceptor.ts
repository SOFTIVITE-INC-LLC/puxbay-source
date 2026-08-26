import { HttpInterceptorFn, HttpErrorResponse, HttpClient } from '@angular/common/http';
import { inject } from '@angular/core';
import { throwError, catchError, switchMap, BehaviorSubject, filter, take } from 'rxjs';

let isRefreshing = false;
const refreshTokenSubject = new BehaviorSubject<boolean | null>(null);

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const http = inject(HttpClient);

  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      const isAuthEndpoint = req.url.includes('/auth/login') ||
                             req.url.includes('/auth/session') ||
                             req.url.includes('/auth/refresh') ||
                             req.url.includes('/auth/logout');

      if (error.status === 401 && !isAuthEndpoint) {
        if (!isRefreshing) {
          isRefreshing = true;
          refreshTokenSubject.next(null);

          return http.post('/api/v1/auth/refresh', {}, { withCredentials: true }).pipe(
            switchMap(() => {
              isRefreshing = false;
              refreshTokenSubject.next(true);
              return next(req);
            }),
            catchError((err) => {
              isRefreshing = false;
              refreshTokenSubject.next(false);
              return throwError(() => error);
            })
          );
        } else {
          return refreshTokenSubject.pipe(
            filter(result => result !== null),
            take(1),
            switchMap((success) => {
              if (success) {
                return next(req);
              }
              return throwError(() => error);
            })
          );
        }
      }
      return throwError(() => error);
    })
  );
};

