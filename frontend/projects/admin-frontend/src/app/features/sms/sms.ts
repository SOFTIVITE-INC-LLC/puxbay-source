import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AdminSMSService, SMSGatewayConfig, AdminSenderIDEntry } from '../../services/sms.service';
import { AlertService } from '../../services/alert.service';

@Component({
  selector: 'app-sms',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './sms.html',
  host: { class: 'block w-full min-h-full' }
})
export class SmsComponent implements OnInit {
  private smsService = inject(AdminSMSService);
  private alertService = inject(AlertService);

  activeTab = signal<'gateway' | 'sender_ids'>('gateway');

  // Gateway Config State
  config = signal<Partial<SMSGatewayConfig>>({
    provider: 'arkesel',
    default_sender_id: 'PUXBAY',
    price_per_sms: 0.20,
    price_currency: 'GHS',
    is_active: true
  });
  isLoadingConfig = signal(true);
  isSavingConfig = signal(false);

  // Live Calculator State
  calculatorAmount = signal<number>(50);

  // Sender IDs State
  senderIDs = signal<AdminSenderIDEntry[]>([]);
  isLoadingSenderIDs = signal(true);
  senderIDFilter = signal<'pending' | 'approved' | 'rejected' | ''>('pending');
  searchQuery = signal<string>('');

  // Rejection Modal State
  isRejectModalOpen = signal(false);
  rejectTarget = signal<AdminSenderIDEntry | null>(null);
  rejectReason = signal('');
  isRejecting = signal(false);

  // Computed KPIs
  pendingCount = computed(() => this.senderIDs().filter(s => s.status === 'pending').length);
  approvedCount = computed(() => this.senderIDs().filter(s => s.status === 'approved').length);
  rejectedCount = computed(() => this.senderIDs().filter(s => s.status === 'rejected').length);
  totalCount = computed(() => this.senderIDs().length);

  // Filtered Sender IDs
  filteredSenderIDs = computed(() => {
    const q = this.searchQuery().toLowerCase().trim();
    const filter = this.senderIDFilter();
    return this.senderIDs().filter(item => {
      const matchFilter = !filter || item.status === filter;
      const matchSearch = !q ||
        (item.sender_id || '').toLowerCase().includes(q) ||
        (item.tenant_name || '').toLowerCase().includes(q) ||
        (item.tenant_id || '').toLowerCase().includes(q) ||
        (item.purpose || '').toLowerCase().includes(q);
      return matchFilter && matchSearch;
    });
  });

  ngOnInit() {
    this.loadConfig();
    this.loadSenderIDs();
  }

  loadConfig() {
    this.isLoadingConfig.set(true);
    this.smsService.getConfig().subscribe({
      next: (res) => {
        this.config.set({ ...res });
        this.isLoadingConfig.set(false);
      },
      error: () => {
        this.isLoadingConfig.set(false);
      }
    });
  }

  loadSenderIDs() {
    this.isLoadingSenderIDs.set(true);
    this.smsService.getSenderIDs().subscribe({
      next: (res) => {
        this.senderIDs.set(res || []);
        this.isLoadingSenderIDs.set(false);
      },
      error: () => {
        this.isLoadingSenderIDs.set(false);
      }
    });
  }

  setFilter(filter: 'pending' | 'approved' | 'rejected' | '') {
    this.senderIDFilter.set(filter);
  }

  saveConfig() {
    this.isSavingConfig.set(true);
    this.smsService.updateConfig(this.config()).subscribe({
      next: () => {
        this.alertService.success('SMS Gateway configuration saved successfully.');
        this.isSavingConfig.set(false);
        this.loadConfig();
      },
      error: (err) => {
        this.alertService.error(err.error?.error || 'Failed to save configuration.');
        this.isSavingConfig.set(false);
      }
    });
  }

  calculateSimulatedCredits(amt: number): number {
    const price = this.config().price_per_sms || 0.20;
    if (price <= 0) return 0;
    return Math.floor(amt / price);
  }

  async approveSenderID(entry: AdminSenderIDEntry) {
    const confirmed = await this.alertService.confirm({
      title: 'Approve Sender ID',
      message: `Approve custom Sender ID "${entry.sender_id}" for tenant "${entry.tenant_name}"?`,
      confirmText: 'Approve Sender ID',
      type: 'info'
    });
    if (confirmed) {
      this.smsService.approveSenderID(entry.id, entry.tenant_id).subscribe({
        next: () => {
          this.alertService.success(`Sender ID "${entry.sender_id}" approved.`);
          this.loadSenderIDs();
        },
        error: (err) => {
          this.alertService.error(err.error?.error || 'Failed to approve Sender ID.');
        }
      });
    }
  }

  openRejectModal(entry: AdminSenderIDEntry) {
    this.rejectTarget.set(entry);
    this.rejectReason.set('');
    this.isRejectModalOpen.set(true);
  }

  closeRejectModal() {
    this.isRejectModalOpen.set(false);
    this.rejectTarget.set(null);
    this.rejectReason.set('');
  }

  confirmReject() {
    const target = this.rejectTarget();
    const reason = this.rejectReason().trim();
    if (!target || !reason) {
      this.alertService.warning('Please provide a reason for rejection.');
      return;
    }

    this.isRejecting.set(true);
    this.smsService.rejectSenderID(target.id, target.tenant_id, reason).subscribe({
      next: () => {
        this.alertService.success(`Sender ID "${target.sender_id}" rejected.`);
        this.isRejecting.set(false);
        this.closeRejectModal();
        this.loadSenderIDs();
      },
      error: (err) => {
        this.alertService.error(err.error?.error || 'Failed to reject Sender ID.');
        this.isRejecting.set(false);
      }
    });
  }

  async deleteSenderID(entry: AdminSenderIDEntry) {
    const confirmed = await this.alertService.confirm({
      title: 'Delete Sender ID',
      message: `Permanently delete Sender ID "${entry.sender_id}" for tenant "${entry.tenant_name}"?`,
      confirmText: 'Delete Sender ID',
      type: 'danger'
    });
    if (confirmed) {
      this.smsService.deleteSenderID(entry.id, entry.tenant_id).subscribe({
        next: () => {
          this.alertService.success(`Sender ID "${entry.sender_id}" deleted.`);
          this.loadSenderIDs();
        },
        error: (err) => {
          this.alertService.error(err.error?.error || 'Failed to delete Sender ID.');
        }
      });
    }
  }
}
