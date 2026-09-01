import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { NotificationSoundService } from './notification-sound.service';

export interface Notification {
  id: string;
  user_id: string;
  type: string;
  title: string;
  message: string;
  is_read: boolean;
  link?: string;
  created_at: string;
}

export interface NotificationSetting {
  user_id: string;
  email_notifications: boolean;
  low_stock_alerts: boolean;
  sales_reports: boolean;
  security_alerts: boolean;
  system_alerts: boolean;
}

export interface NotificationListResult {
  notifications: Notification[];
  total: number;
  unread_count: number;
  page: number;
  page_size: number;
}

@Injectable({
  providedIn: 'root'
})
export class NotificationService {
  private api = inject(ApiService);
  
  notifications = signal<Notification[]>([]);
  latestNotifications = signal<Notification[]>([]);
  unreadCount = signal<number>(0);
  settings = signal<NotificationSetting | null>(null);
  loading = signal<boolean>(false);

  private previousCount = 0;
  private soundService = inject(NotificationSoundService);

  getList(params?: any): Observable<NotificationListResult> {
    this.loading.set(true);
    return this.api.get<NotificationListResult>('/notifications', { params }).pipe(
      tap(res => {
        this.notifications.set(res.notifications || []);
        this.unreadCount.set(res.unread_count || 0);
        this.loading.set(false);
      })
    );
  }

  getLatest(): Observable<{count: number, notifications: Notification[]}> {
    return this.api.get<{count: number, notifications: Notification[]}>('/notifications/latest').pipe(
      tap(res => {
        const prev = this.previousCount;
        const current = res.count || 0;
        this.previousCount = current;
        this.latestNotifications.set(res.notifications || []);
        this.unreadCount.set(current);

        // If new notifications arrived via polling fallback and count grew
        if (current > prev && prev > 0 && res.notifications?.length > 0) {
          const first = res.notifications[0];
          this.soundService.play((first as any).category || first.type || 'general');
        }
      })
    );
  }

  markAsRead(id: string): Observable<any> {
    return this.api.post(`/notifications/${id}/read`, {}).pipe(
      tap(() => {
        this.notifications.update(list => list.map(n => n.id === id ? { ...n, is_read: true } : n));
        this.latestNotifications.update(list => list.map(n => n.id === id ? { ...n, is_read: true } : n));
        this.unreadCount.update(c => Math.max(0, c - 1));
      })
    );
  }

  markAllAsRead(): Observable<any> {
    return this.api.post('/notifications/read-all', {}).pipe(
      tap(() => {
        this.notifications.update(list => list.map(n => ({ ...n, is_read: true })));
        this.latestNotifications.update(list => list.map(n => ({ ...n, is_read: true })));
        this.unreadCount.set(0);
      })
    );
  }

  getSettings(): Observable<NotificationSetting> {
    return this.api.get<NotificationSetting>('/notifications/settings').pipe(
      tap(res => this.settings.set(res))
    );
  }

  updateSettings(settings: Partial<NotificationSetting>): Observable<NotificationSetting> {
    return this.api.put<NotificationSetting>('/notifications/settings', settings).pipe(
      tap(res => this.settings.set(res))
    );
  }

  deleteNotification(id: string): Observable<any> {
    return this.api.delete(`/notifications/${id}`).pipe(
      tap(() => {
        const notif = this.notifications().find(n => n.id === id);
        if (notif && !notif.is_read) {
          this.unreadCount.update(c => Math.max(0, c - 1));
        }
        this.notifications.update(list => list.filter(n => n.id !== id));
        this.latestNotifications.update(list => list.filter(n => n.id !== id));
      })
    );
  }
}
