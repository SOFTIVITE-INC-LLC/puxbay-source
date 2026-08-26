import { ApplicationConfig, provideBrowserGlobalErrorListeners, APP_INITIALIZER } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideCharts, withDefaultRegisterables } from 'ng2-charts';

import { routes } from './app.routes';
import { authInterceptor } from './auth.interceptor';
import { csrfInterceptor } from './csrf.interceptor';
import { apiInterceptor } from './api.interceptor';
import { httpErrorInterceptor } from './http-error.interceptor';
import { AuthService } from './services/auth.service';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideRouter(routes),
    provideHttpClient(withInterceptors([apiInterceptor, csrfInterceptor, authInterceptor, httpErrorInterceptor])),
    provideCharts(withDefaultRegisterables()),
    {
      provide: APP_INITIALIZER,
      useFactory: (authService: AuthService) => () => authService.restoreSession(),
      deps: [AuthService],
      multi: true
    }
  ]
};
