import { inject } from '@angular/core';
import { Router, type CanActivateFn } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { toObservable } from '@angular/core/rxjs-interop';
import { filter, map } from 'rxjs/operators';

export const authGuard: CanActivateFn = (route, state) => {
  if (typeof window === 'undefined') return true;

  const router = inject(Router);
  const authService = inject(AuthService);

  // If already initialized, return result synchronously
  if (authService.isInitialized()) {
    if (authService.isAuthenticated()) return true;
    return router.parseUrl('/login');
  }

  return toObservable(authService.isInitialized).pipe(
    filter(isInit => isInit),
    map(() => {
      if (authService.isAuthenticated()) return true;
      return router.parseUrl('/login');
    })
  );
};
