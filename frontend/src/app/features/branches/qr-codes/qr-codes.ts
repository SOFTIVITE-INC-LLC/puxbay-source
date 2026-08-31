import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { BranchService } from '../../../core/services/branch.service';
import { TenantStore } from '../../../core/services/tenant.store';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';
import { Branch } from '../../../core/models/branch.models';

@Component({
  selector: 'app-qr-codes',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './qr-codes.html',
})
export class QrCodes implements OnInit {
  private branchService = inject(BranchService);
  private tenantStore = inject(TenantStore);
  private storefrontSettings = inject(StorefrontSettingsService);

  branches = this.branchService.branches;
  selectedBranch = signal<Branch | null>(null);
  loading = signal(false);

  subdomain = computed(() =>
    this.tenantStore.subdomain() ||
    (typeof window !== 'undefined' ? localStorage.getItem('dev_tenant') || '' : '')
  );

  // The storefront slug the tenant configured (e.g. "main", "shop")
  storefrontSlug = computed(() => this.storefrontSettings.settings()?.slug || '');

  baseUrl = computed(() => {
    if (typeof window === 'undefined') return '';
    const sub = this.subdomain();
    const host = window.location.host;
    const hostname = window.location.hostname;
    const proto = window.location.protocol;
    const isIP = /^(\d{1,3}\.){3}\d{1,3}$/.test(hostname);

    if (!sub || hostname === 'localhost' || isIP) {
      // Dev/local: use current host + ?tenant= param
      return `${proto}//${host}`;
    }
    // Production: build subdomain.domain.tld URL
    const parts = hostname.split('.');
    if (parts.length >= 2 && parts[0] !== 'www') {
      parts[0] = sub;
    } else {
      parts.unshift(sub);
    }
    const port = window.location.port ? `:${window.location.port}` : '';
    return `${proto}//${parts.join('.')}${port}`;
  });

  kioskLinkUrl = computed(() => {
    const branch = this.selectedBranch();
    const base = this.baseUrl();
    const sub = this.subdomain();
    const hostname = typeof window !== 'undefined' ? window.location.hostname : '';
    const isLocalOrIP = hostname === 'localhost' || /^(\d{1,3}\.){3}\d{1,3}$/.test(hostname);
    const path = branch ? `/kiosk/${branch.id}` : '/kiosk';
    // On local/IP always append ?tenant= so the interceptor can resolve the schema
    const query = isLocalOrIP && sub ? `?tenant=${encodeURIComponent(sub)}` : '';
    return `${base}${path}${query}`;
  });

  storefrontLinkUrl = computed(() => {
    const branch = this.selectedBranch();
    const base = this.baseUrl();
    const sub = this.subdomain();
    const hostname = typeof window !== 'undefined' ? window.location.hostname : '';
    const isLocalOrIP = hostname === 'localhost' || /^(\d{1,3}\.){3}\d{1,3}$/.test(hostname);

    // Points directly to the full e-commerce storefront (/store)
    const path = branch ? `/store/${branch.id}` : '/store';
    const query = isLocalOrIP && sub ? `?tenant=${encodeURIComponent(sub)}` : '';
    return `${base}${path}${query}`;
  });

  kioskQrUrl = computed(() => {
    const link = this.kioskLinkUrl();
    return `https://api.qrserver.com/v1/create-qr-code/?size=280x280&data=${encodeURIComponent(link)}&margin=10&qzone=2&format=svg`;
  });

  storefrontQrUrl = computed(() => {
    const link = this.storefrontLinkUrl();
    return `https://api.qrserver.com/v1/create-qr-code/?size=280x280&data=${encodeURIComponent(link)}&margin=10&qzone=2&format=svg`;
  });

  ngOnInit() {
    this.loading.set(true);
    this.branchService.getBranches().subscribe({
      next: (branches) => {
        this.loading.set(false);
        if (branches.length > 0) {
          this.selectedBranch.set(branches[0]);
        }
      },
      error: () => this.loading.set(false)
    });
    // Load storefront settings to get the configured slug for QR URL
    this.storefrontSettings.loadSettings().subscribe();
  }

  selectBranch(branch: Branch) {
    this.selectedBranch.set(branch);
  }

  printQr(type: 'kiosk' | 'storefront') {
    const branch = this.selectedBranch();
    const branchName = branch?.name || 'All Branches';
    const title = type === 'kiosk' ? 'Self-Service Kiosk' : 'Online Storefront';
    const qrUrl = type === 'kiosk' ? this.kioskQrUrl() : this.storefrontQrUrl();
    const linkUrl = type === 'kiosk' ? this.kioskLinkUrl() : this.storefrontLinkUrl();
    const icon = type === 'kiosk' ? 'shopping_cart' : 'language';
    const color = type === 'kiosk' ? '#005b96' : '#10b981';

    const win = window.open('', '_blank');
    if (!win) return;
    win.document.write(`
      <!DOCTYPE html>
      <html>
      <head>
        <meta charset="UTF-8">
        <title>QR Code – ${title}</title>
        <style>
          * { box-sizing: border-box; margin: 0; padding: 0; }
          body {
            font-family: 'Segoe UI', system-ui, sans-serif;
            background: #fff;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            padding: 40px 20px;
          }
          .card {
            text-align: center;
            width: 360px;
            border: 2px solid ${color};
            border-radius: 20px;
            padding: 32px 28px 28px;
            box-shadow: 0 4px 40px rgba(0,0,0,0.08);
          }
          .icon { font-size: 40px; margin-bottom: 8px; }
          .brand {
            font-size: 11px;
            font-weight: 900;
            letter-spacing: 0.2em;
            text-transform: uppercase;
            color: ${color};
            margin-bottom: 4px;
          }
          .title {
            font-size: 22px;
            font-weight: 900;
            color: #0f172a;
            margin-bottom: 4px;
          }
          .branch {
            font-size: 13px;
            color: #64748b;
            font-weight: 600;
            margin-bottom: 24px;
          }
          .qr-wrap {
            background: #fff;
            border: 1px solid #e2e8f0;
            border-radius: 16px;
            display: inline-block;
            padding: 16px;
            margin-bottom: 20px;
          }
          .qr-wrap img { display: block; width: 220px; height: 220px; }
          .instruction {
            font-size: 13px;
            font-weight: 700;
            color: #64748b;
            margin-bottom: 10px;
          }
          .url {
            font-size: 10px;
            color: #94a3b8;
            word-break: break-all;
            font-family: monospace;
          }
          .footer {
            margin-top: 24px;
            padding-top: 20px;
            border-top: 1px solid #f1f5f9;
            font-size: 10px;
            font-weight: 900;
            letter-spacing: 0.1em;
            text-transform: uppercase;
            color: #cbd5e1;
          }
          @media print {
            body { padding: 0; }
            .card { box-shadow: none; }
          }
        </style>
      </head>
      <body>
        <div class="card">
          <div class="icon">${icon}</div>
          <div class="brand">Puxbay</div>
          <div class="title">${title}</div>
          <div class="branch">${branchName}</div>
          <div class="qr-wrap">
            <img src="${qrUrl}" alt="QR Code" />
          </div>
          <div class="instruction">Scan to ${type === 'kiosk' ? 'start ordering' : 'browse our store'}</div>
          <div class="url">${linkUrl}</div>
          <div class="footer">© ${new Date().getFullYear()} Puxbay Inc.</div>
        </div>
        <script>window.onload = () => setTimeout(() => { window.print(); }, 600 );</script>
      </body>
      </html>
    `);
    win.document.close();
  }

  copyLink(url: string) {
    navigator.clipboard?.writeText(url).catch(() => {});
  }
}
