import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { RouterModule, Router, ActivatedRoute } from '@angular/router';
import { FormBuilder, ReactiveFormsModule, Validators, FormsModule } from '@angular/forms';
import { CartService } from '../../../core/store/services/cart.service';
import { CheckoutService } from '../../../core/store/services/checkout.service';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-checkout',
  standalone: true,
  imports: [CommonModule, RouterModule, ReactiveFormsModule, FormsModule, AppCurrencyPipe],
  templateUrl: './checkout.component.html'
})
export class CheckoutComponent implements OnInit {
  cartService = inject(CartService);
  checkoutService = inject(CheckoutService);
  settingsService = inject(StorefrontSettingsService);
  http = inject(HttpClient);
  router = inject(Router);
  route = inject(ActivatedRoute);
  fb = inject(FormBuilder);

  isLoading = signal(true);
  isProcessing = signal(false);
  checkoutError = signal('');
  deliveryMethod = signal<'pickup'|'delivery'>('pickup');
  paymentMethod = signal<'online'|'pickup'>('online');
  
  // Promo Code State
  promoCodeInput = signal('');
  appliedPromoCode = signal<string | null>(null);
  discountAmount = signal(0);
  isApplyingPromo = signal(false);
  promoMessage = signal<{text: string, type: 'success' | 'error'} | null>(null);

  locationLoading = signal(false);
  addressSuggestions = signal<any[]>([]);
  isSearchingAddress = signal(false);
  addressSearchTimeout: any;

  checkoutForm = this.fb.group({
    firstName: ['', Validators.required],
    lastName: ['', Validators.required],
    email: ['', [Validators.required, Validators.email]],
    phone: [''],
    address: [''],
    city: [''],
    orderNotes: ['']
  });

  ngOnInit() {
    this.cartService.loadCart();
    
    // Load Paystack Script
    if (typeof document !== 'undefined') {
      const script = document.createElement('script');
      script.src = 'https://js.paystack.co/v1/inline.js';
      document.head.appendChild(script);
    }

    setTimeout(() => {
      if (this.cartService.cartDetails().length === 0) {
        this.router.navigate(['/store/cart']);
      }
      this.isLoading.set(false);
    }, 800);
  }

  onEmailBlur() {
    const emailControl = this.checkoutForm.get('email');
    if (emailControl && emailControl.valid && emailControl.value) {
      this.checkoutService.updateCartEmail(emailControl.value).subscribe({
        next: () => console.log('Cart email updated for recovery'),
        error: (err) => console.error('Failed to update cart email', err)
      });
    }
  }

  applyPromoCode() {
    const code = this.promoCodeInput().trim().toUpperCase();
    if (!code) return;
    
    this.isApplyingPromo.set(true);
    this.promoMessage.set(null);
    
    // Simulate API call
    setTimeout(() => {
      if (code === 'WELCOME10') {
        const subtotal = this.cartService.cartTotal();
        const discount = subtotal * 0.10; // 10% discount
        this.discountAmount.set(discount);
        this.appliedPromoCode.set(code);
        this.promoMessage.set({ text: 'Promo code applied successfully! 10% off.', type: 'success' });
      } else {
        this.discountAmount.set(0);
        this.appliedPromoCode.set(null);
        this.promoMessage.set({ text: 'Invalid or expired promo code.', type: 'error' });
      }
      this.isApplyingPromo.set(false);
    }, 600);
  }

  removePromoCode() {
    this.appliedPromoCode.set(null);
    this.discountAmount.set(0);
    this.promoCodeInput.set('');
    this.promoMessage.set(null);
  }

  getLocation() {
    if (navigator.geolocation) {
      this.locationLoading.set(true);
      navigator.geolocation.getCurrentPosition(
        (position) => {
          const lat = position.coords.latitude;
          const lng = position.coords.longitude;
          this.checkoutForm.patchValue({
            address: `${lat}, ${lng}`,
            city: 'GPS Location'
          });
          this.locationLoading.set(false);
        },
        (error) => {
          console.error('Error getting location', error);
          this.checkoutError.set('Failed to get your location. Please enter it manually or check permissions.');
          this.locationLoading.set(false);
        }
      );
    } else {
      this.checkoutError.set('Geolocation is not supported by your browser.');
    }
  }

  onAddressInput(event: any) {
    const query = event.target.value;
    if (this.addressSearchTimeout) clearTimeout(this.addressSearchTimeout);
    
    if (query.length < 3) {
      this.addressSuggestions.set([]);
      return;
    }

    this.isSearchingAddress.set(true);
    this.addressSearchTimeout = setTimeout(() => {
      this.http.get<any[]>(`https://nominatim.openstreetmap.org/search?q=${encodeURIComponent(query)}&format=json&addressdetails=1&limit=5`)
        .subscribe({
          next: (res: any[]) => {
            this.addressSuggestions.set(res);
            this.isSearchingAddress.set(false);
          },
          error: () => {
            this.isSearchingAddress.set(false);
            this.addressSuggestions.set([]);
          }
        });
    }, 500);
  }

  selectAddress(suggestion: any) {
    this.checkoutForm.patchValue({
      address: suggestion.display_name,
      city: suggestion.address.city || suggestion.address.town || suggestion.address.village || ''
    });
    this.addressSuggestions.set([]);
  }

  processCheckout() {
    if (this.checkoutForm.invalid || this.isProcessing()) return;

    if (this.deliveryMethod() === 'delivery') {
      if (!this.checkoutForm.value.address || !this.checkoutForm.value.city) {
        this.checkoutError.set('Address and City are required for delivery.');
        return;
      }
    }

    this.isProcessing.set(true);
    this.checkoutError.set('');

    const settings = this.settingsService.settings();
    if (this.paymentMethod() === 'online') {
      if (!settings?.enable_paystack || !settings?.paystack_public_key) {
        this.checkoutError.set('Payment gateway is not configured.');
        this.isProcessing.set(false);
        return;
      }
    }

    const fullName = `${this.checkoutForm.value.firstName || ''} ${this.checkoutForm.value.lastName || ''}`.trim();
    const branchId = this.route.snapshot.queryParamMap.get('branch_id') ||
                     this.route.snapshot.queryParamMap.get('branchId') ||
                     (typeof window !== 'undefined' ? localStorage.getItem('store_branch_id') : null);

    const payload: any = {
      branch_id: branchId || undefined,
      customer_name: fullName || undefined,
      customer_email: this.checkoutForm.value.email || undefined,
      customer_phone: this.checkoutForm.value.phone || undefined,
      total: this.finalTotal(),
      delivery_method: this.deliveryMethod(),
      delivery_address: this.deliveryMethod() === 'delivery' ? `${this.checkoutForm.value.address}, ${this.checkoutForm.value.city}` : '',
      order_notes: this.checkoutForm.value.orderNotes || '',
      promo_code: this.appliedPromoCode() || '',
      discount_amount: this.discountAmount(),
      items: this.cartService.cartItems().map(i => ({
        product_id: i.product_id,
        quantity: i.quantity
      }))
    };

    // Amount in lowest denomination (e.g. pesewas / cents)
    const amountInLowestUnit = Math.round(this.finalTotal() * 100);

    if (this.paymentMethod() === 'pickup') {
      payload['payment_method'] = 'pickup';
      this.checkoutService.processCheckout(payload).subscribe({
        next: () => {
          this.cartService.clearCart();
          this.router.navigate(['/store/order-confirmation']);
        },
        error: (err) => {
          this.checkoutError.set(err.error?.error || 'Verification failed. Please contact support.');
          this.isProcessing.set(false);
        }
      });
      return;
    }

    payload['payment_method'] = 'online';
    const setupConfig: any = {
      key: settings!.paystack_public_key,
      email: this.checkoutForm.value.email || 'customer@puxbay.com',
      amount: amountInLowestUnit,
      currency: settings?.currency || 'GHS',
      channels: ['mobile_money', 'card', 'bank', 'ussd', 'qr'],
      callback: (response: any) => {
        // Payment successful, verify with backend
        payload['reference'] = response.reference;
        this.checkoutService.processCheckout(payload).subscribe({
          next: () => {
            this.cartService.clearCart();
            this.router.navigate(['/store/order-confirmation']);
          },
          error: (err) => {
            this.checkoutError.set(err.error?.error || 'Verification failed. Please contact support.');
            this.isProcessing.set(false);
          }
        });
      },
      onClose: () => {
        this.checkoutError.set('Payment was cancelled.');
        this.isProcessing.set(false);
      }
    };

    if (settings!.paystack_subaccount_code) {
      setupConfig.subaccount = settings!.paystack_subaccount_code;
    }

    const handler = (window as any).PaystackPop.setup(setupConfig);
    handler.openIframe();
  }

  isFieldInvalid(field: string): boolean {
    const control = this.checkoutForm.get(field);
    return !!(control && control.invalid && control.touched);
  }

  get deliveryFee(): number {
    return this.settingsService.settings()?.delivery_fee || 0;
  }

  finalTotal(): number {
    const cartTotal = this.cartService.cartTotal();
    const withDelivery = this.deliveryMethod() === 'delivery' ? cartTotal + this.deliveryFee : cartTotal;
    return Math.max(0, withDelivery - this.discountAmount());
  }
}
