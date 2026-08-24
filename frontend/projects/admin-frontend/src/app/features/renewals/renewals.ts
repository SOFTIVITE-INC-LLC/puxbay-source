import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { GrowthService } from '../../services/growth.service';

@Component({
  selector: 'app-renewals',
  standalone: true,
  imports: [CommonModule],
  template: `
<div class="p-8 max-w-7xl mx-auto">
  <div class="mb-8">
    <h1 class="text-3xl font-bold tracking-tight text-slate-900">Upcoming Renewals</h1>
    <p class="text-slate-500 mt-1">Active subscriptions renewing in the next 7 days.</p>
  </div>

  <div *ngIf="isLoading()" class="flex justify-center p-20">
    <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-indigo-600"></div>
  </div>

  <div *ngIf="!isLoading()" class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
    <table class="w-full text-left">
      <thead class="bg-slate-50 border-b border-slate-200">
        <tr>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Tenant</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Plan</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Renewal Date</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider text-right">Value</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-100">
        <tr *ngFor="let s of renewals()" class="hover:bg-slate-50">
          <td class="px-6 py-4 font-bold text-slate-900">{{ s.tenant?.name || 'Unknown' }}</td>
          <td class="px-6 py-4 text-sm text-slate-600">{{ s.plan?.name || '-' }}</td>
          <td class="px-6 py-4 text-sm font-bold text-amber-700">{{ s.next_billing_date | date:'mediumDate' }}</td>
          <td class="px-6 py-4 text-right font-bold text-emerald-600">\${{ s.plan?.price || 0 }}/mo</td>
        </tr>
        <tr *ngIf="renewals().length === 0">
          <td colspan="4" class="px-6 py-12 text-center text-slate-500">No renewals in the next 7 days.</td>
        </tr>
      </tbody>
    </table>
  </div>
</div>
  `,
})
export class RenewalsComponent implements OnInit {
  private service = inject(GrowthService);
  renewals = signal<any[]>([]);
  isLoading = signal(true);

  ngOnInit() {
    this.service.getUpcomingRenewals().subscribe({
      next: (res) => { this.renewals.set(res.data || []); this.isLoading.set(false); },
      error: () => this.isLoading.set(false)
    });
  }
}
