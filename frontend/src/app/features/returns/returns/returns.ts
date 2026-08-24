import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule, CurrencyPipe, DatePipe, SlicePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ReturnService } from '../../../core/services/return.service';

@Component({
  selector: 'app-returns',
  standalone: true,
  imports: [CommonModule, FormsModule, DatePipe, SlicePipe, AppCurrencyPipe],
  templateUrl: './returns.html',
})
export class Returns implements OnInit {
  returnService = inject(ReturnService);
  route = inject(ActivatedRoute);
  router = inject(Router);

  // UI State
  isModalOpen = signal(false);
  saving = signal(false);
  processing = signal<string | null>(null);
  searchQuery = signal('');
  statusFilter = signal('');

  // New return form
  newReturn = signal({
    order_id: '',
    reason: '',
    reason_detail: '',
    refund_method: 'cash',
    refund_amount: 0,
  });

  // Computed stats
  pendingCount = computed(() =>
    this.returnService.returns().filter(r => r.status === 'pending').length
  );
  completedCount = computed(() =>
    this.returnService.returns().filter(r => r.status === 'completed').length
  );
  totalRefunded = computed(() =>
    this.returnService
      .returns()
      .filter(r => r.status === 'completed')
      .reduce((sum, r) => sum + (r.total_refund ?? 0), 0)
  );
  filteredReturns = computed(() => {
    const q = this.searchQuery().toLowerCase();
    const s = this.statusFilter();
    return this.returnService.returns().filter(r => {
      const matchesQuery = !q || r.id?.toLowerCase().includes(q) || r.order_id?.toLowerCase().includes(q);
      const matchesStatus = !s || r.status === s;
      return matchesQuery && matchesStatus;
    });
  });

  ngOnInit() {
    this.returnService.getReturns().subscribe();
    
    // Handle Quick Return from Orders
    this.route.queryParams.subscribe(params => {
      if (params['new'] === 'true' && params['order_id']) {
        this.newReturn.set({
          order_id: params['order_id'],
          reason: '',
          reason_detail: '',
          refund_method: 'cash',
          refund_amount: 0,
        });
        this.isModalOpen.set(true);
        // Clear query params so refresh doesn't reopen modal
        this.router.navigate([], { queryParams: {} });
      }
    });
  }

  openNewReturnModal() {
    this.newReturn.set({ order_id: '', reason: '', reason_detail: '', refund_method: 'cash', refund_amount: 0 });
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  submitNewReturn() {
    const r = this.newReturn();
    if (!r.order_id || !r.reason) return;
    this.saving.set(true);
    this.returnService.createReturn(r as any).subscribe({
      next: () => {
        this.saving.set(false);
        this.closeModal();
        this.returnService.getReturns().subscribe();
      },
      error: () => this.saving.set(false),
    });
  }

  approve(id: string) {
    this.processing.set(id);
    this.returnService.approveReturn(id).subscribe({
      next: () => {
        this.processing.set(null);
        this.returnService.getReturns().subscribe();
      },
      error: () => this.processing.set(null),
    });
  }

  reject(id: string) {
    this.processing.set(id);
    this.returnService.rejectReturn(id).subscribe({
      next: () => {
        this.processing.set(null);
        this.returnService.getReturns().subscribe();
      },
      error: () => this.processing.set(null),
    });
  }

  processRefund(id: string) {
    this.processing.set(id);
    this.returnService.processRefund(id).subscribe({
      next: () => {
        this.processing.set(null);
        this.returnService.getReturns().subscribe();
      },
      error: () => this.processing.set(null),
    });
  }

  getStatusClass(status: string): string {
    switch (status) {
      case 'pending':   return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400';
      case 'approved':  return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400';
      case 'completed': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400';
      case 'rejected':  return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400';
      default:          return 'bg-zinc-100 text-zinc-600';
    }
  }
}
