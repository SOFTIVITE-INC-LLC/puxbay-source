import { ToastService } from '../../../core/services/toast';
import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SecurityService, Setup2FAResult } from '../../../core/services/security.service';
import { AlertService } from '../../../core/services/alert.service';
import { QRCodeComponent } from 'angularx-qrcode';

@Component({
  selector: 'app-security',
  standalone: true,
  imports: [CommonModule, FormsModule, QRCodeComponent],
  templateUrl: './security.html',
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
export class Security implements OnInit {
  toastService = inject(ToastService);
  securityService = inject(SecurityService);
  private alertService = inject(AlertService);

  activeTab = signal<'audit' | '2fa' | 'backup'>('audit');
  
  // 2FA State
  twoFaSetup = signal<Setup2FAResult | null>(null);
  verificationCode = signal('');
  twoFaEnabled = signal(false);
  
  // Backup State
  backupMessage = signal('');

  // Filter State
  searchQuery = signal('');
  actionFilter = signal('ALL');

  filteredLogs = computed(() => {
    let logs = this.securityService.auditLogs();
    const query = this.searchQuery().trim().toLowerCase();
    const action = this.actionFilter();

    if (action !== 'ALL') {
      logs = logs.filter(l => l.action?.toUpperCase() === action);
    }
    if (query) {
      logs = logs.filter(l => 
        (l.model_name && l.model_name.toLowerCase().includes(query)) ||
        (l.user?.username && l.user.username.toLowerCase().includes(query)) ||
        (l.action && l.action.toLowerCase().includes(query)) ||
        (l.object_id && l.object_id.toLowerCase().includes(query))
      );
    }
    return logs;
  });

  // Pagination State
  currentPage = signal(1);
  pageSize = signal(10);
  totalPages = computed(() => Math.ceil(this.securityService.totalLogs() / this.pageSize()));

  ngOnInit() {
    this.loadAuditLogs();
  }

  loadAuditLogs() {
    this.securityService.getAuditLogs(this.currentPage(), this.pageSize()).subscribe();
  }

  changePage(delta: number) {
    const newPage = this.currentPage() + delta;
    if (newPage > 0 && newPage <= this.totalPages()) {
      this.currentPage.set(newPage);
      this.loadAuditLogs();
    }
  }

  begin2FASetup() {
    this.securityService.setup2FA().subscribe(res => {
      this.twoFaSetup.set(res);
    });
  }

  verifyAndEnable2FA() {
    if (!this.verificationCode()) return;
    this.securityService.verify2FA(this.verificationCode()).subscribe({
      next: (res) => {
        this.twoFaEnabled.set(true);
        this.twoFaSetup.set(null);
        this.toastService.showSuccess(res.message);
      },
      error: () => this.toastService.showError('Invalid verification code.')
    });
  }

  async disable2FA() {
    if (await this.alertService.confirm('Are you sure you want to disable 2FA? This will reduce your account security.', 'Disable 2FA')) {
      this.securityService.disable2FA({}).subscribe(() => {
        this.twoFaEnabled.set(false);
        this.toastService.showSuccess('2FA has been disabled.');
      });
    }
  }

  triggerBackup() {
    this.securityService.backupDashboard().subscribe(res => {
      this.backupMessage.set('Backup triggered successfully.');
      setTimeout(() => this.backupMessage.set(''), 3000);
    });
  }
}
