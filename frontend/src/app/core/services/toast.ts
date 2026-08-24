import { Injectable, signal } from '@angular/core';

export interface ToastMessage {
  id: number;
  message: string;
  type: 'success' | 'error' | 'warning' | 'info';
}

@Injectable({
  providedIn: 'root'
})
export class ToastService {
  readonly toasts = signal<ToastMessage[]>([]);
  private nextId = 0;

  showSuccess(message: string) { this.addToast(message, 'success'); }
  showError(message: string) { this.addToast(message, 'error'); }
  showWarning(message: string) { this.addToast(message, 'warning'); }
  showInfo(message: string) { this.addToast(message, 'info'); }

  private addToast(message: string, type: ToastMessage['type']) {
    const id = this.nextId++;
    const toast = { id, message, type };
    this.toasts.update(t => [...t, toast]);
    setTimeout(() => this.removeToast(id), 3000);
  }

  removeToast(id: number) {
    this.toasts.update(t => t.filter(toast => toast.id !== id));
  }
}
