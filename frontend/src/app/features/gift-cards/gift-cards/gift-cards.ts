import { Component, inject, OnInit, signal } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GiftCardService, GiftCardCreateInput } from '../../../core/services/gift-card.service';
import { WalletService } from '../../../core/services/wallet.service';
import { CustomerService } from '../../../core/services/customer.service';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-gift-cards',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './gift-cards.html',
})
export class GiftCards implements OnInit {
  giftCardService = inject(GiftCardService);
  walletService = inject(WalletService);
  customerService = inject(CustomerService);
  private toastr = inject(ToastrService);

  isCreateModalOpen = signal(false);
  isBalanceCheckOpen = signal(false);
  isCreditModalOpen = signal(false);
  checkedCard = signal<any>(null);
  selectedCustomer = signal<any>(null);
  creditAdjustment = signal({ points: 0, amount: 0, description: '' });

  newCard = signal<GiftCardCreateInput>({ initial_balance: 50, purchaser_id: '' });
  balanceCheckCode = signal('');

  ngOnInit() {
    this.giftCardService.getCards().subscribe();
    this.customerService.getCustomers().subscribe();
  }

  createCard() {
    const c = this.newCard();
    if (!c.initial_balance || c.initial_balance <= 0) { this.toastr.error('Enter a valid amount.'); return; }
    this.giftCardService.createCard(c).subscribe({
      next: () => { this.toastr.success('Gift card created!'); this.isCreateModalOpen.set(false); },
      error: () => this.toastr.error('Failed to create gift card.')
    });
  }

  checkBalance() {
    const code = this.balanceCheckCode();
    if (!code) return;
    this.giftCardService.checkBalance(code).subscribe({
      next: (res) => this.checkedCard.set(res.gift_card),
      error: () => this.toastr.error('Card not found.')
    });
  }

  disableCard(id: string) {
    this.giftCardService.disableCard(id).subscribe({
      next: () => this.toastr.success('Card disabled.'),
      error: () => this.toastr.error('Failed to disable.')
    });
  }

  openCreditModal(customer: any) {
    this.selectedCustomer.set(customer);
    this.creditAdjustment.set({ points: 0, amount: 0, description: '' });
    this.isCreditModalOpen.set(true);
  }

  adjustLoyaltyPoints() {
    const c = this.selectedCustomer();
    if (!c) return;
    const adj = this.creditAdjustment();
    this.walletService.adjustLoyaltyPoints(c.id, adj.points, adj.description).subscribe({
      next: (res) => {
        this.toastr.success(`Points adjusted! New balance: ${(res as any).new_balance}`);
        this.isCreditModalOpen.set(false);
        this.customerService.getCustomers().subscribe();
      },
      error: () => this.toastr.error('Adjustment failed.')
    });
  }

  getCustomerName(id: string) {
    const c = this.customerService.customers().find(x => x.id === id);
    return c ? c.name : 'Unknown';
  }

  getBalancePercent(card: any) {
    if (!card.initial_balance) return 0;
    return (card.current_balance / card.initial_balance) * 100;
  }
}
