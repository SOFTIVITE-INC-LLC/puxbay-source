import { ToastService } from '../../../core/services/toast';
import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CrmService, CustomerFeedback } from '../../../core/services/crm.service';
import { CustomerService } from '../../../core/services/customer.service';
import { AlertService } from '../../../core/services/alert.service';

@Component({
  selector: 'app-feedback',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './feedback.html',
  styleUrl: './feedback.css',
})
export class Feedback implements OnInit {
  toastService = inject(ToastService);
  private crmService = inject(CrmService);
  private customerService = inject(CustomerService);
  private alertService = inject(AlertService);

  feedbackList = signal<CustomerFeedback[]>([]);
  loading = signal(false);

  // Customer lookup map (id -> name)
  customerNames = signal<Record<string, string>>({});
  
  // List of customers for the dropdown
  customers = signal<{id: string, name: string}[]>([]);

  // Filters & Sort
  searchQuery = signal('');
  selectedRating = signal(0);
  sortBy = signal<'newest' | 'oldest' | 'highest' | 'lowest'>('newest');

  // Pagination
  currentPage = signal(1);
  pageSize = 12;

  // New Feedback Drawer
  isDrawerOpen = signal(false);
  submitting = signal(false);
  newFeedbackCustomerId = signal('');
  newFeedbackRating = signal(5);
  newFeedbackComment = signal('');

  // Keywords to highlight in comments
  private keywords = ['amazing', 'slow', 'expensive', 'quality', 'staff', 'fast', 'friendly', 'rude', 'terrible', 'excellent', 'service'];

  ngOnInit() {
    this.loadFeedback();
    this.customerService.getCustomers({ limit: 500 }).subscribe({
      next: (res) => {
        const map: Record<string, string> = {};
        const list: {id: string, name: string}[] = [];
        (res.data || []).forEach(c => { 
          if (c.id) {
            map[c.id] = c.name; 
            list.push({id: c.id, name: c.name});
          }
        });
        this.customerNames.set(map);
        this.customers.set(list);
      }
    });
  }

  loadFeedback() {
    this.loading.set(true);
    this.crmService.getFeedbackList().subscribe({
      next: (res) => {
        this.feedbackList.set(res.feedback || []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });
  }

  // --- KPIs ---
  get totalReviews() { return this.feedbackList().length; }

  get averageRating() {
    const list = this.feedbackList();
    if (!list.length) return 0;
    return list.reduce((s, f) => s + f.rating, 0) / list.length;
  }

  get satisfactionRate() {
    if (!this.totalReviews) return 0;
    return (this.positiveCount / this.totalReviews) * 100;
  }

  get positiveCount() { return this.feedbackList().filter(f => f.rating >= 4).length; }
  get neutralCount()  { return this.feedbackList().filter(f => f.rating === 3).length; }
  get negativeCount() { return this.feedbackList().filter(f => f.rating <= 2).length; }

  // Count per exact star (1–5)
  starCount(n: number) { return this.feedbackList().filter(f => f.rating === n).length; }
  starPercent(n: number) { 
    if (!this.totalReviews) return 0;
    return (this.starCount(n) / this.totalReviews) * 100; 
  }

  // --- Filter + Sort + Paginate ---
  get sortedFiltered() {
    const q = this.searchQuery().toLowerCase();
    const r = this.selectedRating();
    let list = this.feedbackList().filter(f =>
      (!r || f.rating === r) &&
      (!q || f.comment?.toLowerCase().includes(q) || this.customerName(f.customer_id).toLowerCase().includes(q))
    );

    switch (this.sortBy()) {
      case 'oldest':  list = [...list].sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()); break;
      case 'highest': list = [...list].sort((a, b) => b.rating - a.rating); break;
      case 'lowest':  list = [...list].sort((a, b) => a.rating - b.rating); break;
      default:        list = [...list].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    }
    return list;
  }

  get totalPages() { return Math.max(1, Math.ceil(this.sortedFiltered.length / this.pageSize)); }

  get pagedFeedback() {
    const start = (this.currentPage() - 1) * this.pageSize;
    return this.sortedFiltered.slice(start, start + this.pageSize);
  }

  get pages(): number[] {
    return Array.from({ length: this.totalPages }, (_, i) => i + 1);
  }

  goToPage(p: number) {
    if (p >= 1 && p <= this.totalPages) this.currentPage.set(p);
  }

  onFilterChange() { this.currentPage.set(1); }

  // --- Delete ---
  async deleteFeedback(fb: CustomerFeedback) {
    if (!(await this.alertService.confirm(`Delete this feedback from "${this.customerName(fb.customer_id)}"?`, 'Delete Feedback'))) return;
    this.crmService.deleteFeedback(fb.id).subscribe({
      next: () => this.feedbackList.update(list => list.filter(f => f.id !== fb.id)),
      error: () => this.toastService.showError('Failed to delete feedback.')
    });
  }

  // --- Helpers ---
  customerName(id: string): string {
    return this.customerNames()[id] || 'Anonymous';
  }

  customerInitials(id: string): string {
    const name = this.customerName(id);
    if (!name || name === 'Anonymous') return 'A';
    return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);
  }

  starsArray(n: number): number[] { return Array.from({ length: n }); }
  emptyStarsArray(n: number): number[] { return Array.from({ length: 5 - n }); }

  ratingColor(rating: number): string {
    if (rating >= 4) return 'text-emerald-500';
    if (rating === 3) return 'text-amber-500';
    return 'text-rose-500';
  }

  sentimentLabel(rating: number): string {
    if (rating >= 4) return 'Positive';
    if (rating === 3) return 'Neutral';
    return 'Negative';
  }

  sentimentClass(rating: number): string {
    if (rating >= 4) return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-900/20 dark:text-emerald-400 dark:border-emerald-800';
    if (rating === 3) return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-900/20 dark:text-amber-400 dark:border-amber-800';
    return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-900/20 dark:text-rose-400 dark:border-rose-800';
  }

  satisfactionColor(): string {
    const r = this.satisfactionRate;
    if (r >= 80) return 'text-emerald-500';
    if (r >= 60) return 'text-amber-500';
    return 'text-rose-500';
  }

  timeAgo(dateStr: string): string {
    const diff = Date.now() - new Date(dateStr).getTime();
    const mins  = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days  = Math.floor(diff / 86400000);
    if (days > 0)  return `${days}d ago`;
    if (hours > 0) return `${hours}h ago`;
    if (mins > 0)  return `${mins}m ago`;
    return 'Just now';
  }

  // --- Keyword Highlighting Utility ---
  highlightKeywords(text: string): string {
    if (!text) return '';
    let highlightedText = text;
    this.keywords.forEach(keyword => {
      // Regex for whole word, case insensitive
      const regex = new RegExp(`\\b(${keyword})\\b`, 'gi');
      highlightedText = highlightedText.replace(regex, '<strong class="text-violet-600 dark:text-violet-400 font-bold bg-violet-50 dark:bg-violet-900/20 px-1 rounded">$1</strong>');
    });
    return highlightedText;
  }

  // --- Drawer ---
  openDrawer() { this.isDrawerOpen.set(true); }
  closeDrawer() {
    this.isDrawerOpen.set(false);
    this.newFeedbackCustomerId.set('');
    this.newFeedbackRating.set(5);
    this.newFeedbackComment.set('');
  }

  submitFeedback() {
    if (!this.newFeedbackCustomerId()) return;
    this.submitting.set(true);
    this.crmService.createFeedback({
      customer_id: this.newFeedbackCustomerId(),
      rating: this.newFeedbackRating(),
      comment: this.newFeedbackComment()
    }).subscribe({
      next: (fb) => {
        this.feedbackList.update(list => [fb, ...list]);
        this.submitting.set(false);
        this.closeDrawer();
      },
      error: () => {
        this.submitting.set(false);
        this.toastService.showError('Failed to submit feedback.');
      }
    });
  }
}
