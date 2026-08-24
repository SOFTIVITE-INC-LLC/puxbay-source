import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule, Router } from '@angular/router';
import { StorefrontAuthService } from '../../../../core/store/services/storefront-auth.service';
import { ToastService } from '../../../../core/store/services/toast.service';

@Component({
  selector: 'app-store-register',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './register.component.html'
})
export class RegisterComponent {
  authService = inject(StorefrontAuthService);
  toast = inject(ToastService);
  router = inject(Router);

  name = '';
  email = '';
  password = '';
  isLoading = false;

  onSubmit() {
    if (!this.name || !this.email || !this.password) return;
    this.isLoading = true;
    this.authService.register({ name: this.name, email: this.email, password: this.password }).subscribe({
      next: () => {
        this.isLoading = false;
        this.toast.show('Account created successfully!', 'success');
        this.router.navigate(['/store/account']);
      },
      error: (err) => {
        this.isLoading = false;
        this.toast.show(err.error?.error || 'Registration failed', 'error');
      }
    });
  }
}
