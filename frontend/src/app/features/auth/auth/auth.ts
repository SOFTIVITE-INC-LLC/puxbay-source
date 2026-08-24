import { Component, inject, signal } from '@angular/core';

import { FormsModule } from '@angular/forms';
import { AuthService } from '../../../core/services/auth.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-auth',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './auth.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .dark .glass-panel {
      background: rgba(0, 0, 0, 0.2);
    }
  `,
})
export class Auth {
  authService = inject(AuthService);
  router = inject(Router);

  email = signal('');
  password = signal('');
  errorMsg = signal('');

  onSubmit() {
    this.errorMsg.set('');
    this.authService.login({ email: this.email(), password: this.password() }).subscribe({
      next: () => {
        this.router.navigate(['/dashboard']);
      },
      error: () => {
        this.errorMsg.set('Invalid credentials. Please try again.');
        this.authService.loading.set(false);
      },
    });
  }
}
