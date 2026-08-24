import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AlertService } from '../../services/alert.service';

@Component({
  selector: 'app-alert',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './alert.html',
  styles: `
    @keyframes alertIn {
      from { opacity: 0; transform: scale(0.95) translateY(10px); }
      to { opacity: 1; transform: scale(1) translateY(0); }
    }
    .alert-dialog {
      animation: alertIn 0.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
    }
  `
})
export class AlertComponent {
  alertService = inject(AlertService);

  onConfirm() {
    this.alertService.close(true);
  }

  onCancel() {
    this.alertService.close(false);
  }
}
