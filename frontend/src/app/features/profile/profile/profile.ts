import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { ProfileService, UpdateMeInput } from '../../../core/services/profile.service';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './profile.html',
  styles: `
    @keyframes gradientShift {
      0% { background-position: 0% 50%; }
      50% { background-position: 100% 50%; }
      100% { background-position: 0% 50%; }
    }
    @keyframes blob {
      0% { transform: translate(0px, 0px) scale(1); }
      33% { transform: translate(30px, -50px) scale(1.1); }
      66% { transform: translate(-20px, 20px) scale(0.9); }
      100% { transform: translate(0px, 0px) scale(1); }
    }
    .animated-gradient-text {
      background-size: 300% 300%;
      animation: gradientShift 6s ease infinite;
    }
    .animate-blob {
      animation: blob 7s infinite;
    }
    .animation-delay-2000 {
      animation-delay: 2s;
    }
    .animation-delay-4000 {
      animation-delay: 4s;
    }
    .glass-panel {
      background: rgba(255, 255, 255, 0.7);
      backdrop-filter: blur(24px);
      -webkit-backdrop-filter: blur(24px);
      border: 1px solid rgba(255, 255, 255, 0.8);
      box-shadow: 0 4px 30px rgba(0, 0, 0, 0.05);
    }
    :host-context(.dark) .glass-panel {
      background: rgba(24, 24, 27, 0.6);
      border: 1px solid rgba(255, 255, 255, 0.08);
      box-shadow: 0 4px 30px rgba(0, 0, 0, 0.3);
    }
    .input-premium {
      @apply w-full pl-11 pr-4 py-3.5 rounded-2xl bg-white/60 dark:bg-black/30 text-zinc-900 dark:text-white border border-zinc-200/80 dark:border-zinc-700/50 focus:outline-none focus:border-indigo-500 focus:ring-4 focus:ring-indigo-500/20 font-bold transition-all shadow-sm backdrop-blur-sm;
    }
  `,
})
export class Profile implements OnInit {
  profileService = inject(ProfileService);

  // Editable form state
  firstName = signal('');
  lastName = signal('');
  currentPassword = signal('');
  newPassword = signal('');
  confirmPassword = signal('');
  
  posPin = signal('');
  confirmPosPin = signal('');

  activeTab = signal<'general' | 'security'>('general');
  successMessage = signal<string | null>(null);

  ngOnInit() {
    this.profileService.getProfile().subscribe(() => {
      const p = this.profileService.profile();
      if (p) {
        this.firstName.set(p.first_name || '');
        this.lastName.set(p.last_name || '');
      }
    });
  }

  get userInitials(): string {
    const p = this.profileService.profile();
    if (!p) return '?';
    return `${(p.first_name || 'U')[0]}${(p.last_name || 'X')[0]}`.toUpperCase();
  }

  saveGeneral() {
    const input: UpdateMeInput = {
      first_name: this.firstName(),
      last_name: this.lastName(),
    };
    this.profileService.updateMe(input).subscribe({
      next: () => {
        this.successMessage.set('Profile updated successfully!');
        setTimeout(() => this.successMessage.set(null), 3000);
      }
    });
  }

  savePassword() {
    if (this.newPassword() !== this.confirmPassword()) {
      this.profileService.error.set('New passwords do not match.');
      return;
    }
    if (this.newPassword().length < 8) {
      this.profileService.error.set('New password must be at least 8 characters.');
      return;
    }
    const input: UpdateMeInput = {
      current_password: this.currentPassword(),
      new_password: this.newPassword(),
    };
    this.profileService.updateMe(input).subscribe({
      next: () => {
        this.currentPassword.set('');
        this.newPassword.set('');
        this.confirmPassword.set('');
        this.successMessage.set('Password changed successfully! Please log in again.');
        setTimeout(() => this.successMessage.set(null), 5000);
      }
    });
  }

  savePosPin() {
    if (this.posPin().length !== 4) {
      this.profileService.error.set('POS PIN must be exactly 4 digits.');
      return;
    }
    if (this.posPin() !== this.confirmPosPin()) {
      this.profileService.error.set('POS PINs do not match.');
      return;
    }
    this.profileService.setPosPin(this.posPin()).subscribe({
      next: () => {
        this.posPin.set('');
        this.confirmPosPin.set('');
        this.successMessage.set('POS PIN updated successfully!');
        setTimeout(() => this.successMessage.set(null), 3000);
      }
    });
  }
}
