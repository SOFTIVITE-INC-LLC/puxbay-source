import { HttpInterceptorFn } from '@angular/common/http';
import { environment } from '../environments/environment';

export const apiInterceptor: HttpInterceptorFn = (req, next) => {
  let modifiedReq = req;

  // Set withCredentials: true so that HttpOnly session cookies are sent across origins
  modifiedReq = modifiedReq.clone({
    withCredentials: true
  });

  // If the request is for our API and we have an apiUrl configured (e.g. in prod)
  if (req.url.startsWith('/api/v1') && environment.apiUrl) {
    modifiedReq = modifiedReq.clone({ url: `${environment.apiUrl}${req.url}` });
  }
  
  return next(modifiedReq);
};
