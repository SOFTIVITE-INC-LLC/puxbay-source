import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CrmService } from '../../../core/services/crm.service';
import { ToastService } from '../../../core/services/toast';
import { CustomerService } from '../../../core/services/customer.service';

@Component({
  selector: 'app-helpdesk',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './helpdesk.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.6);
      backdrop-filter: blur(16px);
      border: 1px solid rgba(255, 255, 255, 0.3);
    }
    .dark .glass-panel {
      background: rgba(24, 24, 27, 0.6);
      border: 1px solid rgba(255, 255, 255, 0.05);
    }
  `
})
export class Helpdesk implements OnInit {
  crm = inject(CrmService);
  customerService = inject(CustomerService);
  toast = inject(ToastService);

  isModalOpen = signal(false);
  activeTicket = signal<any>(null);
  ticketMessages = signal<any[]>([]);
  replyText = signal('');

  newTicket = signal({
    customer_id: '',
    subject: '',
    description: '',
    priority: 'medium'
  });

  ngOnInit() {
    this.crm.getTickets().subscribe();
    this.customerService.getCustomers().subscribe();
  }

  openTicketModal() {
    this.newTicket.set({ customer_id: '', subject: '', description: '', priority: 'medium' });
    this.isModalOpen.set(true);
  }

  saveTicket() {
    this.crm.createTicket(this.newTicket()).subscribe({
      next: () => {
        this.toast.showSuccess('Ticket created successfully');
        this.isModalOpen.set(false);
      },
      error: () => this.toast.showError('Failed to create ticket')
    });
  }

  viewTicket(t: any) {
    this.activeTicket.set(t);
    this.crm.getTicketMessages(t.id).subscribe(msgs => this.ticketMessages.set(msgs));
  }

  sendReply() {
    if (!this.replyText().trim()) return;
    this.crm.replyTicket(this.activeTicket().id, this.replyText()).subscribe({
      next: (msg) => {
        this.ticketMessages.update(list => [...list, msg]);
        this.replyText.set('');
      },
      error: () => this.toast.showError('Failed to send reply')
    });
  }
}
