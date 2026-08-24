import { Component, ViewEncapsulation, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { passwordStrengthValidator } from '../../../core/validators/password-strength.validator';
import { computed } from '@angular/core';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterModule],
  templateUrl: './register.html',
  encapsulation: ViewEncapsulation.None
})
export class Register {
  private fb = inject(FormBuilder);
  private auth = inject(AuthService);
  private router = inject(Router);

  step = signal(1);
  errorMessage = signal('');
  isLoading = signal(false);

  registerForm = this.fb.group({
    first_name: ['', Validators.required],
    last_name: ['', Validators.required],
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, passwordStrengthValidator()]],
    company_name: ['', Validators.required],
    subdomain: ['', Validators.required],
    address: ['']
  });

  passwordStrength = computed(() => {
    const password = this.registerForm.get('password')?.value || '';
    if (!password) return 0;
    let strength = 0;
    if (password.length >= 8) strength++;
    if (/[A-Z]+/.test(password)) strength++;
    if (/[0-9]+/.test(password)) strength++;
    if (/[\W_]+/.test(password)) strength++;
    return strength;
  });

  nextStep() {
    this.errorMessage.set('');
    // Optional: add check for step 1 fields if needed
    this.step.set(2);
  }

  onSubmit() {
    this.errorMessage.set('');
    if (this.registerForm.valid) {
      this.isLoading.set(true);
      const formVal = this.registerForm.value;
      const payload = {
        username: formVal.email, // using email as username
        email: formVal.email,
        password: formVal.password,
        first_name: formVal.first_name,
        last_name: formVal.last_name,
        company_name: formVal.company_name,
        subdomain: formVal.subdomain
      };

      this.auth.registerAndLogin(payload).subscribe({
        next: () => {
          this.isLoading.set(false);
          this.router.navigate(['/dashboard']);
        },
        error: (err) => {
          this.isLoading.set(false);
          console.error('Registration failed', err);
          this.errorMessage.set(err.error?.error || 'Registration failed. Please try again.');
        }
      });
    } else {
      this.errorMessage.set('Please fill out all required fields correctly.');
    }
  }
}
