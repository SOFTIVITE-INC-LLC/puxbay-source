import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { PromoCodeService, PromoCode } from '../../services/promo-code.service';

@Component({
  selector: 'app-promo-codes',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './promo-codes.html',
})
export class PromoCodesComponent implements OnInit {
  private service = inject(PromoCodeService);

  codes = signal<PromoCode[]>([]);
  stats = signal<any>(null);
  isLoading = signal(true);
  isModalOpen = signal(false);
  isSaving = signal(false);

  form = signal<PromoCode>({
    code: '',
    discount_type: 'percentage',
    discount_value: 0,
    max_uses: 0,
    valid_until: null
  });

  ngOnInit() {
    this.loadCodes();
  }

  loadCodes() {
    this.isLoading.set(true);
    this.service.getPromoCodes().subscribe({
      next: (res) => {
        this.codes.set(res.data || []);
        this.stats.set(res.stats || null);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load promo codes', err);
        this.isLoading.set(false);
      }
    });
  }

  openModal() {
    this.form.set({
      code: '',
      discount_type: 'percentage',
      discount_value: 0,
      max_uses: 0,
      valid_until: null
    });
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  createCode() {
    this.isSaving.set(true);
    // Format the date if it's set to empty string
    const payload = { ...this.form() };
    
    // Ensure numerical types
    payload.discount_value = Number(payload.discount_value);
    payload.max_uses = Number(payload.max_uses);

    if (!payload.valid_until) {
      delete payload.valid_until;
    } else {
      payload.valid_until = new Date(payload.valid_until).toISOString();
    }
    
    this.service.createPromoCode(payload).subscribe({
      next: () => {
        this.isSaving.set(false);
        this.closeModal();
        this.loadCodes();
      },
      error: (err) => {
        console.error('Failed to create code', err);
        this.isSaving.set(false);
      }
    });
  }

  toggleCode(id: string) {
    this.service.togglePromoCode(id).subscribe({
      next: () => this.loadCodes(),
      error: (err) => console.error('Failed to toggle code', err)
    });
  }

  copyToClipboard(code: string) {
    navigator.clipboard.writeText(code).then(() => {
      // Optional: Add a toast notification here
      console.log('Copied to clipboard: ', code);
    }).catch(err => {
      console.error('Could not copy text: ', err);
    });
  }
  
  isExpired(validUntil: string | null | undefined): boolean {
    if (!validUntil) return false;
    return new Date(validUntil) < new Date();
  }
}
