import {
  ErrorHandler,
  APP_INITIALIZER,
  ApplicationConfig,
  provideBrowserGlobalErrorListeners,
  isDevMode,
  importProvidersFrom,
  DEFAULT_CURRENCY_CODE,
} from '@angular/core';
import { GlobalErrorHandler } from './core/errors/global-error-handler';
import { provideToastr } from 'ngx-toastr';
import { provideAnimations } from '@angular/platform-browser/animations';
import { TenantStore } from './core/services/tenant.store';
import { provideRouter, withPreloading, PreloadAllModules } from '@angular/router';
import { provideHttpClient, withInterceptors, HttpClient, withFetch } from '@angular/common/http';
import { DatePipe } from '@angular/common';
import { routes } from './app.routes';
import { authInterceptor } from './core/interceptors/auth-interceptor';
import { tenantInterceptor } from './core/interceptors/tenant-interceptor';
import { branchInterceptor } from './core/interceptors/branch-interceptor';
import { retryInterceptor } from './core/interceptors/retry-interceptor';
import { globalErrorInterceptor } from './core/interceptors/global-error.interceptor';
import { csrfInterceptor } from './core/interceptors/csrf.interceptor';
import { provideServiceWorker } from '@angular/service-worker';
import { provideClientHydration } from '@angular/platform-browser';

export function initializeTenant(tenantStore: TenantStore) {
  return () => {
    // Extract subdomain from the current host, handling SSR
    const hostname = typeof window !== 'undefined' ? window.location.hostname : '';
    const parts = hostname.split('.');

    let subdomain = '';

    if (parts.length >= 2 && parts[0] !== 'www' && parts[0] !== 'api') {
      subdomain = parts[0];
    } else if (hostname === 'localhost' || hostname === '127.0.0.1') {
      subdomain = typeof localStorage !== 'undefined' ? (localStorage.getItem('dev_tenant') || '') : '';
    }

    if (subdomain) {
      tenantStore.setSubdomain(subdomain);
    }
    return Promise.resolve();
  };
}

// export function HttpLoaderFactory(http: HttpClient) {
//   return new TranslateHttpLoader(http, './assets/i18n/', '.json');
// }

export const appConfig: ApplicationConfig = {
  providers: [
    DatePipe,
    { 
      provide: DEFAULT_CURRENCY_CODE, 
      useFactory: () => typeof localStorage !== 'undefined' ? (localStorage.getItem('currency_code') || 'GHS') : 'GHS'
    },
    provideAnimations(),
    { provide: ErrorHandler, useClass: GlobalErrorHandler },
    {
      provide: APP_INITIALIZER,
      useFactory: initializeTenant,
      deps: [TenantStore],
      multi: true,
    },
    provideBrowserGlobalErrorListeners(),
    provideRouter(routes, withPreloading(PreloadAllModules)),
    provideHttpClient(
      withFetch(),
      withInterceptors([
        authInterceptor,
        tenantInterceptor,
        branchInterceptor,
        retryInterceptor,
        csrfInterceptor,
        globalErrorInterceptor,
      ]),
    ),
    provideToastr({
      positionClass: 'toast-bottom-right',
      preventDuplicates: true,
    }),
    provideServiceWorker('ngsw-worker.js', {
      enabled: !isDevMode(),
      registrationStrategy: 'registerWhenStable:30000',
    }),
    provideServiceWorker('ngsw-worker.js', {
      enabled: !isDevMode(),
      registrationStrategy: 'registerWhenStable:30000',
    }),
    provideClientHydration(),
  ],
};
