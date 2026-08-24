import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { BranchService } from '../services/branch.service';

export const branchInterceptor: HttpInterceptorFn = (req, next) => {
  const branchService = inject(BranchService);
  const activeBranch = branchService.activeBranch();

  if (activeBranch && activeBranch.id) {
    const cloned = req.clone({
      setHeaders: {
        'X-Branch-ID': activeBranch.id
      }
    });
    return next(cloned);
  }
  
  return next(req);
};
