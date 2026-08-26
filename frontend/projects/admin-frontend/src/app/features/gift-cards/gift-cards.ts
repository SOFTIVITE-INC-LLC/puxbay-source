import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GiftCardService, GiftCard } from '../../services/gift-card.service';

@Component({
  selector: 'app-admin-gift-cards',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './gift-cards.html'
})
export class AdminGiftCardsComponent implements OnInit {
  private service = inject(GiftCardService);

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
      error: (err) => {
        console.error('Failed to create', err);
        this.isSaving.set(false);
        alert('Failed to issue gift card.');
      }
    });
  }

  disableCard(id: string) {
    if (confirm('Are you sure you want to disable this gift card?')) {
      this.service.disableGiftCard(id).subscribe({
        next: () => this.loadData(),
        error: () => alert('Failed to disable card.')
      });
    }
  }
}
