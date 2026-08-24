import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SubscriptionService, Subscription } from '../../services/subscription.service';

@Component({
  selector: 'app-subscriptions',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './subscriptions.html',
})
export class SubscriptionsComponent implements OnInit {
  private service = inject(SubscriptionService);

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

  overrideStatus(id: string, newStatus: string) {
    this.openDropdownId.set(null);
    if (confirm(`Are you sure you want to manually change this subscription to '${newStatus}'?`)) {
      this.service.overrideSubscription(id, newStatus).subscribe({
        next: () => {
          this.loadSubscriptions();
        },
        error: (err) => console.error('Failed to override subscription', err)
      });
    }
  }
}
