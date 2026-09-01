import { Component, inject, OnInit, OnDestroy, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { OrderService } from '../../../core/services/order.service';
import { Order } from '../../../core/models/order.models';
import { ReceiptComponent } from '../../../shared/components/receipt/receipt.component';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { PosOverrideModalComponent } from '../../../shared/components/pos-override-modal/pos-override-modal';
import { AlertService } from '../../../core/services/alert.service';

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

  private timerInterval: any = null;

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.loadOrder(id);
    } else {
      this.goBack();
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
    if (order.notes) {
      const match = order.notes.match(/Delivery Address:\s*(.*)/i);
      if (match) return match[1].trim();
    }
    return null;
  }

  isOnlineOrKiosk(order: Order | null): boolean {
    if (!order) return false;
    const type = (order.order_type || '').toLowerCase();
    return ['online', 'kiosk', 'delivery', 'storefront', 'pickup'].includes(type);
  }

  openGoogleMaps(address: string) {
    const query = encodeURIComponent(address);
    window.open(`https://www.google.com/maps/search/?api=1&query=${query}`, '_blank');
  }

  copyText(text: string, label: string) {
    navigator.clipboard.writeText(text).then(() => {
      this.copiedField.set(label);
      setTimeout(() => this.copiedField.set(null), 2000);
    });
  }

  callPhone(phone: string) {
    window.open(`tel:${phone}`, '_self');
  }

  // --- OTP Verification Flow ---
  sendPickupOTP() {
    const o = this.order();
    if (!o?.id) return;
    this.sendingOtp.set(true);
    this.orderService.sendPickupOTP(o.id).subscribe({
      next: (res) => {
        this.sendingOtp.set(false);
        this.otpSent.set(true);
        this.otpMaskedPhone.set(res.masked || res.phone);
        this.startCountdown(60);
        this.alertService.alert(
          `Verification code sent via SMS to ${res.masked || res.phone}. The customer must read out this 6-digit code to verify their order.`,
          'OTP Dispatched',
          'success'
        );
      },
      error: (err) => {
        this.sendingOtp.set(false);
        this.alertService.alert(err.error?.error || 'Failed to dispatch verification code via SMS', 'SMS Dispatch Error', 'danger');
      }
    });
  }

  startCountdown(seconds: number) {
    this.otpCountdown.set(seconds);
    if (this.timerInterval) clearInterval(this.timerInterval);
    this.timerInterval = setInterval(() => {
      const current = this.otpCountdown();
      if (current <= 1) {
        this.otpCountdown.set(0);
        clearInterval(this.timerInterval);
      } else {
        this.otpCountdown.set(current - 1);
      }
    }, 1000);
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
        this.alertService.alert('Customer phone and identity confirmed! You may now hand over the order items.', 'Verification Successful', 'success');
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

  markCompleted(overridePin?: string) {
    const o = this.order();
    if (!o?.id) return;

    // Check if online/kiosk order is unverified
    if (this.isOnlineOrKiosk(o) && !o.is_otp_verified && !overridePin) {
      this.showOtpModal.set(true);
      return;
    }

    this.completingOrder.set(true);
    this.orderService.completeOrder(o.id, overridePin).subscribe({
      next: () => {
        this.completingOrder.set(false);
        this.showOverrideModal.set(false);
        this.showOtpModal.set(false);
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
