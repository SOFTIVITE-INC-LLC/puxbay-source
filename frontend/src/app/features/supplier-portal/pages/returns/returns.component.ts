import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SupplierPortalService, SupplierRMA } from '../../services/supplier-portal.service';
import { AppCurrencyPipe } from '../../../../core/pipes/app-currency.pipe';
import { ToastService } from '../../../../core/services/toast';

@Component({
  selector: 'app-supplier-portal-returns',
  standalone: true,
  imports: [CommonModule, AppCurrencyPipe, FormsModule],
  templateUrl: './returns.component.html'
})
export class SupplierPortalReturnsComponent implements OnInit {
  portalService = inject(SupplierPortalService);
  private toast = inject(ToastService);

  rmas = signal<SupplierRMA[]>([]);
  loading = signal<boolean>(false);

  selectedRMA = signal<SupplierRMA | null>(null);
  replacementNotes = '';
  showReplacementModal = signal<boolean>(false);

  ngOnInit() {
    this.loadRMAs();
  }

  loadRMAs() {
    this.loading.set(true);
    this.portalService.getRMAs().subscribe({
      next: (res) => {
        this.rmas.set(res || []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });
  }

  openReplacementModal(rma: SupplierRMA) {
    this.selectedRMA.set(rma);
    this.replacementNotes = `Replacement batch dispatched for RMA #${rma.rma_number}`;
    this.showReplacementModal.set(true);
  }

  submitReplacement() {
    const rma = this.selectedRMA();
    if (!rma) return;

    this.portalService.dispatchRMAReplacement(rma.id, this.replacementNotes).subscribe({
      next: () => {
        this.toast.showSuccess(`Replacement dispatched for RMA #${rma.rma_number}!`);
        this.showReplacementModal.set(false);
        this.loadRMAs();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to update RMA')
    });
  }

  statusClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s === 'replacement_dispatched' || s === 'refunded') return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30';
    if (s === 'approved' || s === 'pending') return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30';
    if (s === 'rejected') return 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30';
    return 'bg-slate-100 dark:bg-zinc-800 text-slate-700 dark:text-zinc-300 border-slate-200 dark:border-zinc-700';
  }
}
