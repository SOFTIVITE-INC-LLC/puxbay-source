import { inject } from '@angular/core';
import { Router, type CanActivateFn } from '@angular/router';
import { AuthService } from '../services/auth.service';

export const roleGuard: CanActivateFn = (route, state) => {
  if (typeof window === 'undefined') return true;
  
  const auth = inject(AuthService);
  const router = inject(Router);
  
  const currentUser = auth.currentUser();
  const requiredRoles = route.data?.['roles'] as string[];

  if (!currentUser) {
    return router.parseUrl('/login');
  }

  if (requiredRoles && requiredRoles.length > 0) {
    if (requiredRoles.includes(currentUser.role)) {
      return true;
    }
    // Fallback to dashboard if unauthorized
    return router.parseUrl('/dashboard');
  }

  return true;
};
