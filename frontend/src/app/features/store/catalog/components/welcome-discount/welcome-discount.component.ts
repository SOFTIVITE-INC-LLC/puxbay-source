import { inject } from '@angular/core';
import { ToastService } from '../../../../../core/store/services/toast.service';
import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-welcome-discount',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './welcome-discount.component.html'
})
export class WelcomeDiscountComponent implements OnInit {
  toastService = inject(ToastService);
  isVisible = signal(false);
  isRevealed = signal(false);
  email = signal('');
  discountCode = signal('WELCOME10');

  ngOnInit() {
    if (typeof localStorage !== 'undefined') {
      const hasSeenWelcome = localStorage.getItem('hasSeenWelcomeDiscount');
      if (!hasSeenWelcome) {
        setTimeout(() => {
          this.isVisible.set(true);
        }, 5000); // Show after 5 seconds
      }
    }
  }

  revealDiscount() {
    if (!this.email() || !this.email().includes('@')) return;
    this.isRevealed.set(true);
    // Real app would send email to backend here
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('hasSeenWelcomeDiscount', 'true');
    }
  }

  close() {
    this.isVisible.set(false);
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('hasSeenWelcomeDiscount', 'true');
    }
  }

  copyCode() {
    if (typeof navigator !== 'undefined' && navigator.clipboard) {
      navigator.clipboard.writeText(this.discountCode());
    }
    this.toastService.show('Code copied to clipboard!', 'success');
  }
}
