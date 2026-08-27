import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { ToastService } from '../../../core/services/toast';

@Component({
  selector: 'app-verify-email',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './verify-email.html'
})
export class VerifyEmailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private auth = inject(AuthService);
  private toast = inject(ToastService);

  email = signal<string>('');
  token = signal<string>('');
  otpDigits = signal<string[]>(['', '', '', '', '', '']);
  isLoading = signal<boolean>(false);
  isSuccess = signal<boolean>(false);
  errorMessage = signal<string>('');
  resendCountdown = signal<number>(0);
  private countdownTimer: any;

  ngOnInit() {
    this.route.queryParams.subscribe((params) => {
      if (params['email']) {
        this.email.set(params['email']);
      }
      if (params['token']) {
        this.token.set(params['token']);
        this.verifyViaToken(params['token']);
      }
    });
  }

  onDigitInput(event: Event, index: number) {
    const input = event.target as HTMLInputElement;
    const val = input.value.replace(/\D/g, ''); // numbers only
    const digits = [...this.otpDigits()];

    if (val.length > 1) {
      // Handle paste of complete 6-digit code
      const pasted = val.slice(0, 6).split('');
      pasted.forEach((d, i) => {
        if (i < 6) digits[i] = d;
      });
      this.otpDigits.set(digits);
      const nextInput = document.getElementById(`otp-${Math.min(pasted.length, 5)}`);
      nextInput?.focus();
      if (pasted.length === 6) {
        this.submitOTP();
      }
      return;
    }

    digits[index] = val ? val[0] : '';
    this.otpDigits.set(digits);

    if (val && index < 5) {
      const nextInput = document.getElementById(`otp-${index + 1}`);
      nextInput?.focus();
    }

    if (digits.every((d) => d.length === 1)) {
      this.submitOTP();
    }
  }

  onKeyDown(event: KeyboardEvent, index: number) {
    if (event.key === 'Backspace' && !this.otpDigits()[index] && index > 0) {
      const prevInput = document.getElementById(`otp-${index - 1}`);
      prevInput?.focus();
    }
  }

  verifyViaToken(token: string) {
    this.isLoading.set(true);
    this.errorMessage.set('');
    this.auth.verifyEmail(token).subscribe({
      next: () => {
        this.isLoading.set(false);
        this.isSuccess.set(true);
        this.toast.showSuccess('Email verified successfully! Redirecting to login...');
        setTimeout(() => this.router.navigate(['/auth/login']), 2500);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.error || 'Invalid or expired verification link.');
      }
    });
  }

  submitOTP() {
    const code = this.otpDigits().join('');
    if (code.length !== 6) {
      this.errorMessage.set('Please enter all 6 digits of the verification code.');
      return;
    }
    if (!this.email()) {
      this.errorMessage.set('Email address is missing. Please enter your email.');
      return;
    }

    this.isLoading.set(true);
    this.errorMessage.set('');
    this.auth.verifyEmailOTP(this.email(), code).subscribe({
      next: () => {
        this.isLoading.set(false);
        this.isSuccess.set(true);
        this.toast.showSuccess('Email verified successfully! You can now log in.');
        setTimeout(() => this.router.navigate(['/auth/login']), 2000);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.error || 'Invalid or expired verification code.');
      }
    });
  }

  resendCode() {
    if (this.resendCountdown() > 0 || !this.email()) return;

    this.isLoading.set(true);
    this.errorMessage.set('');
    this.auth.resendVerification(this.email()).subscribe({
      next: () => {
        this.isLoading.set(false);
        this.toast.showSuccess('A fresh verification code has been dispatched to your email.');
        this.startCountdown(60);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err?.error?.error || 'Failed to resend code. Please try again.');
      }
    });
  }

  private startCountdown(seconds: number) {
    this.resendCountdown.set(seconds);
    if (this.countdownTimer) clearInterval(this.countdownTimer);
    this.countdownTimer = setInterval(() => {
      const cur = this.resendCountdown();
      if (cur <= 1) {
        clearInterval(this.countdownTimer);
        this.resendCountdown.set(0);
      } else {
        this.resendCountdown.set(cur - 1);
      }
    }, 1000);
  }
}
