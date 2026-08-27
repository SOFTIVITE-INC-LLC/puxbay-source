import { ToastService } from '../../../core/services/toast';
import { Component, OnInit, OnDestroy, computed, inject, signal, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { ProductService } from '../../../core/store/services/product.service';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';
import { ApiService } from '../../../core/services/api.service';

@Component({
  selector: 'app-kiosk',
  standalone: true,
  imports: [CommonModule, AppCurrencyPipe, FormsModule],
  templateUrl: './kiosk.html',
})
export class Kiosk implements OnInit, OnDestroy {
  private route = inject(ActivatedRoute);
  private platformId = inject(PLATFORM_ID);
  productService = inject(ProductService);
  storefrontService = inject(StorefrontSettingsService);
  toastService = inject(ToastService);
  private api = inject(ApiService);

  branchId = '';
  isIdle = signal(true);
  isCartOpen = signal(false);
  isProcessing = signal(false);
  loading = signal(false);
  idleTimer: any;
  IDLE_TIMEOUT = 60000; // 60 seconds

  // Customer info
  isCustomerStep = signal(false);
  customerName = '';
  customerPhone = '';
  kioskCustomer = signal<any>(null);
  isRegisteringCustomer = signal(false);

  cart = signal<{ product: any, qty: number }[]>([]);
  orderPlaced = signal(false);
  orderNumber = signal<string | null>(null);
  countdown = signal(5);
  countdownTimer: any;

  // Modals
  selectedProduct = signal<any>(null);
  modalQty = 1;
  isPaymentModalOpen = signal(false);

  // Products & Categories
  productsList = signal<any[]>([]);
  categoriesList = signal<any[]>([]);
  activeCategoryId = signal<string | null>(null);
  searchQuery = signal('');

  categories = computed(() => this.categoriesList());

  products = computed(() => {
    let prods = this.productsList();
    if (this.activeCategoryId()) {
      prods = prods.filter(p => p.category_id === this.activeCategoryId() || p.category?.id === this.activeCategoryId());
    }
    const q = this.searchQuery().toLowerCase().trim();
    if (q) {
      prods = prods.filter(p =>
        p.name?.toLowerCase().includes(q) ||
        p.description?.toLowerCase().includes(q) ||
        p.sku?.toLowerCase().includes(q)
      );
    }
    return prods;
  });

  cartTotal = computed(() =>
    this.cart().reduce((sum, item) => sum + (item.qty * (item.product.selling_price || 0)), 0)
  );

  ngOnInit() {
    this.branchId = this.route.snapshot.paramMap.get('branchId') ||
      this.route.snapshot.queryParamMap.get('branch_id') ||
      this.route.snapshot.queryParamMap.get('branchId') || '';

    this.loadKioskData();

    // Setup idle detection
    if (isPlatformBrowser(this.platformId)) {
      document.addEventListener('touchstart', this.resetIdleTimer.bind(this));
      document.addEventListener('click', this.resetIdleTimer.bind(this));
    }
    this.resetIdleTimer();
  }

  loadKioskData() {
    this.loading.set(true);
    const params: any = { page_size: 10000 };
    if (this.branchId && this.branchId !== 'default') {
      params.branch_id = this.branchId;
    }

    this.productService.getProducts(params).subscribe({
      next: (res) => {
        this.productsList.set(res.products || []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });

    const bId = this.branchId && this.branchId !== 'default' ? this.branchId : undefined;
    this.productService.getCategories(bId).subscribe({
      next: (cats) => this.categoriesList.set(cats || [])
    });

    this.storefrontService.loadSettings().subscribe();

    // Load Paystack inline.js if not already loaded
    if (typeof document !== 'undefined' && !document.getElementById('paystack-inline-js')) {
      const script = document.createElement('script');
      script.id = 'paystack-inline-js';
      script.src = 'https://js.paystack.co/v1/inline.js';
      document.head.appendChild(script);
    }
  }

  tenantCurrency = computed(() => {
    return this.storefrontService.settings()?.currency || 'GHS';
  });

  getCategoryName(id: string | null): string {
    if (!id) return 'Our Menu';
    const cat = this.categoriesList().find(c => c.id === id);
    return cat ? cat.name : 'Our Menu';
  }

  ngOnDestroy() {
    if (isPlatformBrowser(this.platformId)) {
      document.removeEventListener('touchstart', this.resetIdleTimer.bind(this));
      document.removeEventListener('click', this.resetIdleTimer.bind(this));
    }
    clearTimeout(this.idleTimer);
    clearInterval(this.countdownTimer);
  }

  resetIdleTimer() {
    clearTimeout(this.idleTimer);
    if (!this.isIdle()) {
      this.idleTimer = setTimeout(() => {
        this.startOver();
        this.isIdle.set(true);
      }, this.IDLE_TIMEOUT);
    }
  }

  startSession() {
    this.isIdle.set(false);
    this.isCustomerStep.set(true);
    this.resetIdleTimer();
  }

  skipCustomerStep() {
    this.kioskCustomer.set(null);
    this.isCustomerStep.set(false);
    this.resetIdleTimer();
  }

  registerCustomer() {
    if (!this.customerName.trim()) {
      this.isCustomerStep.set(false);
      return;
    }
    this.isRegisteringCustomer.set(true);
    this.api.post<any>('/kiosk/customers', {
      name: this.customerName.trim(),
      phone: this.customerPhone.trim() || undefined
    }).subscribe({
      next: (customer) => {
        this.kioskCustomer.set(customer);
        this.isRegisteringCustomer.set(false);
        this.isCustomerStep.set(false);
        this.resetIdleTimer();
      },
      error: () => {
        this.isRegisteringCustomer.set(false);
        this.isCustomerStep.set(false);
        this.resetIdleTimer();
      }
    });
  }

  startOver() {
    this.cart.set([]);
    this.orderPlaced.set(false);
    this.orderNumber.set(null);
    this.isPaymentModalOpen.set(false);
    this.selectedProduct.set(null);
    this.activeCategoryId.set(null);
    this.searchQuery.set('');
    this.isCustomerStep.set(false);
    this.kioskCustomer.set(null);
    this.customerName = '';
    this.customerPhone = '';
    this.isProcessing.set(false);
    this.isIdle.set(true);
    clearInterval(this.countdownTimer);
  }

  setCategory(id: string | null) {
    this.activeCategoryId.set(id);
    this.searchQuery.set('');
    this.resetIdleTimer();
  }

  openProductModal(product: any) {
    this.selectedProduct.set(product);
    this.modalQty = 1;
    this.resetIdleTimer();
  }

  closeProductModal() {
    this.selectedProduct.set(null);
    this.resetIdleTimer();
  }

  confirmAddToCart(product: any, qty: number) {
    const existing = this.cart().find(i => i.product.id === product.id);
    if (existing) {
      this.cart.update(c => c.map(i => i.product.id === product.id ? { ...i, qty: i.qty + qty } : i));
    } else {
      this.cart.update(c => [...c, { product, qty }]);
    }
    this.closeProductModal();
  }

  addToCart(product: any) {
    this.openProductModal(product);
  }

  increaseQuantity(productId: string) {
    this.cart.update(c => c.map(i => i.product.id === productId ? { ...i, qty: i.qty + 1 } : i));
    this.resetIdleTimer();
  }

  decreaseQuantity(productId: string) {
    this.cart.update(c => c.map(i => {
      if (i.product.id === productId && i.qty > 1) {
        return { ...i, qty: i.qty - 1 };
      }
      return i;
    }));
    this.resetIdleTimer();
  }

  removeFromCart(productId: string) {
    this.cart.update(c => c.filter(i => i.product.id !== productId));
    this.resetIdleTimer();
  }

  clearCart() {
    this.cart.set([]);
    this.isCartOpen.set(false);
    this.resetIdleTimer();
  }

  openPaymentModal() {
    if (this.cart().length === 0 || this.isProcessing()) return;
    this.isPaymentModalOpen.set(true);
    this.resetIdleTimer();
  }

  closePaymentModal() {
    this.isPaymentModalOpen.set(false);
    this.resetIdleTimer();
  }

  processPayment(method: 'card' | 'cash' | 'mobile_money') {
    const settings = this.storefrontService.settings();
    const hasPaystack = !!(settings?.paystack_public_key && (settings?.enable_paystack !== false));

    if ((method === 'mobile_money' || method === 'card') && hasPaystack) {
      this.payWithPaystack(method, settings);
    } else {
      this.payNow(method);
    }
  }

  payWithPaystack(method: 'card' | 'mobile_money', settings: any) {
    if (this.cart().length === 0 || this.isProcessing()) return;
    this.isProcessing.set(true);

    const total = this.cartTotal();
    const amountInLowestUnit = Math.round(total * 100);
    const currency = this.tenantCurrency();
    const customer = this.kioskCustomer();
    const cleanPhone = (this.customerPhone || '').replace(/[^0-9]/g, '');
    const customerEmail = (customer?.email && customer.email.includes('@') && customer.email.includes('.'))
      ? customer.email
      : (cleanPhone ? `kiosk-${cleanPhone}@puxbay.com` : 'kiosk-customer@puxbay.com');

    const channels = method === 'mobile_money'
      ? ['mobile_money', 'qr', 'ussd']
      : ['card', 'mobile_money', 'bank', 'ussd', 'qr'];

    const setupConfig: any = {
      key: settings.paystack_public_key,
      email: customerEmail,
      amount: amountInLowestUnit,
      currency: currency,
      channels: channels,
      callback: (response: any) => {
        this.payNow(method, response.reference);
      },
      onClose: () => {
        this.isProcessing.set(false);
        this.toastService.showInfo('Payment was cancelled.');
      }
    };

    if (settings.paystack_subaccount_code) {
      setupConfig.subaccount = settings.paystack_subaccount_code;
    }

    if (typeof window !== 'undefined' && (window as any).PaystackPop) {
      const handler = (window as any).PaystackPop.setup(setupConfig);
      handler.openIframe();
    } else {
      // If Paystack script is not loaded, fallback to direct order creation
      this.payNow(method);
    }
  }

  payNow(method: string = 'card', paymentRef?: string) {
    if (this.cart().length === 0) return;
    this.isProcessing.set(true);

    const items = this.cart().map(item => ({
      product_id: item.product.id,
      quantity: item.qty,
      unit_price: item.product.selling_price || 0,
      discount: 0,
      total: item.qty * (item.product.selling_price || 0)
    }));

    const total = this.cartTotal();

    const body = {
      branch_id: this.branchId !== 'default' ? this.branchId : undefined,
      customer_id: this.kioskCustomer()?.id ?? undefined,
      subtotal: total,
      tax: 0,
      discount: 0,
      total,
      amount_paid: total,
      payment_method: method,
      payment_ref: paymentRef,
      items
    };

    this.api.post<any>('/kiosk/orders', body).subscribe({
      next: (order) => {
        this.isProcessing.set(false);
        this.isPaymentModalOpen.set(false);
        this.orderPlaced.set(true);
        this.orderNumber.set(order.receipt_number || order.id?.substring(0, 8) || '101');

        // Automatically print receipt via hidden iframe
        if (order.receipt_token) {
          this.api.get(`/public/receipts/${order.receipt_token}`, {
            headers: { 'Accept': 'text/html' },
            responseType: 'text'
          } as any).subscribe({
            next: (html: any) => {
              const iframe = document.createElement('iframe');
              iframe.style.position = 'absolute';
              iframe.style.width = '0px';
              iframe.style.height = '0px';
              iframe.style.border = 'none';
              document.body.appendChild(iframe);

              const doc = iframe.contentWindow?.document;
              if (doc) {
                doc.open();
                doc.write(html);
                doc.close();

                setTimeout(() => {
                  iframe.contentWindow?.focus();
                  iframe.contentWindow?.print();
                  setTimeout(() => document.body.removeChild(iframe), 1000);
                }, 500);
              }
            }
          });
        }

        this.cart.set([]);
        this.isCartOpen.set(false);

        this.countdown.set(5);
        this.countdownTimer = setInterval(() => {
          this.countdown.update(c => c - 1);
          if (this.countdown() <= 0) {
            clearInterval(this.countdownTimer);
            this.startOver();
          }
        }, 1000);
      },
      error: (err) => {
        this.isProcessing.set(false);
        this.toastService.showError('Failed to process order. Please try again or ask for assistance.');
        console.error('Kiosk checkout error:', err);
      }
    });
  }
}
