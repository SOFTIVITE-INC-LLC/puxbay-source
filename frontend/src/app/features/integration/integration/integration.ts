import { ToastService } from '../../../core/services/toast';
import { Component, OnInit, inject, signal } from '@angular/core';
import { ApiService } from '../../../core/services/api.service';
import { CommonModule } from '@angular/common';

interface TenantIntegration {
  provider: string;
  is_active: boolean;
  last_sync_at?: string;
}

@Component({
  selector: 'app-integration',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './integration.html',
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
export class Integration implements OnInit {
  toastService = inject(ToastService);
  private api = inject(ApiService);
  
  integrations = signal<TenantIntegration[]>([]);
  xeroIntegration = signal<TenantIntegration | null>(null);
  isSyncing = signal(false);

  ngOnInit() {
    this.fetchIntegrations();
  }

  fetchIntegrations() {
    this.api.get<TenantIntegration[]>('/integrations').subscribe({
      next: (data) => {
        this.integrations.set(data);
        const xero = data.find(i => i.provider === 'xero');
        if (xero) this.xeroIntegration.set(xero);
      },
      error: (err) => console.error('Failed to load integrations', err)
    });
  }

  connectXero() {
    // Redirects to backend which handles OAuth flow
    window.location.href = '/api/v1/integrations/xero/connect';
  }

  syncXero() {
    if (this.isSyncing()) return;
    this.isSyncing.set(true);
    this.api.post('/integrations/xero/sync', {}).subscribe({
      next: () => {
        this.toastService.showSuccess('Accounting sync has been queued successfully!');
        this.isSyncing.set(false);
      },
      error: (err) => {
        console.error('Sync failed', err);
        this.toastService.showError('Failed to queue accounting sync.');
        this.isSyncing.set(false);
      }
    });
  }
}
