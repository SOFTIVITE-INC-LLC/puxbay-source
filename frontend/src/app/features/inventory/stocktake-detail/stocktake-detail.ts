import { Component, OnInit, inject, OnDestroy, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../../core/services/api.service';
import { ToastService } from '../../../core/services/toast';
import { AlertService } from '../../../core/services/alert.service';

@Component({
  selector: 'app-stocktake-detail',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './stocktake-detail.html',
  styles: [':host { display: block; }']
})
export class StocktakeDetailComponent implements OnInit, OnDestroy {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private api = inject(ApiService);
  private toast = inject(ToastService);
  private alertService = inject(AlertService);

  id = signal<string | null>(null);
  session = signal<any>(null);
  loading = signal<boolean>(true);
  portalUrl = signal<string>('');
  qrUrl = signal<string>('');
  
  private pollInterval: any;

  ngOnInit() {
    this.id.set(this.route.snapshot.paramMap.get('id'));
    if (this.id()) {
      this.loadSession();
      // Poll every 5 seconds for live analytics updates
      this.pollInterval = setInterval(() => this.loadSession(true), 5000);
    }
  }

  ngOnDestroy() {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
    }
  }

  loadSession(silent = false) {
    if (!silent) this.loading.set(true);
    this.api.get(`/inventory/stocktakes/${this.id()}`).subscribe({
      next: (res: any) => {
        try {
          this.session.set(res?.data || res);
          
          // Build portal URL based on current host
          const baseUrl = window.location.origin;
          const token = this.session()?.access_token || '';
          this.portalUrl.set(`${baseUrl}/stocktake/portal/${token}`);
          this.qrUrl.set(`https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=${encodeURIComponent(this.portalUrl())}`);
        } catch (e) {
          console.error('Error processing session:', e);
        } finally {
          if (!silent) {
            this.loading.set(false);
          }
        }
      },
      error: (err) => {
        console.error('Failed to load session:', err);
        if (!silent) this.toast.showError('Failed to load stocktake session: ' + (err.message || 'Unknown error'));
        if (!silent) {
          this.loading.set(false);
        }
        
        // If it fails on the first load, also stop polling to prevent spamming errors
        if (!silent && this.pollInterval) {
          clearInterval(this.pollInterval);
        }
      }
    });
  }

  async finalizeSession() {
    if (await this.alertService.confirm('Are you sure you want to finalize this stocktake? All discrepancies will be committed as stock adjustments.', 'Finalize Stocktake')) {
      this.api.post(`/inventory/stocktakes/${this.id()}/finalize`, {}).subscribe({
        next: () => {
          this.toast.showSuccess('Stocktake finalized successfully!');
          this.router.navigate(['/inventory']);
        },
        error: (err) => {
          console.error('Finalize error:', err);
          this.toast.showError('Failed to finalize stocktake');
        }
      });
    }
  }

  copyPortalUrl() {
    if (this.portalUrl()) {
      navigator.clipboard.writeText(this.portalUrl()).then(() => {
        this.toast.showSuccess('Portal URL copied to clipboard');
      }).catch(err => {
        console.error('Could not copy text: ', err);
        this.toast.showError('Failed to copy URL');
      });
    }
  }

  openPortalUrl() {
    if (this.portalUrl()) {
      window.open(this.portalUrl(), '_blank');
    }
  }

  goBack() {
    this.router.navigate(['/inventory']);
  }
}
