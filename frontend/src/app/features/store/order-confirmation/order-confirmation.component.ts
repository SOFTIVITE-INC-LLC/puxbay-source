import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { StorefrontAuthService } from '../../../core/store/services/storefront-auth.service';
import { HttpClient } from '@angular/common/http';
import { ToastService } from '../../../core/store/services/toast.service';
import { SessionService } from '../../../core/store/services/session.service';

@Component({
  selector: 'app-order-confirmation',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './order-confirmation.component.html'
})
export class OrderConfirmationComponent {
  authService = inject(StorefrontAuthService);
  http = inject(HttpClient);
  toast = inject(ToastService);
  sessionService = inject(SessionService);
  router = inject(Router);

  isGuest = signal(!this.authService.currentUser());
  name = signal('');
  password = signal('');
  isConverting = signal(false);

  convertAccount() {
    if (!this.password() || !this.name() || this.isConverting()) return;
    this.isConverting.set(true);

    const sessionId = this.sessionService.getSessionId();
    
    this.http.post('/api/v1/storefront/checkout/convert-guest', {
      name: this.name(),
      password: this.password()
    }, {
      headers: { 'X-Session-ID': sessionId || '' }
    }).subscribe({
      next: (res: any) => {
        this.authService.setToken(res.token);
        this.authService.currentUser.set(res.customer);
        this.isGuest.set(false);
        this.toast.show('Account created successfully!', 'success');
        this.router.navigate(['/store/account']);
      },
      error: (err) => {
        this.toast.show(err.error?.error || 'Failed to create account.', 'error');
        this.isConverting.set(false);
      }
    });
  }
}
