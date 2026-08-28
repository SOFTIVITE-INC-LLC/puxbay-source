import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SupplierPortalService, SupplierProfile, SupplierPayoutAccount, SupplierTeamMember, SupplierDocument, SupplierAPIKey, SupplierWebhook } from '../../services/supplier-portal.service';
import { ToastService } from '../../../../core/services/toast';

@Component({
  selector: 'app-supplier-portal-settings',
  standalone: true,
  imports: [CommonModule, DatePipe, FormsModule],
  templateUrl: './settings.component.html'
})
export class SupplierPortalSettingsComponent implements OnInit {
  portalService = inject(SupplierPortalService);
  private toast = inject(ToastService);

  profile = signal<SupplierProfile | null>(null);
  
  activeTab = signal<'payout' | 'team' | 'documents' | 'profile' | 'api'>('payout');

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

  // Compliance Documents state
  documents = signal<SupplierDocument[]>([]);
  showUploadDocModal = signal<boolean>(false);
  docType = 'business_license';
  docName = '';
  docURL = '';
  docExpiry = '';

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
    this.loadDocuments();
  }

  loadTeam() {
    this.portalService.getTeam().subscribe({
      next: (members) => this.teamMembers.set(members || []),
      error: () => {}
    });
  }

  loadDocuments() {
    this.portalService.getDocuments().subscribe({
      next: (docs) => this.documents.set(docs || []),
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

  submitDocument() {
    if (!this.docName || !this.docURL) {
      this.toast.showError('Please provide a document title and file link');
      return;
    }

    this.portalService.uploadDocument({
      document_type: this.docType,
      document_name: this.docName,
      file_url: this.docURL,
      expiry_date: this.docExpiry ? new Date(this.docExpiry).toISOString() : undefined
    }).subscribe({
      next: (d) => {
        this.toast.showSuccess(`Uploaded ${d.document_name} to compliance vault!`);
        this.showUploadDocModal.set(false);
        this.docName = '';
        this.docURL = '';
        this.docExpiry = '';
        this.loadDocuments();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to upload document')
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

  // ── API Keys & Webhooks ──
  apiKeys = signal<SupplierAPIKey[]>([]);
  webhooks = signal<SupplierWebhook[]>([]);
  newKeyName = '';
  newWebhookURL = '';
  newWebhookEvents = 'po.created,invoice.paid';
  newKeyPlain = signal<string | null>(null);
  loadingAPI = signal<boolean>(false);

  loadAPIData() {
    this.portalService.listApiKeys().subscribe({ next: (k) => this.apiKeys.set(k || []) });
    this.portalService.listWebhooks().subscribe({ next: (w) => this.webhooks.set(w || []) });
  }

  createKey() {
    if (!this.newKeyName) { this.toast.showError('Enter a name for the API key'); return; }
    this.loadingAPI.set(true);
    this.portalService.createApiKey(this.newKeyName).subscribe({
      next: (res) => {
        this.newKeyPlain.set(res.plain_key);
        this.newKeyName = '';
        this.loadAPIData();
        this.loadingAPI.set(false);
      },
      error: (err) => { this.toast.showError(err.error?.error || 'Failed'); this.loadingAPI.set(false); }
    });
  }

  revokeKey(id: string) {
    this.portalService.revokeApiKey(id).subscribe({
      next: () => { this.toast.showSuccess('API key revoked'); this.loadAPIData(); },
      error: (err) => this.toast.showError(err.error?.error || 'Failed')
    });
  }

  createWebhook() {
    if (!this.newWebhookURL) { this.toast.showError('Enter a webhook URL'); return; }
    this.portalService.createWebhook(this.newWebhookURL, this.newWebhookEvents).subscribe({
      next: () => { this.toast.showSuccess('Webhook endpoint registered!'); this.newWebhookURL = ''; this.loadAPIData(); },
      error: (err) => this.toast.showError(err.error?.error || 'Failed')
    });
  }

  deleteWebhook(id: string) {
    this.portalService.deleteWebhook(id).subscribe({
      next: () => { this.toast.showSuccess('Webhook deleted'); this.loadAPIData(); },
      error: (err) => this.toast.showError(err.error?.error || 'Failed')
    });
  }

  switchToApiTab() {
    this.activeTab.set('api');
    this.newKeyPlain.set(null);
    this.loadAPIData();
  }
}
