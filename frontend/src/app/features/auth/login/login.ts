import { Component, inject, signal } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { BranchService } from '../../../core/services/branch.service';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [RouterModule, ReactiveFormsModule, CommonModule],
  templateUrl: './login.html',
  styles: ``,
})
export class Login {
  loginForm: FormGroup;
  errorMessage = signal<string | null>(null);
  isLoading = signal(false);
  requires2FA = signal(false);

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private branchService: BranchService,
    private router: Router
  ) {
    this.loginForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', Validators.required],
      totpCode: ['']
    });
  }

  onSubmit() {
    if (this.loginForm.valid) {
      this.isLoading.set(true);
      this.errorMessage.set(null);
      const { email, password, totpCode } = this.loginForm.value;

      this.authService.login({ username: email, password, totpCode }).subscribe({
        next: () => {
          this.isLoading.set(false);
          const user = this.authService.currentUser();
          const role = user?.role;
          const branchId = user?.branch_id;

          const navigate = () => {
            let targetPath = '/dashboard';
            if (role === 'sales') {
              targetPath = '/pos';
            } else if (role === 'supplier') {
              targetPath = '/b2b';
            }

            const userSubdomain = user?.subdomain;
            if (userSubdomain && typeof window !== 'undefined') {
              const hostname = window.location.hostname;
              if (hostname !== 'localhost' && hostname !== '127.0.0.1') {
                const parts = hostname.split('.');
                const currentSubdomain = (parts.length >= 2 && parts[0] !== 'www' && parts[0] !== 'api') ? parts[0] : null;
                
                if (currentSubdomain !== userSubdomain) {
                  if (currentSubdomain) {
                    parts[0] = userSubdomain;
                  } else {
                    parts.unshift(userSubdomain);
                  }
                  const newHostname = parts.join('.');
                  const port = window.location.port ? `:${window.location.port}` : '';
                  window.location.href = `${window.location.protocol}//${newHostname}${port}${targetPath}`;
                  return;
                }
              }
            }
            this.router.navigate([targetPath]);
          };

          if (branchId) {
            this.branchService.getBranch(branchId).subscribe({
              next: (branch) => {
                this.branchService.setActiveBranch(branch);
                navigate();
              },
              error: () => navigate() // Fallback if branch fails to load
            });
          } else {
            navigate();
          }
        },
        error: (err) => {
          this.isLoading.set(false);
          if (err.error?.error === '2fa_required') {
            this.requires2FA.set(true);
            this.loginForm.get('totpCode')?.setValidators([Validators.required, Validators.minLength(6), Validators.maxLength(6)]);
            this.loginForm.get('totpCode')?.updateValueAndValidity();
          } else if (err.error?.error === 'password_change_required') {
            // Redirect to force change password UI, passing the credentials
            this.router.navigate(['/force-change-password'], {
              state: {
                username: email,
                temporaryPassword: password
              }
            });
          } else {
            this.errorMessage.set(err.error?.error || 'Login failed. Please check your credentials.');
          }
        }
      });
    } else {
      this.errorMessage.set('Please provide valid credentials.');
    }
  }
}
