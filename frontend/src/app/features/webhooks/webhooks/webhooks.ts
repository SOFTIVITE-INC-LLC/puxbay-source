import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { WebhookService, Webhook, WEBHOOK_EVENTS } from '../../../core/services/webhook.service';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-webhooks',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './webhooks.html',
})
export class Webhooks implements OnInit {
  webhookService = inject(WebhookService);
  private toastr = inject(ToastrService);

  availableEvents = WEBHOOK_EVENTS;
  activeTab = signal<'endpoints' | 'logs'>('endpoints');
  isModalOpen = signal(false);
  selectedWebhook = signal<Webhook | null>(null);

  currentWebhook = signal<Partial<Webhook>>({
    is_active: true,
    event_types: []
  });

  ngOnInit() {
    this.webhookService.getWebhooks().subscribe();
  }

  openAddModal() {
    this.currentWebhook.set({ is_active: true, event_types: [], url: '' });
    this.isModalOpen.set(true);
  }

  openEditModal(wh: Webhook) {
    this.currentWebhook.set({ ...wh });
    this.isModalOpen.set(true);
  }

  toggleEvent(event: string) {
    const current = this.currentWebhook().event_types || [];
    const updated = current.includes(event) ? current.filter(e => e !== event) : [...current, event];
    this.currentWebhook.update(w => ({ ...w, event_types: updated }));
  }

  isEventSelected(event: string) {
    return (this.currentWebhook().event_types || []).includes(event);
  }

  saveWebhook() {
    const wh = this.currentWebhook();
    if (!wh.url || !wh.event_types?.length) {
      this.toastr.error('Please provide a URL and at least one event.');
      return;
    }
    const action = wh.id
      ? this.webhookService.updateWebhook(wh.id, wh)
      : this.webhookService.createWebhook(wh);
    action.subscribe({
      next: () => { this.toastr.success('Webhook saved!'); this.isModalOpen.set(false); },
      error: () => this.toastr.error('Failed to save webhook.')
    });
  }

  deleteWebhook(id: string) {
    this.webhookService.deleteWebhook(id).subscribe({
      next: () => this.toastr.success('Webhook deleted.'),
      error: () => this.toastr.error('Delete failed.')
    });
  }

  testWebhook(id: string) {
    this.webhookService.testWebhook(id).subscribe({
      next: () => this.toastr.success('Test payload sent!'),
      error: () => this.toastr.error('Test failed.')
    });
  }

  loadLogs(wh: Webhook) {
    this.selectedWebhook.set(wh);
    this.activeTab.set('logs');
    this.webhookService.getDeliveries(wh.id).subscribe();
  }

  retryDelivery(deliveryId: string) {
    if (!this.selectedWebhook()) return;
    this.webhookService.retryDelivery(this.selectedWebhook()!.id, deliveryId).subscribe({
      next: () => this.toastr.success('Retry queued.'),
      error: () => this.toastr.error('Retry failed.')
    });
  }
}
