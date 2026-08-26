import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GiftCardService, GiftCard } from '../../services/gift-card.service';
import { AlertService } from '../../services/alert.service';

@Component({
  selector: 'app-admin-gift-cards',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './gift-cards.html'
})
export class AdminGiftCardsComponent implements OnInit {
  private service = inject(GiftCardService);
  private alert = inject(AlertService);

  cards = signal<GiftCard[]>([]);
  isLoading = signal(true);
  isSaving = signal(false);
  isModalOpen = signal(false);
  
  newCard = signal({
    initial_balance: 50,
    custom_code: '',
    expires_at: ''
  });

  ngOnInit() {
    this.loadData();
  }

  loadData() {
    this.isLoading.set(true);
    this.service.getGiftCards().subscribe({
      next: (res) => {
        this.cards.set(res.data || []);
        this.isLoading.set(false);
      },
      error: () => this.isLoading.set(false)
    });
  }

  openCreateModal() {
    this.newCard.set({ initial_balance: 50, custom_code: '', expires_at: '' });
    this.isModalOpen.set(true);
  }

  closeCreateModal() {
    this.isModalOpen.set(false);
  }

  createCard() {
    this.isSaving.set(true);
    this.service.createGiftCard(this.newCard()).subscribe({
      next: () => {
        this.isSaving.set(false);
        this.closeCreateModal();
        this.loadData();
      },
      error: () => {
        this.isSaving.set(false);
        this.alert.error('Failed to issue gift card. Please check the details and try again.', 'Error');
      }
    });
  }

  async disableCard(id: string) {
    const confirmed = await this.alert.confirm({
      title: 'Disable Gift Card',
      message: 'Are you sure you want to disable this gift card? It will no longer be redeemable.',
      confirmText: 'Disable',
      cancelText: 'Cancel',
      type: 'warning'
    });
    if (confirmed) {
      this.service.disableGiftCard(id).subscribe({
        next: () => { this.alert.success('Gift card disabled.'); this.loadData(); },
        error: () => this.alert.error('Failed to disable gift card.')
      });
    }
  }
}
