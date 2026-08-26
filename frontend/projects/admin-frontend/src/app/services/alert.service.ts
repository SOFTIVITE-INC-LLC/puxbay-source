import { Injectable, signal } from '@angular/core';

export type AlertType = 'success' | 'error' | 'warning' | 'info';

export interface ToastMessage {
  id: string;
  type: AlertType;
  title?: string;
  message: string;
  duration?: number;
}

export interface ConfirmOptions {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  type?: 'danger' | 'warning' | 'info';
}

@Injectable({
  providedIn: 'root'
})
export class AlertService {
  toasts = signal<ToastMessage[]>([]);
  confirmDialog = signal<{
    options: ConfirmOptions;
    resolve: (value: boolean) => void;
  } | null>(null);

  /**
   * Show a success toast notification
   */
  success(message: string, title?: string, duration: number = 4000) {
    this.showToast('success', message, title, duration);
  }

  /**
   * Show an error toast notification
   */
  error(message: string, title?: string, duration: number = 5000) {
    this.showToast('error', message, title, duration);
  }

  /**
   * Show a warning toast notification
   */
  warning(message: string, title?: string, duration: number = 4500) {
    this.showToast('warning', message, title, duration);
  }

  /**
   * Show an info toast notification
   */
  info(message: string, title?: string, duration: number = 4000) {
    this.showToast('info', message, title, duration);
  }

  private showToast(type: AlertType, message: string, title?: string, duration: number = 4000) {
    const id = Math.random().toString(36).substring(2, 9);
    const toast: ToastMessage = { id, type, message, title, duration };

    this.toasts.update(current => [...current, toast]);

    if (duration > 0) {
      setTimeout(() => {
        this.removeToast(id);
      }, duration);
    }
  }

  removeToast(id: string) {
    this.toasts.update(current => current.filter(t => t.id !== id));
  }

  /**
   * Show a styled confirmation modal and return a Promise<boolean>
   */
  confirm(options: ConfirmOptions): Promise<boolean> {
    return new Promise<boolean>((resolve) => {
      this.confirmDialog.set({
        options: {
          confirmText: 'Confirm',
          cancelText: 'Cancel',
          type: 'danger',
          ...options
        },
        resolve: (result: boolean) => {
          this.confirmDialog.set(null);
          resolve(result);
        }
      });
    });
  }

  handleConfirm(result: boolean) {
    const dialog = this.confirmDialog();
    if (dialog) {
      dialog.resolve(result);
    }
  }
}
