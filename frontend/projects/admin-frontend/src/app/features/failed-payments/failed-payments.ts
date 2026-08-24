import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { GrowthService } from '../../services/growth.service';

@Component({
  selector: 'app-failed-payments',
  standalone: true,
  imports: [CommonModule],
  template: `
<div class="p-8 max-w-7xl mx-auto">
  <div class="mb-8">
    <h1 class="text-3xl font-bold tracking-tight text-slate-900">Failed Payments</h1>
    <p class="text-slate-500 mt-1">Past-due and failed payment records requiring action.</p>
  </div>

  <div *ngIf="isLoading()" class="flex justify-center p-20">
    <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-indigo-600"></div>
  </div>

  <div *ngIf="!isLoading()" class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
    <table class="w-full text-left">
      <thead class="bg-slate-50 border-b border-slate-200">
        <tr>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Tenant</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Reference</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Amount</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Status</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Date</th>
          <th class="px-6 py-4 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-100">
        <tr *ngFor="let p of payments()" class="hover:bg-slate-50">
          <td class="px-6 py-4 font-bold text-slate-900">{{ p.subscription?.tenant?.name || 'Unknown' }}</td>
          <td class="px-6 py-4 font-mono text-xs text-slate-500">{{ p.paystack_reference || '-' }}</td>
          <td class="px-6 py-4 font-bold text-rose-600">\${{ p.amount }}</td>
          <td class="px-6 py-4">
            <span class="px-2 py-1 text-xs font-bold bg-rose-100 text-rose-700 rounded-full">{{ p.status | titlecase }}</span>
          </td>
          <td class="px-6 py-4 text-sm text-slate-500">{{ p.created_at | date:'mediumDate' }}</td>
          <td class="px-6 py-4 text-right">
            <button class="px-3 py-1.5 text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg cursor-pointer transition">Retry</button>
          </td>
        </tr>
        <tr *ngIf="payments().length === 0">
          <td colspan="6" class="px-6 py-12 text-center">
            <span class="material-symbols-outlined text-4xl text-slate-300 block mb-2">check_circle</span>
            <p class="text-slate-500">No failed payments. All clear!</p>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</div>
  `,
})
export class FailedPaymentsComponent implements OnInit {
  private service = inject(GrowthService);
  payments = signal<any[]>([]);
  isLoading = signal(true);

  ngOnInit() {
    this.service.getFailedPayments().subscribe({
      next: (res) => { this.payments.set(res.data || []); this.isLoading.set(false); },
      error: () => this.isLoading.set(false)
    });
  }
}
