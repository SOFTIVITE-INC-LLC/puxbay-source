import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { PaymentService, Payment } from '../../services/payment.service';

@Component({
  selector: 'app-payments',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './payments.html',
})
export class PaymentsComponent implements OnInit {
  private service = inject(PaymentService);

  payments = signal<Payment[]>([]);
  stats = signal<any>(null);
  isLoading = signal(true);
  
  // Filtering
  filter = signal<string>('all');
  
  // Dropdown state
  openDropdownId = signal<string | null>(null);

  ngOnInit() {
    this.loadPayments();
    
    document.addEventListener('click', (e) => {
      if (!(e.target as Element).closest('.action-menu-container')) {
        this.openDropdownId.set(null);
      }
    });
  }

  toggleDropdown(id: string, event: Event) {
    event.stopPropagation();
    if (this.openDropdownId() === id) {
      this.openDropdownId.set(null);
    } else {
      this.openDropdownId.set(id);
    }
  }

  setFilter(status: string) {
    this.filter.set(status);
  }

  get filteredPayments() {
    const f = this.filter();
    if (f === 'all') return this.payments();
    // Adjust logic to map 'succeeded' tab to 'succeeded' or 'paid'
    if (f === 'succeeded') return this.payments().filter(p => p.status === 'succeeded' || p.status === 'paid');
    return this.payments().filter(p => p.status === f);
  }

  loadPayments() {
    this.isLoading.set(true);
    this.service.getPayments().subscribe({
      next: (res) => {
        this.payments.set(res.data || []);
        this.stats.set(res.stats || null);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load payments', err);
        this.isLoading.set(false);
      }
    });
  }
}
