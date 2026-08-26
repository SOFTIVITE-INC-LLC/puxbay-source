import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SubscriptionService, Subscription } from '../../services/subscription.service';
import { AlertService } from '../../services/alert.service';

@Component({
  selector: 'app-subscriptions',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './subscriptions.html',
})
export class SubscriptionsComponent implements OnInit {
  private service = inject(SubscriptionService);
  private alert = inject(AlertService);

  subscriptions = signal<Subscription[]>([]);
  stats = signal<any>(null);
  isLoading = signal(true);
  
  // Filtering
  filter = signal<string>('all');
  
  // Dropdown state
  openDropdownId = signal<string | null>(null);

  ngOnInit() {
    this.loadSubscriptions();
    
    document.addEventListener('click', (e) => {
      if (!(e.target as Element).closest('.action-menu-container')) {
        this.openDropdownId.set(null);
      }
    });
  }

  toggleDropdown(id: string, event: Event) {
    event.stopPropagation();
    if (this.openDropdownId() === id) {
      this.openDropdownId.set(null);
    } else {
      this.openDropdownId.set(id);
    }
  }

  setFilter(status: string) {
    this.filter.set(status);
  }

  get filteredSubscriptions() {
    const f = this.filter();
    if (f === 'all') return this.subscriptions();
    return this.subscriptions().filter(s => s.status === f);
  }

  loadSubscriptions() {
    this.isLoading.set(true);
    this.service.getSubscriptions().subscribe({
      next: (res) => {
        this.subscriptions.set(res.data || []);
        this.stats.set(res.stats || null);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load subscriptions', err);
        this.isLoading.set(false);
      }
    });
  }

  async overrideStatus(id: string, newStatus: string) {
    this.openDropdownId.set(null);
    const confirmed = await this.alert.confirm({
      title: 'Override Subscription Status',
      message: `Manually change this subscription to '${newStatus}'? This will override the current billing status immediately.`,
      confirmText: 'Apply Override',
      cancelText: 'Cancel',
      type: 'warning'
    });
    if (confirmed) {
      this.service.overrideSubscription(id, newStatus).subscribe({
        next: () => {
          this.alert.success(`Subscription status changed to '${newStatus}'.`);
          this.loadSubscriptions();
        },
        error: () => this.alert.error('Failed to override subscription status.')
      });
    }
  }
}
