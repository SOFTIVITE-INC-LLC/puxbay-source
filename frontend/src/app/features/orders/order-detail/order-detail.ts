import { Component, inject, OnInit, OnDestroy, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { OrderService } from '../../../core/services/order.service';
import { Order } from '../../../core/models/order.models';
import { ReceiptComponent } from '../../../shared/components/receipt/receipt.component';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { PosOverrideModalComponent } from '../../../shared/components/pos-override-modal/pos-override-modal';
import { AlertService } from '../../../core/services/alert.service';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';
import { NotificationSoundService } from '../../../core/services/notification-sound.service';

export interface SplitPaymentRow {
  method: string;
  amount: number;
  ref?: string;
}

@Component({
  selector: 'app-order-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, ReceiptComponent, AppCurrencyPipe, PosOverrideModalComponent],
  templateUrl: './order-detail.html'
})
export class OrderDetail implements OnInit, OnDestroy {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private alertService = inject(AlertService);
  public orderService = inject(OrderService);
  public storefrontSettingsService = inject(StorefrontSettingsService);
  private soundService = inject(NotificationSoundService);

  order = signal<Order | null>(null);
  loading = signal(true);
  voidingOrder = signal(false);
  completingOrder = signal(false);
  showVoidConfirm = signal(false);
  showOverrideModal = signal(false);

  // OTP Verification state
  sendingOtp = signal(false);
  verifyingOtp = signal(false);
  otpCode = signal('');
  otpSent = signal(false);
  otpMaskedPhone = signal('');
  otpCountdown = signal(0);
  showOtpModal = signal(false);
  copiedField = signal<string | null>(null);

  // Payment Modal & Split Payment State
  showPaymentModal = signal(false);
  selectedPaymentMethod = signal<'cash' | 'mobile' | 'card' | 'bank_transfer' | 'split'>('cash');
  tenderedAmount = signal<number>(0);
  momoNetwork = signal<'MTN' | 'Telecel' | 'AT'>('MTN');
  momoPhoneNumber = signal<string>('');
  momoRef = signal<string>('');
  cardAuthCode = signal<string>('');
  bankRef = signal<string>('');
  paymentNotes = signal<string>('');
  splitPayments = signal<SplitPaymentRow[]>([
    { method: 'cash', amount: 0 },
    { method: 'mobile', amount: 0 }
  ]);

  private timerInterval: any = null;

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.loadOrder(id);
    } else {
      this.goBack();
    }

    // Load storefront settings (to check Paystack credentials)
    this.storefrontSettingsService.loadSettings().subscribe();

    // Dynamically inject Paystack inline script if not present
    if (typeof document !== 'undefined' && !document.getElementById('paystack-inline-js')) {
      const s = document.createElement('script');
      s.id = 'paystack-inline-js';
      s.src = 'https://js.paystack.co/v1/inline.js';
      document.head.appendChild(s);
    }
  }

  ngOnDestroy() {
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
    }
  }

  loadOrder(id: string) {
    this.loading.set(true);
    this.orderService.getOrder(id).subscribe({
      next: (order) => {
        this.order.set(order);
        this.loading.set(false);
        if (order) {
          this.tenderedAmount.set(order.total || 0);
          this.momoPhoneNumber.set(this.getCustomerPhone(order) || '');
          this.initSplitRows(order.total || 0);
        }
      },
      error: () => {
        this.loading.set(false);
        this.goBack();
      }
    });
  }

  goBack() {
    this.router.navigate(['/orders']);
  }

  getCustomerName(order: Order | null): string {
    if (!order) return 'Unknown';
    if (order.customer_name) return order.customer_name;
    if (order.customer) {
      const name = `${order.customer.first_name || ''} ${order.customer.last_name || ''}`.trim() || order.customer.name;
      if (name) return name;
    }
    if (order.customer_id) return 'Walk-in Customer';
    return 'Walk-in Customer';
  }

  getCustomerPhone(order: Order | null): string | null {
    if (!order) return null;
    if (order.customer_phone && order.customer_phone.trim()) return order.customer_phone.trim();
    if (order.customer?.phone && order.customer.phone.trim()) return order.customer.phone.trim();
    if (order.notes) {
      const match = order.notes.match(/Phone:\s*([+0-9\s()-]+)/i);
      if (match) return match[1].trim();
    }
    return null;
  }

  getCustomerEmail(order: Order | null): string | null {
    if (!order) return null;
    if (order.customer?.email && order.customer.email.trim()) return order.customer.email.trim();
    return null;
  }

  getDeliveryAddress(order: Order | null): string | null {
    if (!order) return null;
    if (order.delivery_address && order.delivery_address.trim()) return order.delivery_address.trim();
    if (order.customer?.address && order.customer.address.trim()) return order.customer.address.trim();
    if (order.notes) {
      const match = order.notes.match(/Delivery Address:\s*([^\n]+)/i);
      if (match) return match[1].trim();
    }
    return null;
  }

  openGoogleMaps(address: string) {
    if (!address) return;
    const query = encodeURIComponent(address);
    window.open(`https://www.google.com/maps/search/?api=1&query=${query}`, '_blank');
  }

  copyText(text: string, label: string) {
    if (!text) return;
    navigator.clipboard.writeText(text).then(() => {
      this.copiedField.set(label);
      setTimeout(() => this.copiedField.set(null), 2000);
    });
  }

  isOnlineOrKiosk(order: Order | null): boolean {
    if (!order) return false;
    const t = order.order_type?.toLowerCase();
    return t === 'online' || t === 'kiosk' || t === 'delivery' || t === 'storefront' || t === 'pickup';
  }

  copyToClipboard(text: string, field: string) {
    this.copyText(text, field);
  }

  callPhone(phone: string) {
    if (!phone) return;
    window.open(`tel:${phone}`, '_self');
  }

  sendPickupOTP() {
    const o = this.order();
    if (!o?.id) return;
    this.sendingOtp.set(true);
    this.orderService.sendPickupOTP(o.id).subscribe({
      next: (res) => {
        this.sendingOtp.set(false);
        this.otpSent.set(true);
        this.otpMaskedPhone.set(res.masked || res.phone);
        this.alertService.alert(`Verification code dispatched to ${res.masked || res.phone}`, 'OTP Sent', 'success');
        this.otpCountdown.set(60);
        if (this.timerInterval) clearInterval(this.timerInterval);
        this.timerInterval = setInterval(() => {
          const current = this.otpCountdown();
          if (current > 1) {
            this.otpCountdown.set(current - 1);
          } else {
            this.otpCountdown.set(0);
            clearInterval(this.timerInterval);
          }
        }, 1000);
      },
      error: (err) => {
        this.sendingOtp.set(false);
        this.alertService.alert(err.error?.error || 'Failed to dispatch verification code', 'OTP Error', 'danger');
      }
    });
  }

  verifyPickupOTP() {
    const o = this.order();
    const code = this.otpCode().trim();
    if (!o?.id || !code) return;
    this.verifyingOtp.set(true);
    this.orderService.verifyPickupOTP(o.id, code).subscribe({
      next: () => {
        this.verifyingOtp.set(false);
        this.otpCode.set('');
        this.showOtpModal.set(false);
        this.alertService.alert('Customer phone and identity confirmed!', 'Verification Successful', 'success');
        this.loadOrder(o.id!);
      },
      error: (err) => {
        this.verifyingOtp.set(false);
        this.alertService.alert(err.error?.error || 'Invalid or expired OTP code', 'Verification Failed', 'danger');
      }
    });
  }

  confirmVoidOrder() { this.showVoidConfirm.set(true); }
  cancelVoid() { this.showVoidConfirm.set(false); }

  voidOrder(overridePin?: string) {
    const o = this.order();
    if (!o?.id) return;
    this.voidingOrder.set(true);
    this.showVoidConfirm.set(false);
    this.orderService.voidOrder(o.id, overridePin).subscribe({
      next: () => {
        this.voidingOrder.set(false);
        this.showOverrideModal.set(false);
        this.loadOrder(o.id!);
      },
      error: (err) => { 
        this.voidingOrder.set(false); 
        if (err.status === 403 && err.error?.error?.includes('Manager override')) {
          this.showOverrideModal.set(true);
        } else if (overridePin) {
          this.alertService.alert(err.error?.error || 'Invalid override PIN', 'PIN Override Error', 'danger');
        }
      }
    });
  }

  // Open Payment Modal
  markPaid() {
    const o = this.order();
    if (!o?.id) return;
    this.tenderedAmount.set(o.total || 0);
    this.momoPhoneNumber.set(this.getCustomerPhone(o) || '');
    this.initSplitRows(o.total || 0);
    this.showPaymentModal.set(true);
  }

  initSplitRows(total: number) {
    const half = Math.round((total / 2) * 100) / 100;
    const remaining = Math.round((total - half) * 100) / 100;
    this.splitPayments.set([
      { method: 'cash', amount: half },
      { method: 'mobile', amount: remaining }
    ]);
  }

  addSplitRow() {
    this.splitPayments.update(list => [...list, { method: 'card', amount: 0 }]);
  }

  removeSplitRow(index: number) {
    this.splitPayments.update(list => list.filter((_, i) => i !== index));
  }

  getSplitTotal(): number {
    return this.splitPayments().reduce((sum, p) => sum + (Number(p.amount) || 0), 0);
  }

  getSplitDifference(): number {
    const total = this.order()?.total || 0;
    return Math.round((this.getSplitTotal() - total) * 100) / 100;
  }

  // Pay with Paystack inline popup
  payWithPaystack() {
    const settings = this.storefrontSettingsService.settings();
    const pubKey = settings?.paystack_public_key;
    if (!pubKey) {
      this.alertService.alert('Paystack public key is not configured in Settings. Please enter the transaction reference manually.', 'Paystack Gateway', 'warning');
      return;
    }

    const o = this.order();
    if (!o) return;
    const total = o.total || 0;
    const email = this.getCustomerEmail(o) || 'customer@puxbay.com';
    const amountInPesewas = Math.round(total * 100);

    const setupConfig: any = {
      key: pubKey,
      email,
      amount: amountInPesewas,
      currency: settings?.currency || 'GHS',
      channels: ['mobile_money', 'card', 'bank', 'ussd', 'qr'],
      callback: (response: any) => {
        this.momoRef.set(response.reference);
        this.submitPayment('mobile', total, `Paystack Ref: ${response.reference}`);
      },
      onClose: () => {
        this.alertService.alert('Paystack payment prompt closed.', 'Payment Cancelled', 'info');
      }
    };

    if (settings?.paystack_subaccount_code) {
      setupConfig.subaccount = settings.paystack_subaccount_code;
    }

    if (typeof window !== 'undefined' && (window as any).PaystackPop) {
      const handler = (window as any).PaystackPop.setup(setupConfig);
      handler.openIframe();
    } else {
      this.alertService.alert('Paystack inline library is still loading. Please try again in a second.', 'Payment Gateway', 'info');
    }
  }

  // Submit payment and record order as paid
  submitPayment(overrideMethod?: string, overrideAmount?: number, extraNotes?: string) {
    const o = this.order();
    if (!o?.id) return;

    const total = o.total || 0;
    const method = overrideMethod || this.selectedPaymentMethod();
    let amountPaid = overrideAmount !== undefined ? overrideAmount : (method === 'cash' ? (this.tenderedAmount() || total) : total);
    let splitDetails = '';

    if (method === 'split') {
      const diff = this.getSplitDifference();
      if (diff < 0) {
        this.alertService.alert(`Split payment total is less than order total by ${Math.abs(diff).toFixed(2)}. Please balance the amounts.`, 'Incomplete Split Amount', 'warning');
        return;
      }
      splitDetails = this.splitPayments()
        .filter(p => (Number(p.amount) || 0) > 0)
        .map(p => `${p.method.toUpperCase()}: ${p.amount}${p.ref ? ` (${p.ref})` : ''}`)
        .join(', ');
      amountPaid = this.getSplitTotal();
    }

    let notes = extraNotes || this.paymentNotes();
    if (method === 'mobile' && this.momoRef()) {
      notes = notes ? `${notes} | MoMo Ref: ${this.momoRef()}` : `MoMo Ref: ${this.momoRef()}`;
    } else if (method === 'card' && this.cardAuthCode()) {
      notes = notes ? `${notes} | Card Auth: ${this.cardAuthCode()}` : `Card Auth: ${this.cardAuthCode()}`;
    } else if (method === 'bank_transfer' && this.bankRef()) {
      notes = notes ? `${notes} | Bank Ref: ${this.bankRef()}` : `Bank Ref: ${this.bankRef()}`;
    }

    this.completingOrder.set(true);
    const payload = {
      override_pin: 'OVERRIDE',
      payment_method: method,
      amount_paid: amountPaid,
      payment_status: 'paid',
      notes,
      split_details: splitDetails
    };

    this.orderService.completeOrder(o.id, payload).subscribe({
      next: () => {
        this.completingOrder.set(false);
        this.showPaymentModal.set(false);
        this.soundService.playPosCompletedSound();
        this.alertService.alert(`Order #${o.order_number} has been recorded as PAID via ${method.toUpperCase()}.`, 'Payment Completed', 'success');
        this.loadOrder(o.id!);
      },
      error: (err) => {
        this.completingOrder.set(false);
        this.alertService.alert(err.error?.error || 'Failed to record payment', 'Payment Error', 'danger');
      }
    });
  }

  markCompleted(overridePin?: string) {
    const o = this.order();
    if (!o?.id) return;

    // Check if online/kiosk order is unverified
    if (this.isOnlineOrKiosk(o) && !o.is_otp_verified && !overridePin) {
      this.showOtpModal.set(true);
      return;
    }

    // If order is un-paid, open Payment Modal so user can choose payment method & Paystack/Cash/Split
    if (o.payment_status !== 'paid' || (o.amount_paid || 0) < (o.total || 0)) {
      this.markPaid();
      return;
    }

    this.completingOrder.set(true);
    this.orderService.completeOrder(o.id, overridePin).subscribe({
      next: () => {
        this.completingOrder.set(false);
        this.showOverrideModal.set(false);
        this.showOtpModal.set(false);
        this.soundService.playPosCompletedSound();
        this.loadOrder(o.id!);
      },
      error: (err) => {
        this.completingOrder.set(false);
        if (err.status === 403 && err.error?.error?.includes('Manager override')) {
          this.showOverrideModal.set(true);
        } else {
          this.alertService.alert(err.error?.error || 'Failed to complete order', 'Order Completion Error', 'danger');
        }
      }
    });
  }

  quickReturn() {
    const o = this.order();
    if (!o?.id) return;
    this.router.navigate(['/returns'], { queryParams: { new: 'true', order_id: o.id } });
  }

  printReceipt() {
    window.print();
  }
}
