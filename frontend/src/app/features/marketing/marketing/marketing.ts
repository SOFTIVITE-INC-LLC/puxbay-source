import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MarketingService, Campaign, CustomerSegment } from '../../../core/services/marketing.service';
import { SMSService, SMSSenderID } from '../../../core/services/sms.service';
import { ToastService } from '../../../core/services/toast';
import { SettingsService } from '../../../core/services/settings.service';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';

@Component({
  selector: 'app-marketing',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './marketing.html',
  host: { class: 'block w-full min-h-full' },
})
export class Marketing implements OnInit {
  marketingService = inject(MarketingService);
  smsService = inject(SMSService);
  private toastService = inject(ToastService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  settingsService = inject(SettingsService);
  storefrontSettingsService = inject(StorefrontSettingsService);

  activeTab = signal<'campaigns' | 'segments' | 'promotions' | 'discounts' | 'sms'>('campaigns');

  setTab(tabId: string) {
    this.activeTab.set(tabId as any);
    if (tabId === 'sms') {
      this.loadSMSData();
    }
  }

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

    // Automatic verification when returning from Paystack redirect
    this.route.queryParams.subscribe(params => {
      const tab = params['tab'];
      if (tab) {
        this.setTab(tab);
      }

      const ref = params['reference'] || params['trxref'];
      if (ref && ref.startsWith('SMS-TOPUP-')) {
        this.setTab('sms');
        this.smsService.verifyTopup(ref).subscribe({
          next: (res: any) => {
            this.toastService.showSuccess(`Payment verified! ${res.credits_added || ''} SMS credits added to your wallet.`);
            this.loadSMSData();
            // Clean up reference parameter from URL so it doesn't re-trigger
            this.router.navigate([], {
              relativeTo: this.route,
              queryParams: { tab: 'sms' },
              replaceUrl: true
            });
          },
          error: () => {
            this.loadSMSData();
            this.router.navigate([], {
              relativeTo: this.route,
              queryParams: { tab: 'sms' },
              replaceUrl: true
            });
          }
        });
      }
    });
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

  // ── SMS & Sender ID ──────────────────────────────────────────────────

  isSenderIDModalOpen = signal(false);
  newSenderID = signal('');
  newSenderPurpose = signal('');
  isSubmittingSenderID = signal(false);

  isTopupModalOpen = signal(false);
  topupAmount = signal<number>(20);
  topupEmail = signal('');
  isProcessingTopup = signal(false);

  // SMS Preview Simulator
  previewTemplate = signal<'receipt' | 'delivery' | 'promo' | 'cart'>('receipt');
  
  get previewMessage(): string {
    const brand = this.smsService.activeSenderID()?.sender_id || 'PUXBAY';
    switch (this.previewTemplate()) {
      case 'receipt':
        return `Hi Alex! Your order #PB-4821 of GHS 180.00 at ${brand} has been confirmed. Track status in your dashboard. Thank you!`;
      case 'delivery':
        return `Your package from ${brand} is out for delivery with rider Kwame (0244123456). Expect delivery within 45 mins.`;
      case 'promo':
        return `FLASH SALE! Get 25% OFF all store items today only at ${brand}. Use code FLASH25 at checkout: puxbay.com/s/store`;
      case 'cart':
        return `You left items in your cart at ${brand}! Complete your checkout today and get free delivery on us: puxbay.com/s/cart`;
    }
  }

  loadSMSData() {
    this.smsService.getWallet().subscribe({
      error: () => {}
    });
    this.smsService.getSenderIDs().subscribe({
      error: () => {}
    });
    this.smsService.getTransactions().subscribe({
      error: () => {}
    });
  }

  openSenderIDModal() {
    this.newSenderID.set('');
    this.newSenderPurpose.set('');
    this.isSenderIDModalOpen.set(true);
  }

  closeSenderIDModal() {
    this.isSenderIDModalOpen.set(false);
  }

  submitSenderID() {
    const id = this.newSenderID().trim().toUpperCase();
    const purpose = this.newSenderPurpose().trim();
    if (!id || id.length < 3 || id.length > 11) {
      this.toastService.showError('Sender ID must be between 3 and 11 alphanumeric characters');
      return;
    }
    this.isSubmittingSenderID.set(true);
    this.smsService.submitSenderID(id, purpose).subscribe({
      next: () => {
        this.toastService.showSuccess(`Sender ID "${id}" submitted for approval!`);
        this.isSubmittingSenderID.set(false);
        this.closeSenderIDModal();
        this.smsService.getSenderIDs().subscribe();
      },
      error: (err) => {
        this.toastService.showError(err.error?.error || 'Failed to submit Sender ID');
        this.isSubmittingSenderID.set(false);
      }
    });
  }

  openTopupModal() {
    this.topupAmount.set(20);
    this.isTopupModalOpen.set(true);
  }

  closeTopupModal() {
    this.isTopupModalOpen.set(false);
  }

  setTopupAmount(amt: number) {
    this.topupAmount.set(amt);
  }

  calculateCredits(amt: number): number {
    const rate = this.smsService.rate() || 0.20;
    return Math.floor(amt / rate);
  }

  processSMSTopup() {
    const amt = this.topupAmount();
    if (!amt || amt < 1) {
      this.toastService.showError('Please enter a valid top-up amount');
      return;
    }

    const email = this.topupEmail().trim() || 'billing@puxbay.com';
    this.isProcessingTopup.set(true);

    this.smsService.initiateTopup(amt, email).subscribe({
      next: (res: any) => {
        // If Paystack returned an authorization URL (Standard Checkout)
        if (res.authorization_url) {
          this.isProcessingTopup.set(false);
          this.closeTopupModal();
          window.location.href = res.authorization_url;
          return;
        }

        const publicKey = res.public_key || this.storefrontSettingsService.settings()?.paystack_public_key;

        // If Paystack inline is available on window
        if ((window as any).PaystackPop && publicKey) {
          const handler = (window as any).PaystackPop.setup({
            key: publicKey,
            email: email,
            amount: Math.round(amt * 100), // kobo/pesewas
            currency: this.smsService.currency() || 'GHS',
            ref: res.reference,
            callback: (response: any) => {
              this.smsService.verifyTopup(response.reference).subscribe({
                next: () => {
                  this.toastService.showSuccess(`${res.credits} SMS credits added to your wallet!`);
                  this.isProcessingTopup.set(false);
                  this.closeTopupModal();
                  this.loadSMSData();
                },
                error: (err) => {
                  this.toastService.showError(err.error?.error || 'Failed to verify payment');
                  this.isProcessingTopup.set(false);
                }
              });
            },
            onClose: () => {
              this.isProcessingTopup.set(false);
            }
          });
          handler.openIframe();
        } else {
          // If no Paystack key is set on platform, verify directly (sandbox/demo)
          this.smsService.verifyTopup(res.reference).subscribe({
            next: () => {
              this.toastService.showSuccess(`${res.credits} SMS credits added to your wallet!`);
              this.isProcessingTopup.set(false);
              this.closeTopupModal();
              this.loadSMSData();
            },
            error: (err) => {
              this.toastService.showError(err.error?.error || 'Failed to verify payment');
              this.isProcessingTopup.set(false);
            }
          });
        }
      },
      error: (err) => {
        this.toastService.showError(err.error?.error || 'Failed to initiate top-up');
        this.isProcessingTopup.set(false);
      }
    });
  }
}
