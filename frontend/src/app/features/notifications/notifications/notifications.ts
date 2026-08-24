import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { NotificationService, Notification } from '../../../core/services/notification.service';
import { ToastService } from '../../../core/services/toast';

type Tab = 'all' | 'alert' | 'success' | 'info';

interface NotificationGroup {
  label: string;
  items: Notification[];
}

@Component({
  selector: 'app-notifications',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './notifications.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .dark .glass-panel {
      background: rgba(0, 0, 0, 0.2);
    }
    .notif-item {
      transition: all 0.2s ease-in-out;
    }
    .notif-item:hover {
      transform: translateX(4px);
    }
  `,
})
export class Notifications implements OnInit {
  notificationService = inject(NotificationService);
  toastService = inject(ToastService);

  activeTab = signal<Tab>('all');
  selectedIds = signal<Set<string>>(new Set());
  
  // Computed property to filter and group notifications
  groupedNotifications = computed<NotificationGroup[]>(() => {
    let list = this.notificationService.notifications();
    
    // Filter by tab
    const tab = this.activeTab();
    if (tab !== 'all') {
      list = list.filter(n => {
        if (tab === 'alert') return n.type === 'alert' || n.type === 'warning' || n.type === 'error';
        if (tab === 'info') return n.type === 'info' || !n.type;
        return n.type === tab;
      });
    }

    // Group by Date
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    const groups: { [key: string]: Notification[] } = {
      'Today': [],
      'Yesterday': [],
      'Older': []
    };

    list.forEach(n => {
      const d = new Date(n.created_at);
      d.setHours(0, 0, 0, 0);
      
      if (d.getTime() === today.getTime()) {
        groups['Today'].push(n);
      } else if (d.getTime() === yesterday.getTime()) {
        groups['Yesterday'].push(n);
      } else {
        groups['Older'].push(n);
      }
    });

    return [
      { label: 'Today', items: groups['Today'] },
      { label: 'Yesterday', items: groups['Yesterday'] },
      { label: 'Older', items: groups['Older'] }
    ].filter(g => g.items.length > 0);
  });

  isAllSelected = computed(() => {
    const list = this.filteredList();
    return list.length > 0 && this.selectedIds().size === list.length;
  });

  filteredList = computed(() => {
    return this.groupedNotifications().flatMap(g => g.items);
  });

  ngOnInit() {
    this.notificationService.getList().subscribe();
  }

  setTab(tab: Tab) {
    this.activeTab.set(tab);
    this.selectedIds.set(new Set()); // clear selection on tab change
  }

  toggleSelection(id: string) {
    const set = new Set(this.selectedIds());
    if (set.has(id)) {
      set.delete(id);
    } else {
      set.add(id);
    }
    this.selectedIds.set(set);
  }

  toggleAll() {
    if (this.isAllSelected()) {
      this.selectedIds.set(new Set());
    } else {
      const allIds = this.filteredList().map(n => n.id);
      this.selectedIds.set(new Set(allIds));
    }
  }

  markRead(id: string) {
    this.notificationService.markAsRead(id).subscribe();
  }

  markAllAsRead() {
    this.notificationService.markAllAsRead().subscribe(() => {
      this.toastService.showSuccess('All notifications marked as read');
    });
  }

  markSelectedAsRead() {
    const ids = Array.from(this.selectedIds());
    if (ids.length === 0) return;
    
    // In a real app, there would be a bulk endpoint. Here we'll simulate by calling individual.
    let completed = 0;
    ids.forEach(id => {
      this.notificationService.markAsRead(id).subscribe(() => {
        completed++;
        if (completed === ids.length) {
          this.toastService.showSuccess(`${ids.length} notifications marked as read`);
          this.selectedIds.set(new Set());
        }
      });
    });
  }

  deleteSelected() {
    const ids = Array.from(this.selectedIds());
    if (ids.length === 0) return;
    
    let completed = 0;
    ids.forEach(id => {
      this.notificationService.deleteNotification(id).subscribe(() => {
        completed++;
        if (completed === ids.length) {
          this.toastService.showSuccess(`${ids.length} notifications deleted`);
          this.selectedIds.set(new Set());
        }
      });
    });
  }

  testToast() {
    const types = ['success', 'alert', 'info', 'warning'];
    const type = types[Math.floor(Math.random() * types.length)];
    
    if (type === 'success') {
      this.toastService.showSuccess('This is a test success notification!');
    } else if (type === 'alert' || type === 'warning') {
      this.toastService.showError('Warning: Inventory levels are low for SKU-1029.');
    } else {
      this.toastService.showSuccess('System update completed successfully.'); // Assuming only success/error exist in toastService
    }
  }
}
