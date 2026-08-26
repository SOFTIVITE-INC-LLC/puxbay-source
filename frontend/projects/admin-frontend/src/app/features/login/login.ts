import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './login.html',
})
export class LoginComponent {
  private authService = inject(AuthService);
  private router = inject(Router);

  username = signal('');
  password = signal('');
  error = signal('');
  isLoading = signal(false);

  onSubmit() {
    this.error.set('');
    this.isLoading.set(true);

    this.authService.login({
      username: this.username(),
      password: this.password()
    }).subscribe({
      next: () => {
        this.isLoading.set(false);
        const targetRoute = this.authService.getDefaultRoute();
        this.router.navigateByUrl(targetRoute).then((navigated) => {
          if (!navigated) {
            window.location.href = targetRoute;
          }
        });
      },
      error: (err) => {
        this.isLoading.set(false);
        this.error.set(err.error?.error || 'Invalid credentials');
      }
    });
  }
}
