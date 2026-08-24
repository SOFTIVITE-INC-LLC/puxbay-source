import { Component, inject, OnInit, signal } from '@angular/core';

import { FormsModule } from '@angular/forms';
import { ServicesService } from '../../../core/services/services.service';

@Component({
  selector: 'app-services',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './services.html',
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
export class Services implements OnInit {
  servicesService = inject(ServicesService);

  isModalOpen = signal(false);
  modalTitle = signal('Add Service');

  currentService = signal<any>({
    duration_minutes: 30,
    price: 0,
    is_active: true,
  });

  ngOnInit() {
    this.servicesService.getServices().subscribe();
  }

  openAddModal() {
    this.modalTitle.set('Add Service');
    this.currentService.set({
      duration_minutes: 30,
      price: 0,
      is_active: true,
    });
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  saveService() {
    const s = this.currentService();
    this.servicesService.createService(s).subscribe(() => this.closeModal());
  }
}
