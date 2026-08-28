import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SupplierPortalService, SupplierProfile, SupplierPayoutAccount, SupplierTeamMember } from '../../services/supplier-portal.service';
import { ToastService } from '../../../../core/services/toast';

@Component({
  selector: 'app-supplier-portal-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './settings.component.html'
})
export class SupplierPortalSettingsComponent implements OnInit {
  portalService = inject(SupplierPortalService);
  private toast = inject(ToastService);

  profile = signal<SupplierProfile | null>(null);
  
  activeTab = signal<'payout' | 'team' | 'profile'>('payout');

  // Payout state
  accountType = 'bank';
  bankName = '';
  accountNumber = '';
  accountName = '';
  routingCode = '';
  momoNetwork = 'MTN';
  momoNumber = '';

  // Team state
  teamMembers = signal<SupplierTeamMember[]>([]);
  showInviteModal = signal<boolean>(false);
  newMemberName = '';
  newMemberEmail = '';
  newMemberRole: 'admin' | 'finance' | 'warehouse' | 'sales' = 'warehouse';
  newMemberPhone = '';

  loading = signal<boolean>(false);

  ngOnInit() {
    this.portalService.currentSupplier$.subscribe(s => {
      if (s) this.profile.set(s);
    });

    this.portalService.getPayoutAccount().subscribe({
      next: (acc) => {
        if (acc && acc.account_type) {
          this.accountType = acc.account_type;
          this.bankName = acc.bank_name || '';
          this.accountNumber = acc.account_number || '';
          this.accountName = acc.account_name || '';
          this.routingCode = acc.routing_code || '';
          this.momoNetwork = acc.momo_network || 'MTN';
          this.momoNumber = acc.momo_number || '';
        }
      }
    });

    this.loadTeam();
  }

  loadTeam() {
    this.portalService.getTeam().subscribe({
      next: (members) => this.teamMembers.set(members || []),
      error: () => {}
    });
  }

  inviteTeamMember() {
    if (!this.newMemberName || !this.newMemberEmail) return;

    this.portalService.inviteTeamMember({
      full_name: this.newMemberName,
      email: this.newMemberEmail,
      role: this.newMemberRole,
      phone: this.newMemberPhone || undefined
    }).subscribe({
      next: (created) => {
        this.toast.showSuccess(`Invited ${created.full_name} (${created.role}) to team!`);
        this.showInviteModal.set(false);
        this.newMemberName = '';
        this.newMemberEmail = '';
        this.newMemberPhone = '';
        this.loadTeam();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to invite team member')
    });
  }

  savePayout() {
    const payload: Partial<SupplierPayoutAccount> = {
      account_type: this.accountType,
      bank_name: this.accountType === 'bank' ? this.bankName : undefined,
      account_number: this.accountType === 'bank' ? this.accountNumber : undefined,
      account_name: this.accountType === 'bank' ? this.accountName : undefined,
      routing_code: this.accountType === 'bank' ? this.routingCode : undefined,
      momo_network: this.accountType === 'momo' ? this.momoNetwork : undefined,
      momo_number: this.accountType === 'momo' ? this.momoNumber : undefined,
      is_default: true
    };

    this.portalService.savePayoutAccount(payload).subscribe({
      next: () => this.toast.showSuccess('Payout account settings saved successfully!'),
      error: (err) => this.toast.showError(err.error?.error || 'Failed to save payout settings')
    });
  }
}
