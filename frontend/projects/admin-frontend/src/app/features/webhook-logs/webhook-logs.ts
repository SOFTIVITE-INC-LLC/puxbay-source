import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { GrowthService, WebhookEvent } from '../../services/growth.service';
import { AlertService } from '../../services/alert.service';

@Component({
  selector: 'app-webhook-logs',
  standalone: true,
  imports: [CommonModule],
  template: `
<div class="p-8 max-w-7xl mx-auto">
  <div class="mb-8">
    <h1 class="text-3xl font-bold tracking-tight text-slate-900">Webhook Logs</h1>
    <p class="text-slate-500 mt-1">Review outbound webhook event delivery attempts.</p>
  </div>

  <div *ngIf="isLoading()" class="flex justify-center p-20">
    <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-indigo-600"></div>
  </div>

  <div *ngIf="!isLoading()" class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
    <table class="w-full text-left">
      <thead class="bg-slate-50 border-b border-slate-200">
        <tr>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Event</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Status</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Response</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Attempts</th>
          <th class="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Date</th>
          <th class="px-6 py-4 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-100">
        <tr *ngFor="let e of events()" class="hover:bg-slate-50">
          <td class="px-6 py-4 font-mono font-bold text-slate-900">{{ e.event_type }}</td>
          <td class="px-6 py-4">
            <span class="px-2 py-1 text-xs font-bold rounded-full" 
                  [ngClass]="e.status === 'success' ? 'bg-emerald-100 text-emerald-700' : 
                             e.status === 'pending' ? 'bg-amber-100 text-amber-700' : 'bg-rose-100 text-rose-700'">
              {{ e.status | titlecase }}
            </span>
          </td>
          <td class="px-6 py-4 text-sm font-mono text-slate-600">{{ e.response_code || '-' }}</td>
          <td class="px-6 py-4 text-sm text-slate-600">{{ e.attempts }}</td>
          <td class="px-6 py-4 text-sm text-slate-500">{{ e.created_at | date:'medium' }}</td>
          <td class="px-6 py-4 text-right">
            <button (click)="retry(e.id)" 
                    [disabled]="isRetrying(e.id) || e.status === 'success' || e.status === 'pending'"
                    class="px-3 py-1.5 text-xs font-bold text-slate-700 bg-slate-100 hover:bg-slate-200 disabled:opacity-50 rounded-lg cursor-pointer transition">
              {{ isRetrying(e.id) ? 'Retrying...' : 'Retry' }}
            </button>
          </td>
        </tr>
        <tr *ngIf="events().length === 0">
          <td colspan="6" class="px-6 py-12 text-center text-slate-500">No webhook events logged.</td>
        </tr>
      </tbody>
    </table>
  </div>
</div>
  `,
})
export class WebhookLogsComponent implements OnInit {
  private service = inject(GrowthService);
  private alert = inject(AlertService);
  events = signal<WebhookEvent[]>([]);
  isLoading = signal(true);
  retryingIds = signal<Set<string>>(new Set());

  ngOnInit() {
    this.loadEvents();
  }

  loadEvents() {
    this.service.getWebhookEvents().subscribe({
      next: (res) => { this.events.set(res.data || []); this.isLoading.set(false); },
      error: () => this.isLoading.set(false)
    });
  }

  isRetrying(id: string): boolean {
    return this.retryingIds().has(id);
  }

  retry(id: string) {
    const current = this.retryingIds();
    current.add(id);
    this.retryingIds.set(new Set(current));

    this.service.retryWebhookEvent(id).subscribe({
      next: () => {
        this.loadEvents(); // Reload to show pending status
        const next = this.retryingIds();
        next.delete(id);
        this.retryingIds.set(new Set(next));
      },
      error: () => {
        const next = this.retryingIds();
        next.delete(id);
        this.retryingIds.set(new Set(next));
        this.alert.error('Failed to queue retry. Please try again.', 'Retry Failed');
      }
    });
  }
}
