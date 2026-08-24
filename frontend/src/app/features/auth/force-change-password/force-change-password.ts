import { Component, inject, signal, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router, ActivatedRoute, RouterModule } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-force-change-password',
  standalone: true,
  imports: [RouterModule, ReactiveFormsModule, CommonModule],
  templateUrl: './force-change-password.html',
  styles: ``,
})
export class ForceChangePassword implements OnInit {
  form: FormGroup;
  errorMessage = signal<string | null>(null);
  isLoading = signal(false);
  
  username = '';
  temporaryPassword = '';

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute
  ) {
    this.form = this.fb.group({
      newPassword: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', Validators.required]
    }, { validators: this.passwordMatchValidator });
  }

  ngOnInit() {
    // Get state from the router navigation
    const state = window.history.state;
    if (state && state.username && state.temporaryPassword) {
      this.username = state.username;
      this.temporaryPassword = state.temporaryPassword;
    } else {
      // If no state, they shouldn't be here
      this.router.navigate(['/login']);
    }
  }

  passwordMatchValidator(g: FormGroup) {
    return g.get('newPassword')?.value === g.get('confirmPassword')?.value
      ? null : { mismatch: true };
  }

  onSubmit() {
    if (this.form.valid) {
      this.isLoading.set(true);
      this.errorMessage.set(null);
      const newPassword = this.form.get('newPassword')?.value;

      this.authService.changeTemporaryPassword(this.username, this.temporaryPassword, newPassword).subscribe({
        next: () => {
          this.isLoading.set(false);
          const role = this.authService.currentUser()?.role;
          if (role === 'sales') {
            this.router.navigate(['/pos']);
          } else if (role === 'supplier') {
            this.router.navigate(['/b2b']);
          } else {
            this.router.navigate(['/dashboard']);
          }
        },
        error: (err: any) => {
          this.isLoading.set(false);
          this.errorMessage.set(err.error?.error || 'Failed to change password. Please try again.');
        }
      });
    } else {
      if (this.form.hasError('mismatch')) {
        this.errorMessage.set('Passwords do not match.');
      } else {
        this.errorMessage.set('Please provide a valid password (min 8 characters).');
      }
    }
  }
}
