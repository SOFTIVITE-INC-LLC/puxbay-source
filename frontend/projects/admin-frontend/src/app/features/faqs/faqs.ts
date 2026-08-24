import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { FaqService, FAQ } from '../../services/faq.service';

@Component({
  selector: 'app-faqs',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './faqs.html',
})
export class FaqsComponent implements OnInit {
  private service = inject(FaqService);

  faqs = signal<FAQ[]>([]);
  isLoading = signal(true);
  isModalOpen = signal(false);
  isSaving = signal(false);

  form = signal<FAQ>({
    question: '',
    answer: '',
    order_index: 0
  });

  ngOnInit() {
    this.loadFaqs();
  }

  loadFaqs() {
    this.isLoading.set(true);
    this.service.getFAQs().subscribe({
      next: (data) => {
        this.faqs.set(data || []);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load FAQs', err);
        this.isLoading.set(false);
      }
    });
  }

  openModal() {
    this.form.set({ question: '', answer: '', order_index: this.faqs().length });
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  createFaq() {
    this.isSaving.set(true);
    this.service.createFAQ(this.form()).subscribe({
      next: () => {
        this.isSaving.set(false);
        this.closeModal();
        this.loadFaqs();
      },
      error: (err) => {
        console.error('Failed to create FAQ', err);
        this.isSaving.set(false);
      }
    });
  }

  toggleFaq(id: number) {
    this.service.toggleFAQ(id).subscribe({
      next: () => this.loadFaqs(),
      error: (err) => console.error('Failed to toggle FAQ', err)
    });
  }
}
