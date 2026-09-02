import { Component, OnInit, signal, ViewChildren, QueryList, ElementRef, AfterViewInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { BranchService } from '../../../core/services/branch.service';

@Component({
  selector: 'app-two-factor',
  standalone: true,
  imports: [CommonModule, FormsModule, ReactiveFormsModule, RouterModule],
  templateUrl: './two-factor.html',
  styles: ``,
})
export class TwoFactorComponent implements OnInit, AfterViewInit {
  otpDigits = signal<string[]>(['', '', '', '', '', '']);
  username = signal<string>('');
  password = '';
  
  errorMessage = signal<string | null>(null);
  isLoading = signal(false);
  hasCredentials = signal(false);

  // Fallback direct credentials form if user navigated directly
  directForm: FormGroup;

  @ViewChildren('otpInput') otpInputs!: QueryList<ElementRef<HTMLInputElement>>;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private branchService: BranchService,
    private router: Router
  ) {
    this.directForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', Validators.required],
    });
  }

  ngOnInit() {
    const state = typeof window !== 'undefined' ? window.history.state : null;
    if (state && state.username && state.password) {
      this.username.set(state.username);
      this.password = state.password;
      this.hasCredentials.set(true);
    } else {
      this.hasCredentials.set(false);
    }
  }

  ngAfterViewInit() {
    if (this.hasCredentials()) {
      setTimeout(() => {
        this.focusDigit(0);
      }, 100);
    }
  }

  focusDigit(index: number) {
    const inputs = this.otpInputs?.toArray();
    if (inputs && inputs[index]) {
      inputs[index].nativeElement.focus();
      inputs[index].nativeElement.select();
    }
  }

  onDigitInput(event: Event, index: number) {
    const input = event.target as HTMLInputElement;
    let val = input.value.replace(/\D/g, '');

    // Handle full paste into a single box
    if (val.length > 1) {
      this.handlePastedCode(val);
      return;
    }

    const currentDigits = [...this.otpDigits()];
    currentDigits[index] = val;
    this.otpDigits.set(currentDigits);
    this.errorMessage.set(null);

    if (val && index < 5) {
      this.focusDigit(index + 1);
    }

    // Auto submit when all 6 digits entered
    if (currentDigits.every(d => d.length === 1)) {
      this.submit2FA();
    }
  }

  onKeyDown(event: KeyboardEvent, index: number) {
    if (event.key === 'Backspace') {
      const currentDigits = [...this.otpDigits()];
      if (!currentDigits[index] && index > 0) {
        currentDigits[index - 1] = '';
        this.otpDigits.set(currentDigits);
        this.focusDigit(index - 1);
      } else {
        currentDigits[index] = '';
        this.otpDigits.set(currentDigits);
      }
    } else if (event.key === 'ArrowLeft' && index > 0) {
      this.focusDigit(index - 1);
    } else if (event.key === 'ArrowRight' && index < 5) {
      this.focusDigit(index + 1);
    }
  }

  onPaste(event: ClipboardEvent) {
    event.preventDefault();
    const text = event.clipboardData?.getData('text') || '';
    this.handlePastedCode(text);
  }

  handlePastedCode(code: string) {
    const clean = code.replace(/\D/g, '').slice(0, 6);
    if (!clean) return;

    const digits = ['', '', '', '', '', ''];
    for (let i = 0; i < clean.length; i++) {
      digits[i] = clean[i];
    }
    this.otpDigits.set(digits);
    this.errorMessage.set(null);

    if (clean.length === 6) {
      this.focusDigit(5);
      this.submit2FA();
    } else {
      this.focusDigit(clean.length);
    }
  }

  submit2FA() {
    const code = this.otpDigits().join('');
    if (code.length !== 6) {
      this.errorMessage.set('Please enter all 6 digits of your authentication code.');
      return;
    }

    let userEmail = this.username();
    let userPass = this.password;

    if (!this.hasCredentials()) {
      if (!this.directForm.valid) {
        this.errorMessage.set('Please enter your email and password.');
        return;
      }
      userEmail = this.directForm.value.email;
      userPass = this.directForm.value.password;
    }

    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.authService.login({ username: userEmail, password: userPass, totpCode: code }).subscribe({
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
                window.location.href = `https://${userSubdomain}.puxbay.com/dashboard`;
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
            error: () => navigate()
          });
        } else {
          navigate();
        }
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(err.error?.message || err.error?.error || 'Invalid 2FA authentication code. Please check your app and try again.');
        // Reset code on failure
        this.otpDigits.set(['', '', '', '', '', '']);
        setTimeout(() => this.focusDigit(0), 50);
      }
    });
  }

  backToLogin() {
    this.router.navigate(['/auth/login']);
  }
}
