import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MarketingService, Campaign, CustomerSegment } from '../../../core/services/marketing.service';
import { ToastService } from '../../../core/services/toast';

@Component({
  selector: 'app-marketing',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './marketing.html',
})
export class Marketing implements OnInit {
  marketingService = inject(MarketingService);
  private toastService = inject(ToastService);

  activeTab = signal<'campaigns' | 'segments' | 'promotions' | 'discounts'>('campaigns');

  // Campaign Modal
  isModalOpen = signal(false);
  modalTitle = signal('Create Campaign');
  currentCampaign = signal<Partial<Campaign>>({
    status: 'draft',
    budget: 0,
    type: 'email',
    is_automated: false,
    trigger_event: 'manual',
  });

  searchQuery = signal('');

  // Overview Stats
  totalCampaigns = computed(() => this.marketingService.campaigns().length);
  activeCampaigns = computed(() => this.marketingService.campaigns().filter(c => c.status === 'active').length);
  totalBudget = computed(() => this.marketingService.campaigns().reduce((sum, c) => sum + (Number(c.budget) || 0), 0));
  totalRevenueGenerated = computed(() =>
    this.marketingService.campaigns().reduce((sum, c) => sum + (Number((c as any).revenue_generated) || 0), 0)
  );

  get filteredCampaigns() {
    const q = this.searchQuery().toLowerCase();
    return this.marketingService.campaigns().filter(c =>
      (c.name || '').toLowerCase().includes(q) ||
      (c.type || '').toLowerCase().includes(q)
    );
  }

  ngOnInit() {
    this.marketingService.getCampaigns().subscribe();
    this.marketingService.getPromotions().subscribe();
    this.marketingService.getDiscounts().subscribe();
    this.marketingService.getSegments().subscribe();
  }

  // ── Campaign CRUD ────────────────────────────────────────────────────

  openAddModal() {
    this.modalTitle.set('Create Campaign');
    this.currentCampaign.set({ status: 'draft', budget: 0, type: 'email', is_automated: false, trigger_event: 'manual' });
    this.isModalOpen.set(true);
  }

  openEditModal(campaign: Campaign) {
    this.modalTitle.set('Edit Campaign');
    this.currentCampaign.set({ ...campaign });
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  saveCampaign() {
    const c = this.currentCampaign();
    if (c.id) {
      this.marketingService.updateCampaign(c.id, c).subscribe({
        next: () => {
          this.toastService.showSuccess('Campaign updated successfully');
          this.closeModal();
        },
        error: () => this.toastService.showError('Failed to update campaign')
      });
    } else {
      this.marketingService.createCampaign(c).subscribe({
        next: () => {
          this.toastService.showSuccess('Campaign created successfully');
          this.closeModal();
        },
        error: () => this.toastService.showError('Failed to create campaign')
      });
    }
  }

  deleteCampaign(id: string, event: Event) {
    event.stopPropagation();
    this.marketingService.deleteCampaign(id).subscribe({
      next: () => this.toastService.showSuccess('Campaign deleted'),
      error: () => this.toastService.showError('Failed to delete campaign')
    });
  }

  sendCampaign(campaign: Campaign, event: Event) {
    event.stopPropagation();
    this.marketingService.sendCampaign(campaign.id).subscribe({
      next: () => this.toastService.showSuccess(`Campaign "${campaign.name}" sent!`),
      error: (e: any) => this.toastService.showError(e?.error?.error || 'Failed to send campaign')
    });
  }

  toggleCampaignStatus(campaign: Campaign, event: Event) {
    event.stopPropagation();
    const newStatus = campaign.status === 'active' ? 'paused' : 'active';
    this.marketingService.updateCampaign(campaign.id, { ...campaign, status: newStatus }).subscribe({
      next: () => this.toastService.showSuccess(`Campaign ${newStatus}`),
      error: () => this.toastService.showError('Failed to update status')
    });
  }

  triggerEvent(eventType: string) {
    this.marketingService.triggerEventCampaigns(eventType).subscribe({
      next: (res: any) => this.toastService.showSuccess(`Triggered ${res.triggered} campaign(s) for "${eventType}"`),
      error: () => this.toastService.showError('Failed to trigger campaigns')
    });
  }

  // ── Segments ────────────────────────────────────────────────────────

  isSegmentModalOpen = signal(false);
  editingSegmentId = signal<string | null>(null);
  currentSegment = signal<Partial<CustomerSegment>>({ name: '', description: '', criteria_json: '{}' });
  savingSegment = signal(false);

  openSegmentModal(segment?: CustomerSegment) {
    if (segment) {
      this.editingSegmentId.set(segment.id);
      this.currentSegment.set({ ...segment });
    } else {
      this.editingSegmentId.set(null);
      this.currentSegment.set({ name: '', description: '', criteria_json: '{}' });
    }
    this.isSegmentModalOpen.set(true);
  }

  closeSegmentModal() { this.isSegmentModalOpen.set(false); }

  saveSegment() {
    const s = this.currentSegment();
    this.savingSegment.set(true);
    const editId = this.editingSegmentId();
    const request$ = editId
      ? this.marketingService.updateSegment(editId, s)
      : this.marketingService.createSegment(s);

    request$.subscribe({
      next: () => {
        this.savingSegment.set(false);
        this.toastService.showSuccess(editId ? 'Segment updated' : 'Segment created');
        this.closeSegmentModal();
      },
      error: () => {
        this.savingSegment.set(false);
        this.toastService.showError('Failed to save segment');
      }
    });
  }

  deleteSegment(id: string) {
    this.marketingService.deleteSegment(id).subscribe({
      next: () => this.toastService.showSuccess('Segment deleted'),
      error: () => this.toastService.showError('Failed to delete segment')
    });
  }

  // ── Promotions ──────────────────────────────────────────────────────

  isPromoModalOpen = signal(false);
  currentPromo = signal<any>({ status: 'draft', type: 'bogo', name: '' });

  openPromoModal(promo?: any) {
    this.currentPromo.set(promo ? { ...promo } : { status: 'draft', type: 'bogo', name: '' });
    this.isPromoModalOpen.set(true);
  }
  closePromoModal() { this.isPromoModalOpen.set(false); }

  savePromotion() {
    const p = this.currentPromo();
    const request$ = p.id
      ? this.marketingService.updatePromotion(p.id, p)
      : this.marketingService.createPromotion(p);

    request$.subscribe({
      next: () => {
        this.toastService.showSuccess(p.id ? 'Promotion updated' : 'Promotion created');
        this.closePromoModal();
      },
      error: () => this.toastService.showError('Failed to save promotion')
    });
  }

  deletePromotion(id: string) {
    this.marketingService.deletePromotion(id).subscribe({
      next: () => this.toastService.showSuccess('Promotion deleted'),
      error: () => this.toastService.showError('Failed to delete promotion')
    });
  }

  // ── Discounts ───────────────────────────────────────────────────────

  isDiscountModalOpen = signal(false);
  currentDiscount = signal<any>({ status: 'active', type: 'percentage', value: 10, current_uses: 0, code: '' });

  openDiscountModal() {
    this.currentDiscount.set({ status: 'active', type: 'percentage', value: 10, current_uses: 0, code: '' });
    this.isDiscountModalOpen.set(true);
  }
  closeDiscountModal() { this.isDiscountModalOpen.set(false); }

  saveDiscount() {
    this.marketingService.createDiscount(this.currentDiscount()).subscribe({
      next: () => {
        this.toastService.showSuccess('Discount code created');
        this.closeDiscountModal();
      },
      error: () => this.toastService.showError('Failed to create discount')
    });
  }

  deleteDiscount(id: string) {
    this.marketingService.deleteDiscount(id).subscribe({
      next: () => this.toastService.showSuccess('Discount code deleted'),
      error: () => this.toastService.showError('Failed to delete discount code')
    });
  }

  // ── Loyalty Redemption ──────────────────────────────────────────────

  isRedeemModalOpen = signal(false);
  redeemForm = signal({ customer_id: '', points: 100, discount_value: 10 });

  openRedeemModal() { this.isRedeemModalOpen.set(true); }
  closeRedeemModal() { this.isRedeemModalOpen.set(false); }

  redeemPoints() {
    const f = this.redeemForm();
    if (!f.customer_id || f.points <= 0 || f.discount_value <= 0) {
      this.toastService.showError('Please fill all fields correctly');
      return;
    }
    this.marketingService.redeemPointsForDiscount(f.customer_id, f.points, f.discount_value).subscribe({
      next: (code: any) => {
        this.toastService.showSuccess(`Discount code "${code.code}" generated!`);
        this.closeRedeemModal();
        this.marketingService.getDiscounts().subscribe();
      },
      error: () => this.toastService.showError('Failed to redeem points')
    });
  }
}
