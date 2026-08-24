import { ErrorHandler, Injectable, Injector, NgZone } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { ToastrService } from 'ngx-toastr';

@Injectable()
export class GlobalErrorHandler implements ErrorHandler {
  constructor(private injector: Injector, private zone: NgZone) {}

  handleError(error: Error | HttpErrorResponse): void {
    const isBrowser = typeof window !== 'undefined';
    
    let message = 'An unexpected error occurred.';
    
    if (error instanceof HttpErrorResponse) {
      if (isBrowser && !navigator.onLine) {
        message = 'No Internet Connection';
      } else {
        message = error.error?.message || `Server error: ${error.status}`;
      }
    } else {
      // Client Error
      message = error.message ? error.message : error.toString();
      console.error('UI Crash:', error);
    }

    if (isBrowser) {
      const toastr = this.injector.get(ToastrService);
      this.zone.run(() => {
        toastr.error(message, 'Error', {
          timeOut: 5000,
        });
      });
    }
  }
}
