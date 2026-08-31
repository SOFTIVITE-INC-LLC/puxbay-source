import { Component, inject, signal, OnInit } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { WalletService } from '../../../core/services/wallet.service';
import { CustomerService } from '../../../core/services/customer.service';
import { ToastrService } from 'ngx-toastr';
import { SettingsService } from '../../../core/services/settings.service';

@Component({
  selector: 'app-wallet',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './wallet.html',
  styles: `
    .card-gradient { background-color: #011f4b; }
    .glass {
      background: rgba(255,255,255,0.7);
      backdrop-filter: blur(20px);
      -webkit-backdrop-filter: blur(20px);
      border: 1px solid rgba(255,255,255,0.5);
    }
    :host-context(.dark) .glass {
      background: rgba(15,15,25,0.7);
      border: 1px solid rgba(255,255,255,0.08);
    }
  `
})
export class Wallet implements OnInit {
  walletSvc = inject(WalletService);
  customerSvc = inject(CustomerService);
  private toastr = inject(ToastrService);
  settingsService = inject(SettingsService);

  activeTab = signal<'dashboard' | 'lookup' | 'adjustments'>('dashboard');
  selectedCustomerId = signal<string | null>(null);
  phoneInput = signal('');

  // Adjustment modals
  isLoyaltyModalOpen = signal(false);
  isCreditModalOpen = signal(false);
  isGiftCardModalOpen = signal(false);
  loyaltyAdjust = signal({ points: 0, note: '' });
  creditAdjust = signal({ amount: 0, note: '' });
  giftCardAmount = signal(0);
  Math = Math;

  ngOnInit() {
    this.customerSvc.getCustomers().subscribe();
  }

  lookupCustomer() {
    const phone = this.phoneInput().trim();
    if (!phone) return;
    this.walletSvc.lookupByPhone(phone).subscribe({
      next: (res) => {
        this.selectedCustomerId.set(res.customer_id);
        this.loadDashboard(res.customer_id);
        this.activeTab.set('dashboard');
      },
      error: () => {}
    });
  }

  selectCustomer(id: string) {
    this.selectedCustomerId.set(id);
    this.loadDashboard(id);
    this.activeTab.set('dashboard');
  }

  loadDashboard(customerId: string) {
    this.walletSvc.loadDashboard(customerId).subscribe({
      error: () => this.toastr.error('Failed to load wallet data.')
    });
  }

  adjustLoyalty() {
    const id = this.selectedCustomerId();
    if (!id) return;
    const { points, note } = this.loyaltyAdjust();
    this.walletSvc.adjustLoyaltyPoints(id, points, note).subscribe({
      next: () => {
        this.toastr.success('Loyalty points adjusted!');
        this.isLoyaltyModalOpen.set(false);
        this.loadDashboard(id);
      },
      error: () => this.toastr.error('Failed to adjust points.')
    });
  }

  adjustCredit() {
    const id = this.selectedCustomerId();
    if (!id) return;
    const { amount, note } = this.creditAdjust();
    this.walletSvc.adjustStoreCredit(id, amount, note).subscribe({
      next: () => {
        this.toastr.success('Store credit updated!');
        this.isCreditModalOpen.set(false);
        this.loadDashboard(id);
      },
      error: () => this.toastr.error('Failed to adjust store credit.')
    });
  }

  issueGiftCard() {
    const id = this.selectedCustomerId();
    const amount = this.giftCardAmount();
    if (!id || amount <= 0) return;
    this.walletSvc.issueGiftCard(id, amount).subscribe({
      next: () => {
        this.toastr.success('Gift card issued!');
        this.isGiftCardModalOpen.set(false);
        this.loadDashboard(id);
      },
      error: () => this.toastr.error('Failed to issue gift card.')
    });
  }

  loyaltyTypeIcon(type: string): string {
    return { earn: 'add_circle', redeem: 'redeem', adjust: 'tune', expire: 'schedule' }[type] ?? 'stars';
  }

  loyaltyTypeColor(type: string): string {
    return {
      earn: 'text-emerald-500',
      redeem: 'text-amber-500',
      adjust: 'text-blue-500',
      expire: 'text-rose-500'
    }[type] ?? 'text-slate-500';
  }

  formatDate(d: string): string {
    return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }
}
