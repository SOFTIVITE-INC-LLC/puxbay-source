import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { SupplierPortalService } from '../../services/supplier-portal.service';
import { ToastService } from '../../../../core/services/toast';

@Component({
  selector: 'app-supplier-portal-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './login.component.html'
})
export class SupplierPortalLoginComponent {
  email = '';
  password = '';
  loading = signal(false);

  private portalService = inject(SupplierPortalService);
  private router = inject(Router);
  private toast = inject(ToastService);

  onSubmit() {
    if (!this.email || !this.password) return;
    
    this.loading.set(true);
    this.portalService.login({ email: this.email, password: this.password }).subscribe({
      next: () => {
        this.loading.set(false);
        this.router.navigate(['/supplier-portal/dashboard']);
      },
      error: (err) => {
        this.loading.set(false);
        this.toast.showError(err.error?.error || 'Login failed. Please check your credentials.');
      }
    });
  }
}
