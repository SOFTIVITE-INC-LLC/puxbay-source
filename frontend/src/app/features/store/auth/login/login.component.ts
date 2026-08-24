import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule, Router } from '@angular/router';
import { StorefrontAuthService } from '../../../../core/store/services/storefront-auth.service';
import { ToastService } from '../../../../core/store/services/toast.service';

@Component({
  selector: 'app-store-login',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './login.component.html'
})
export class LoginComponent {
  authService = inject(StorefrontAuthService);
  toast = inject(ToastService);
  router = inject(Router);

  email = '';
  password = '';
  isLoading = false;

  onSubmit() {
    if (!this.email || !this.password) return;
    this.isLoading = true;
    this.authService.login({ email: this.email, password: this.password }).subscribe({
      next: () => {
        this.isLoading = false;
        this.toast.show('Successfully logged in!', 'success');
        this.router.navigate(['/store/account']);
      },
      error: (err) => {
        this.isLoading = false;
        this.toast.show(err.error?.error || 'Login failed', 'error');
      }
    });
  }
}
