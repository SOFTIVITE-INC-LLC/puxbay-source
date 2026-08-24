import { inject } from '@angular/core';
import { Router, type CanActivateFn } from '@angular/router';
import { AuthService } from '../services/auth.service';

export const permissionGuard: CanActivateFn = (route, state) => {
  if (typeof window === 'undefined') return true;
  
  const auth = inject(AuthService);
  const router = inject(Router);
  
  const currentUser = auth.currentUser();
  const requiredPermissions = route.data?.['permissions'] as string[];
  const requiredRoles = route.data?.['roles'] as string[]; // Legacy fallback

  if (!currentUser) {
    return router.parseUrl('/login');
  }

  // Convert role to lowercase for case-insensitive matching
  const userRole = (currentUser.role || '').toLowerCase();

  // Superadmin / Admin bypass
  if (userRole === 'superadmin' || userRole === 'admin') {
    return true;
  }

  if (requiredPermissions && requiredPermissions.length > 0) {
    for (const perm of requiredPermissions) {
      if (auth.hasPermission(perm)) {
        return true;
      }
    }
  } else if (requiredRoles && requiredRoles.length > 0) {
    // Legacy fallback checking roles directly
    if (requiredRoles.includes(userRole)) {
      return true;
    }
  } else {
    // No specific permission required for this route
    return true;
  }

  // Fallback to dashboard if unauthorized
  return router.parseUrl('/dashboard');
};
