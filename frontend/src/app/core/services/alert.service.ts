import { Injectable, signal } from '@angular/core';

export interface AlertState {
  isOpen: boolean;
  type: 'alert' | 'confirm';
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
    title: '',
    message: '',
    confirmText: 'OK',
    cancelText: 'Cancel'
  });

  alert(message: string, title: string = 'Alert'): Promise<boolean> {
    return new Promise((resolve) => {
      this.state.set({
        isOpen: true,
        type: 'alert',
        title,
        message,
        confirmText: 'OK',
        cancelText: '',
        resolve
      });
    });
  }

  confirm(message: string, title: string = 'Confirm'): Promise<boolean> {
    return new Promise((resolve) => {
      this.state.set({
        isOpen: true,
        type: 'confirm',
        title,
        message,
        confirmText: 'Confirm',
        cancelText: 'Cancel',
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
