import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { PrivacyService } from '../../../core/services/privacy.service';

@Component({
  selector: 'app-privacy',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './privacy.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .dark .glass-panel {
      background: rgba(0, 0, 0, 0.2);
    }
  `,
})
export class Privacy implements OnInit {
  privacyService = inject(PrivacyService);

  ngOnInit() {
    this.privacyService.getRequests().subscribe();
  }
}
