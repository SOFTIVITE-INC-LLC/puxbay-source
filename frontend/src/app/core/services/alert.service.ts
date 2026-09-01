import { Injectable, signal } from '@angular/core';

export interface AlertState {
  isOpen: boolean;
  type: 'alert' | 'confirm';
  variant: 'danger' | 'warning' | 'info' | 'success';
  title: string;
  message: string;
  confirmText: string;
  cancelText: string;
  resolve?: (value: boolean) => void;
}

@Injectable({
  providedIn: 'root'
})
export class AlertService {
  readonly state = signal<AlertState>({
    isOpen: false,
    type: 'alert',
    variant: 'info',
    title: '',
    message: '',
    confirmText: 'OK',
    cancelText: 'Cancel'
  });

  alert(message: string, title: string = 'Notice', variant: 'info' | 'success' | 'warning' | 'danger' = 'info'): Promise<boolean> {
    return new Promise((resolve) => {
      this.state.set({
        isOpen: true,
        type: 'alert',
        variant,
        title,
        message,
        confirmText: 'Got it',
        cancelText: '',
        resolve
      });
    });
  }

  confirm(
    message: string,
    title: string = 'Confirm Action',
    confirmText: string = 'Confirm',
    cancelText: string = 'Cancel',
    variant?: 'danger' | 'warning' | 'info' | 'success'
  ): Promise<boolean> {
    // Auto-detect destructive actions if variant not explicitly provided
    let computedVariant: 'danger' | 'warning' | 'info' | 'success' = variant || 'warning';
    if (!variant) {
      const lower = (title + ' ' + message).toLowerCase();
      if (lower.includes('delete') || lower.includes('remove') || lower.includes('disable') || lower.includes('archive') || lower.includes('void')) {
        computedVariant = 'danger';
        if (confirmText === 'Confirm') confirmText = 'Delete';
      }
    }

    return new Promise((resolve) => {
      this.state.set({
        isOpen: true,
        type: 'confirm',
        variant: computedVariant,
        title,
        message,
        confirmText,
        cancelText,
        resolve
      });
    });
  }

  close(result: boolean) {
    const currentState = this.state();
    if (currentState.resolve) {
      currentState.resolve(result);
    }
    this.state.update(s => ({ ...s, isOpen: false }));
  }
}
