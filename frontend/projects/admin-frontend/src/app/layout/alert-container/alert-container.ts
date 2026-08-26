import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AlertService } from '../../services/alert.service';

@Component({
  selector: 'app-alert-container',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './alert-container.html'
})
export class AlertContainerComponent {
  alert = inject(AlertService);
}
