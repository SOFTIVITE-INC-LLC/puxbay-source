import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule, DatePipe, SlicePipe, TitleCasePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ReturnService } from '../../../core/services/return.service';
import { Return } from '../../../core/models/order.models';

@Component({
  selector: 'app-returns',
  standalone: true,
  imports: [CommonModule, FormsModule, DatePipe, SlicePipe, TitleCasePipe, AppCurrencyPipe],
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

  // Detail slide-over
  selectedReturn = signal<Return | null>(null);

  // New return form
  newReturn = signal<{
    order_id: string;
    reason: string;
    reason_detail: string;
    refund_method: string;
    refund_amount: number;
    items: { product_id: string; quantity: number; reason: string; restock: boolean }[];
  }>({
    order_id: '',
    reason: '',
    reason_detail: '',
    refund_method: 'cash',
    refund_amount: 0,
    items: [],
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
          items: [],
        });
        this.isModalOpen.set(true);
        // Clear query params so refresh doesn't reopen modal
        this.router.navigate([], { queryParams: {} });
      }
    });
  }

  // ── Modal ──────────────────────────────────────────────
  openNewReturnModal() {
    this.newReturn.set({ order_id: '', reason: '', reason_detail: '', refund_method: 'cash', refund_amount: 0, items: [] });
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  // ── Return Items ───────────────────────────────────────
  addReturnItem() {
    this.newReturn.update(r => ({
      ...r,
      items: [...(r.items ?? []), { product_id: '', quantity: 1, reason: '', restock: false }],
    }));
  }

  removeReturnItem(index: number) {
    this.newReturn.update(r => ({
      ...r,
      items: r.items.filter((_, i) => i !== index),
    }));
  }

  updateReturnItem(index: number, field: string, value: any) {
    this.newReturn.update(r => ({
      ...r,
      items: r.items.map((item, i) => i === index ? { ...item, [field]: value } : item),
    }));
  }

  // ── Submit ─────────────────────────────────────────────
  submitNewReturn() {
    const r = this.newReturn();
    if (!r.order_id || !r.reason) return;
    this.saving.set(true);
    const payload: any = {
      order_id: r.order_id,
      reason: r.reason,
      reason_detail: r.reason_detail,
      refund_method: r.refund_method,
      refund_amount: r.refund_amount,
    };
    if (r.items && r.items.length > 0) {
      payload.items = r.items.map(item => ({
        product_id: item.product_id || undefined,
        quantity: item.quantity,
        reason: item.reason || undefined,
        restock: item.restock,
      }));
    }
    this.returnService.createReturn(payload).subscribe({
      next: () => {
        this.saving.set(false);
        this.closeModal();
        this.returnService.getReturns().subscribe();
      },
      error: () => this.saving.set(false),
    });
  }

  // ── Detail Panel ───────────────────────────────────────
  openDetail(ret: Return) {
    this.selectedReturn.set(ret);
  }

  closeDetail() {
    this.selectedReturn.set(null);
  }

  // ── Actions ────────────────────────────────────────────
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
