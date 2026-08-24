import { HttpInterceptorFn } from '@angular/common/http';
import { environment } from '../environments/environment';

export const apiInterceptor: HttpInterceptorFn = (req, next) => {
  // If the request is for our API and we have an apiUrl configured (e.g. in prod)
  if (req.url.startsWith('/api/v1') && environment.apiUrl) {
    const apiReq = req.clone({ url: `${environment.apiUrl}${req.url}` });
    return next(apiReq);
  }
  return next(req);
};
