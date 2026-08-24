import { HttpInterceptorFn } from '@angular/common/http';

/**
 * Auth interceptor — this is now a passthrough.
 *
 * Authentication is handled via HttpOnly `pux_session` cookies sent automatically
 * by the browser. No Authorization header is needed.
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  return next(req);
};
