import { Injectable, inject, signal, computed, effect } from '@angular/core';
import { CatalogService } from '../../core/services/catalog.service';
import { OrderService } from '../../core/services/order.service';
import { CustomerService } from '../../core/services/customer.service';
import { CategoryService } from '../../core/services/category.service';
import { SettingsService } from '../../core/services/settings.service';
import { StorefrontSettingsService } from '../../core/store/services/storefront-settings.service';
import { Customer } from '../../core/models/models';
import { Product } from '../../core/models/product.models';
import { ToastrService } from 'ngx-toastr';
import { OfflineDbService } from '../../core/services/offline-db.service';
import { PrinterService } from '../../core/services/printer.service';
import { GiftCardService } from '../../core/services/gift-card.service';

@Injectable({ providedIn: 'root' })
export class PosFacade {
  private catalogService = inject(CatalogService);
  private orderService = inject(OrderService);
  private customerService = inject(CustomerService);
  private categoryService = inject(CategoryService);
  private toastr = inject(ToastrService);
  private settingsService = inject(SettingsService);
  readonly offlineDb = inject(OfflineDbService);
  readonly printer = inject(PrinterService);
  private giftCardService = inject(GiftCardService);
  private storefrontSettings = inject(StorefrontSettingsService);

  // --- CORE STATE ---
  searchQuery = signal('');
  selectedCategoryId = signal<string | null>('all');
  cart = signal<{ product: Product, quantity: number, discount: number, discountType: 'fixed' | 'percent' }[]>([]);
  payments = signal<{ method: string, amount: number, code?: string }[]>([]);
  selectedCustomer = signal<Customer | null>(null);

  // --- UI STATE ---
  showCustomerDropdown = signal(false);
  isCheckoutModalOpen = signal(false);
  isSuccessModalOpen = signal(false);
  isMobileCartOpen = signal(false);
  isParkedSalesModalOpen = signal(false);
  isHardwareModalOpen = signal(false);
  isShiftModalOpen = signal(false);
  isCustomItemModalOpen = signal(false);
  isVoidModalOpen = signal(false);
  checkoutSuccessOrder = signal<any | null>(null);
  theme = signal<'light' | 'dark'>('light');
  isOffline = signal(!navigator.onLine);
  isFullscreen = signal(false);
  isCheckoutLoading = signal(false);
  paymentAmountInput = signal<number | null>(null);

  // QR Pay
  isQRModalOpen = signal(false);
  qrPaymentData = signal<{ reference: string; amount: number; qrUrl: string } | null>(null);
  private qrPollTimer: any = null;

  // --- ADVANCED STATE ---
  parkedSales = signal<{ cart: any[], customer: any, time: Date }[]>([]);
  isPrinterConnected = signal(false);
  printerPort: any = null;
  shiftStatus = signal<'open' | 'closed'>('closed');
  shiftDetails = signal<any>(null);

  // Kitchen Dispatch
  isSentToKitchen = signal(false);

  // Loyalty
  loyaltyPoints = signal(0);
  pointsRedeemed = signal(0);
  isAutoPromoEnabled = signal(true);

  // Broadcast Channel for CDS
  private bc = new BroadcastChannel('pos_sync_channel');

  // --- DATA ---
  customers = this.customerService.customers;
  categories = this.categoryService.categories;

  constructor() {
    window.addEventListener('online', () => this.isOffline.set(false));
    window.addEventListener('offline', () => this.isOffline.set(true));

    if (document.documentElement.classList.contains('dark')) {
      this.theme.set('dark');
    }

    document.addEventListener('fullscreenchange', () => {
      this.isFullscreen.set(!!document.fullscreenElement);
    });

    // Listen for CDS requests
    this.bc.onmessage = (event) => {
      if (event.data?.type === 'request_sync') {
        this.syncCDS();
      }
    };

    // Auto-sync CDS on cart changes
    effect(() => {
      this.cart(); // trigger on cart change
      this.isSuccessModalOpen(); // trigger on success screen
      this.syncCDS();
    });
  }

  init() {
    this.catalogService.getProducts({ limit: -1 }).subscribe(res => {
      // Cache products for offline use
      if (navigator.onLine && res?.data?.length) {
        this.offlineDb.cacheProducts(res.data.map((p: any) => ({
          id: p.id, name: p.name, sku: p.sku,
          selling_price: p.selling_price, cost_price: p.cost_price,
          current_stock: p.current_stock, stock_unit: p.stock_unit,
          image_url: p.image_url, category: p.category,
          track_inventory: p.track_inventory, is_active: p.is_active,
          branch_id: p.branch_id, cached_at: Date.now(),
        }))).catch(() => { });
      }
    });
    this.customerService.getCustomers({ limit: -1 }).subscribe();
    this.categoryService.getCategories({ limit: -1 }).subscribe();
    this.settingsService.getSettings().subscribe();
    // Load storefront settings (Paystack public key + subaccount code)
    this.storefrontSettings.loadSettings().subscribe();
    // Load Paystack inline.js for Mobile Money / card payments
    if (typeof document !== 'undefined' && !document.getElementById('paystack-inline-js')) {
      const s = document.createElement('script');
      s.id = 'paystack-inline-js';
      s.src = 'https://js.paystack.co/v1/inline.js';
      document.head.appendChild(s);
    }
    // On reconnect, attempt to sync any queued offline orders
    window.addEventListener('online', () => this._syncOfflineOrders(), { once: false });
  }

  // --- COMPUTEDS ---
  readonly recommendations = computed(() => this.catalogService.products().slice(0, 8));

  readonly filteredProducts = computed(() => {
    const q = this.searchQuery().toLowerCase();
    const catId = this.selectedCategoryId();
    return this.catalogService.products().filter(p => {
      if (!p.is_active) return false;
      const matchesSearch = !q || p.name.toLowerCase().includes(q) || (p.barcode && p.barcode.toLowerCase().includes(q)) || (p.sku && p.sku.toLowerCase().includes(q));
      const matchesCategory = !catId || catId === 'all' || p.category_id === catId;
      return matchesSearch && matchesCategory;
    });
  });

  readonly cartSubtotal = computed(() => {
    return this.cart().reduce((sum, item) => sum + ((item.product.selling_price || 0) * item.quantity), 0);
  });

  readonly rawDiscountTotal = computed(() => {
    return this.cart().reduce((sum, item) => {
      if (item.discountType === 'percent') {
        return sum + (((item.product.selling_price || 0) * item.quantity) * (item.discount / 100));
      }
      return sum + item.discount;
    }, 0);
  });

  // PROMO ENGINE: Dynamic settings
  readonly promoThreshold = computed(() => this.settingsService.settings()?.promo_threshold || 100);
  readonly promoPercent = computed(() => this.settingsService.settings()?.promo_discount_percent || 10);
  readonly autoPromoDiscount = computed(() => {
    if (!this.isAutoPromoEnabled()) return 0;
    const threshold = this.promoThreshold();
    const percent = this.promoPercent();
    return (this.cartSubtotal() > threshold && percent > 0) ? this.cartSubtotal() * (percent / 100) : 0;
  });

  readonly cartDiscountTotal = computed(() => this.rawDiscountTotal() + this.autoPromoDiscount() + this.pointsRedeemed());

  readonly cartTax = computed(() => {
    return this.cart().reduce((sum, item) => {
      let lineTotal = (item.product.selling_price || 0) * item.quantity;
      let lineDiscount = item.discountType === 'percent' ? lineTotal * (item.discount / 100) : item.discount;
      const taxable = Math.max(0, lineTotal - lineDiscount);
      return sum + (taxable * (item.product.tax_rate || 0) / 100);
    }, 0);
  });

  readonly cartTotal = computed(() => Math.max(0, this.cartSubtotal() - this.cartDiscountTotal() + this.cartTax()));

  readonly amountPaid = computed(() => this.payments().reduce((sum, p) => sum + p.amount, 0));
  readonly remainingBalance = computed(() => Math.max(0, this.cartTotal() - this.amountPaid()));
  readonly changeDue = computed(() => Math.max(0, this.amountPaid() - this.cartTotal()));

  readonly quickCashOptions = computed(() => {
    const total = this.remainingBalance();
    if (total === 0) return [];
    const options = [total];
    const rounded5 = Math.ceil(total / 5) * 5;
    if (rounded5 > total && rounded5 - total < 5) options.push(rounded5);
    const rounded10 = Math.ceil(total / 10) * 10;
    if (rounded10 > total && !options.includes(rounded10)) options.push(rounded10);
    const rounded50 = Math.ceil(total / 50) * 50;
    if (rounded50 > total && !options.includes(rounded50)) options.push(rounded50);
    const rounded100 = Math.ceil(total / 100) * 100;
    if (rounded100 > total && !options.includes(rounded100)) options.push(rounded100);
    return options.sort((a, b) => a - b);
  });

  // --- ACTIONS ---
  toggleTheme() {
    if (this.theme() === 'light') {
      this.theme.set('dark');
      document.documentElement.classList.add('dark');
    } else {
      this.theme.set('light');
      document.documentElement.classList.remove('dark');
    }
  }

  toggleFullscreen() {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch(() => { });
    } else if (document.exitFullscreen) {
      document.exitFullscreen();
    }
  }

  syncCDS() {
    this.bc.postMessage({
      cart: this.cart(),
      subtotal: this.cartSubtotal(),
      tax: this.cartTax(),
      discount: this.cartDiscountTotal(),
      total: this.cartTotal(),
      isSuccess: this.isSuccessModalOpen()
    });
  }

  openCDSWindow() {
    window.open('/cds', 'CDS', 'width=1024,height=768');
  }

  scanBarcode(barcode: string) {
    if (!barcode.trim()) return;
    const product = this.catalogService.products().find(p => p.barcode === barcode || p.sku === barcode);
    if (product) {
      this.addToCart(product);
      this.toastr.success(`Added ${product.name}`);
      this.searchQuery.set('');
    } else {
      this.toastr.warning(`Unknown barcode: ${barcode}`);
    }
  }

  addToCart(product: Product) {
    if (this.isSentToKitchen()) {
      this.toastr.warning('Cannot modify dispatched order. Please void first.');
      return;
    }
    this.cart.update(items => {
      const existing = items.find(i => i.product.id === product.id);
      if (existing) {
        return items.map(i => i.product.id === product.id ? { ...i, quantity: i.quantity + 1 } : i);
      }
      return [...items, { product, quantity: 1, discount: 0, discountType: 'fixed' }];
    });
  }

  addCustomItem(name: string, price: number) {
    if (this.isSentToKitchen()) return;
    if (!name || price <= 0) return;
    const product: Product = {
      id: `custom_${Date.now()}`,
      name: name,
      selling_price: price,
      cost_price: 0,
      stock: 999,
      is_active: true,
      category_id: 'custom'
    } as any;
    this.addToCart(product);
    this.isCustomItemModalOpen.set(false);
  }

  updateQuantity(id: string, amount: number) {
    if (this.isSentToKitchen()) return;
    this.cart.update(items => items.map(i => {
      if (i.product.id === id) {
        return { ...i, quantity: Math.max(1, i.quantity + amount) };
      }
      return i;
    }));
  }

  removeFromCart(id: string) {
    if (this.isSentToKitchen()) return;
    this.cart.update(items => items.filter(i => i.product.id !== id));
  }

  clearCart() {
    if (this.isSentToKitchen()) {
      // Require void pin
      this.voidPin.set(''); // reset PIN when opening modal
      this.isVoidModalOpen.set(true);
      return;
    }
    this.resetCartState();
  }

  private resetCartState() {
    this.cart.set([]);
    this.payments.set([]);
    this.paymentAmountInput.set(null);
    this.selectedCustomer.set(null);
    this.isSentToKitchen.set(false);
    this.pointsRedeemed.set(0);
    this.isAutoPromoEnabled.set(true);
  }

  voidPin = signal('');

  appendVoidPin(digit: string) {
    if (this.voidPin().length < 8) {
      this.voidPin.update(p => p + digit);
    }
  }

  backspaceVoidPin() {
    this.voidPin.update(p => p.slice(0, -1));
  }

  verifyVoid() {
    if (this.voidPin() === '1234') {
      this.toastr.success('Manager override approved. Cart voided.');
      this.resetCartState();
      this.isVoidModalOpen.set(false);
    } else {
      this.toastr.error('Invalid Manager PIN');
      this.voidPin.set('');
    }
  }

  selectCustomer(c: Customer | null) {
    this.selectedCustomer.set(c);
    this.showCustomerDropdown.set(false);
    // Mock loyalty logic: load points
    this.loyaltyPoints.set(c ? Math.floor(Math.random() * 50) + 10 : 0);
  }

  redeemLoyaltyPoints() {
    if (this.loyaltyPoints() > 0) {
      const discountVal = this.loyaltyPoints();
      this.pointsRedeemed.set(discountVal);
      this.loyaltyPoints.set(0);
      this.toastr.success(`Redeemed ${discountVal} points for $${discountVal} off!`);
    }
  }

  sendToKitchen() {
    if (this.cart().length === 0) return;
    this.isSentToKitchen.set(true);
    this.toastr.success('Order dispatched to Kitchen Display!');
  }

  // --- CHECKOUT ---
  openCheckout() {
    if (this.cart().length === 0) return;
    this.isCheckoutModalOpen.set(true);
  }

  closeCheckout() {
    this.isCheckoutModalOpen.set(false);
    this.isSuccessModalOpen.set(false);
    this.payments.set([]);
    this.paymentAmountInput.set(null);
  }

  addPaymentMethod(method: string, amount?: number, code?: string) {
    let amt = amount;
    if (amt === undefined) {
      // If paymentAmountInput is populated, use it. Otherwise, use remaining balance.
      amt = this.paymentAmountInput() || this.remainingBalance();
    }
    
    // Cap at remaining balance (but we could allow overpayment if we want change, but let's just cap for now)
    if (amt > this.remainingBalance()) {
      amt = this.remainingBalance();
    }

    if (amt <= 0) return;
    
    const existing = this.payments().find(p => p.method === method);
    if (existing && !code) {
      this.payments.update(p => p.map(x => x.method === method ? { ...x, amount: x.amount + amt } : x));
    } else {
      this.payments.update(p => [...p, { method, amount: amt, code }]);
    }
    this.paymentAmountInput.set(null); // Clear input after adding
  }

  removePaymentMethod(index: number) {
    this.payments.update(p => p.filter((_, i) => i !== index));
  }

  /**
   * Opens the Paystack popup for the remaining balance.
   * Supports MTN MoMo, Telecel Cash, AirtelTigo, Visa & Mastercard.
   * Payments are routed directly to the tenant's Paystack subaccount if configured.
   */
  paystackMobileCheckout() {
    const remaining = this.remainingBalance();
    if (remaining <= 0) {
      this.toastr.info('No remaining balance to charge.');
      return;
    }

    const settings = this.storefrontSettings.settings();
    const pubKey = settings?.paystack_public_key;
    if (!pubKey) {
      // Fallback: just log as a regular mobile money payment (offline / no key configured)
      this.addPaymentMethod('mobile');
      this.toastr.warning('Paystack not configured — logged as Mobile Money.');
      return;
    }

    // We need an email to initialise Paystack; use customer email or a placeholder
    const email = this.selectedCustomer()?.email || 'pos-sale@puxbay.com';
    const amountInPesewas = Math.round(remaining * 100);

    const setupConfig: any = {
      key: pubKey,
      email,
      amount: amountInPesewas,
      currency: 'GHS',
      channels: ['mobile_money', 'card', 'bank', 'ussd', 'qr'],
      callback: (response: any) => {
        // Payment successful — record it as 'mobile' with the Paystack reference
        this.payments.update(p => [
          ...p,
          { method: 'mobile', amount: remaining, code: response.reference }
        ]);
        this.toastr.success(`Mobile Money payment confirmed (${response.reference})`);
        // Auto-complete sale if fully paid
        if (this.remainingBalance() <= 0) {
          this.processCheckout();
        }
      },
      onClose: () => {
        this.toastr.info('Mobile Money payment cancelled.');
      }
    };

    if (settings?.paystack_subaccount_code) {
      setupConfig.subaccount = settings.paystack_subaccount_code;
      setupConfig.bearer = 'subaccount';
    }

    if ((window as any).PaystackPop) {
      const handler = (window as any).PaystackPop.setup(setupConfig);
      handler.openIframe();
    } else {
      this.toastr.error('Paystack script is still loading. Please try again.');
    }
  }

  /** Opens Paystack popup restricted to QR channel only. Customer scans the QR on their phone. */
  paystackQRCheckout() {
    const remaining = this.remainingBalance();
    if (remaining <= 0) {
      this.toastr.info('No remaining balance to charge.');
      return;
    }

    const settings = this.storefrontSettings.settings();
    const pubKey = settings?.paystack_public_key;
    if (!pubKey) {
      this.toastr.warning('Paystack not configured — QR Pay unavailable.');
      return;
    }

    const email = this.selectedCustomer()?.email || 'pos-qr@puxbay.com';
    const amountInPesewas = Math.round(remaining * 100);

    const setupConfig: any = {
      key: pubKey,
      email,
      amount: amountInPesewas,
      currency: 'GHS',
      channels: ['qr'],
      callback: (response: any) => {
        this.payments.update(p => [
          ...p,
          { method: 'qr', amount: remaining, code: response.reference }
        ]);
        this.toastr.success(`QR payment confirmed (${response.reference})`);
        if (this.remainingBalance() <= 0) {
          this.processCheckout();
        }
      },
      onClose: () => {
        this.toastr.info('QR payment cancelled.');
      }
    };

    if (settings?.paystack_subaccount_code) {
      setupConfig.subaccount = settings.paystack_subaccount_code;
      setupConfig.bearer = 'subaccount';
    }

    if ((window as any).PaystackPop) {
      const handler = (window as any).PaystackPop.setup(setupConfig);
      handler.openIframe();
    } else {
      this.toastr.error('Paystack script is still loading. Please try again.');
    }
  }

  openQRModal() {
    const remaining = this.remainingBalance();
    if (remaining <= 0) {
      this.toastr.info('No remaining balance to charge.');
      return;
    }
    const settings = this.storefrontSettings.settings();
    const pubKey = settings?.paystack_public_key;
    if (!pubKey) {
      this.toastr.warning('Paystack not configured — QR Pay unavailable.');
      return;
    }
    // Generate a unique reference for this transaction
    const reference = `POS-QR-${Date.now()}-${Math.random().toString(36).substring(2, 7).toUpperCase()}`;
    // Build a Paystack payment page URL as QR content
    // The customer scans this and pays via their phone
    const email = this.selectedCustomer()?.email || 'pos-qr@puxbay.com';
    const amountInPesewas = Math.round(remaining * 100);
    // Encode as Paystack inline URL so the popup opens on the customer device
    const qrUrl = `https://paystack.com/pay/?key=${encodeURIComponent(pubKey)}&email=${encodeURIComponent(email)}&amount=${amountInPesewas}&currency=GHS&ref=${reference}${
      settings?.paystack_subaccount_code ? `&subaccount=${settings.paystack_subaccount_code}&bearer=subaccount` : ''
    }`;
    this.qrPaymentData.set({ reference, amount: remaining, qrUrl });
    this.isQRModalOpen.set(true);
    // Poll for payment every 4 seconds (max 75 polls = 5 mins)
    let polls = 0;
    this.qrPollTimer = setInterval(() => {
      polls++;
      if (polls > 75) {
        clearInterval(this.qrPollTimer);
        this.toastr.warning('QR payment session expired.');
        this.isQRModalOpen.set(false);
        return;
      }
      // Check Paystack transaction status via backend verify endpoint
      fetch(`/api/v1/pos/verify-payment?reference=${reference}`, {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('access_token') || ''}` }
      })
        .then(r => r.json())
        .then(data => {
          if (data?.status === 'success' || data?.data?.status === 'success') {
            clearInterval(this.qrPollTimer);
            this.isQRModalOpen.set(false);
            this.payments.update(p => [
              ...p,
              { method: 'qr', amount: remaining, code: reference }
            ]);
            this.toastr.success(`QR payment confirmed!`);
            if (this.remainingBalance() <= 0) {
              this.processCheckout();
            }
          }
        })
        .catch(() => {}); // silently ignore network errors
    }, 4000);
  }

  closeQRModal() {
    if (this.qrPollTimer) {
      clearInterval(this.qrPollTimer);
      this.qrPollTimer = null;
    }
    this.isQRModalOpen.set(false);
    this.qrPaymentData.set(null);
  }


  processCheckout() {
    if (this.amountPaid() < this.cartTotal()) {
      this.toastr.error('Insufficient payment amount');
      return;
    }

    const orderItems = this.cart().map(item => {
      let lineTotal = (item.product.selling_price || 0) * item.quantity;
      let lineDiscount = item.discountType === 'percent' ? lineTotal * (item.discount / 100) : item.discount;
      return {
        product_id: item.product.id,
        quantity: item.quantity,
        unit_price: item.product.selling_price || 0,
        discount: lineDiscount,
        tax: Math.max(0, lineTotal - lineDiscount) * ((item.product.tax_rate || 0) / 100),
        total: Math.max(0, lineTotal - lineDiscount) * (1 + (item.product.tax_rate || 0) / 100),
        cost_price: item.product.cost_price || 0
      };
    });

    const newOrder = {
      customer_id: this.selectedCustomer()?.id,
      subtotal: this.cartSubtotal(),
      tax: this.cartTax(),
      discount: this.cartDiscountTotal(),
      total: this.cartTotal(),
      amount_paid: this.amountPaid(),
      status: 'completed',
      payment_status: 'paid',
      payments: this.payments(),
      order_type: 'pos',
      items: orderItems
    };

    if (this.isOffline()) {
      // Queue the order for later sync via IndexedDB
      this.offlineDb.queueOrder(newOrder).then(localId => {
        this.toastr.warning(`Offline Mode — Order #${localId} saved locally. It will sync automatically when you reconnect.`, 'Queued', { timeOut: 5000 });
      });
      this.resetCartState();
      this.closeCheckout();
      return;
    }

    this.isCheckoutLoading.set(true);
    this.orderService.createOrder(newOrder as any).subscribe({
      next: (res) => {
        this.isCheckoutLoading.set(false);
        this.toastr.success(`Order completed! #${res.order_number}`);
        // Enrich response with product names since backend createOrder doesn't preload Product relation
        if (res && res.items) {
          res.items = res.items.map((rItem: any) => {
            const cartItem = this.cart().find(c => c.product.id === rItem.product_id);
            if (cartItem) {
              rItem.product_name = cartItem.product.name;
            }
            return rItem;
          });
        }
        this.resetCartState();
        this.closeCheckout();
        this.checkoutSuccessOrder.set(res);
        this.isSuccessModalOpen.set(true);

        // Redeem any gift cards that were applied
        const gcPayments = this.payments().filter(p => p.method === 'gift_card' && p.code);
        gcPayments.forEach(p => {
          this.giftCardService.redeemCard(p.code!, p.amount).subscribe({
            error: (err) => console.error('Failed to redeem gift card after sale', err)
          });
        });

        // Trigger sync of any queued offline orders now that we're back online
        this._syncOfflineOrders();
      },
      error: () => {
        this.isCheckoutLoading.set(false);
        this.toastr.error(`Failed to process order.`);
      }
    });
  }

  /** Sync queued offline orders to the backend */
  async _syncOfflineOrders(): Promise<void> {
    const result = await this.offlineDb.syncPendingOrders(async (payload) => {
      return new Promise((resolve, reject) => {
        this.orderService.createOrder(payload).subscribe({ next: resolve, error: reject });
      });
    });
    if (result.synced > 0) {
      this.toastr.success(`${result.synced} offline order(s) synced successfully!`, 'Sync Complete');
    }
    if (result.failed > 0) {
      this.toastr.warning(`${result.failed} order(s) could not be synced. Will retry later.`, 'Partial Sync');
    }
  }

  // --- PARKED SALES ---
  parkSale() {
    if (this.cart().length === 0) return;
    this.parkedSales.update(sales => [
      ...sales,
      { cart: [...this.cart()], customer: this.selectedCustomer(), time: new Date() }
    ]);
    this.toastr.info('Sale parked successfully');
    this.resetCartState();
  }

  resumeSale(index: number) {
    const sale = this.parkedSales()[index];
    if (sale) {
      this.cart.set(sale.cart);
      this.selectedCustomer.set(sale.customer);
      this.removeParkedSale(index);
      this.isParkedSalesModalOpen.set(false);
    }
  }

  removeParkedSale(index: number) {
    this.parkedSales.update(sales => sales.filter((_, i) => i !== index));
  }

  // --- SHIFTS ---
  toggleShift() {
    if (this.shiftStatus() === 'closed') {
      this.shiftStatus.set('open');
      this.shiftDetails.set({ startTime: new Date(), float: 100 });
      this.toastr.success('Shift Opened');
    } else {
      this.shiftStatus.set('closed');
      this.shiftDetails.set(null);
      this.toastr.success('Shift Closed');
    }
    this.isShiftModalOpen.set(false);
  }

  // --- HARDWARE ---
  async connectPrinter() {
    await this.printer.connect();
    if (this.printer.isConnected()) {
      this.toastr.success('Receipt printer connected');
    }
  }

  async printReceipt(order: any = null) {
    const receiptOrder = order || this.checkoutSuccessOrder();
    if (!receiptOrder) return;

    const token = localStorage.getItem('auth_token');
    const tenant = window.location.hostname.split('.')[0] === 'localhost' ? 'thinkce' : window.location.hostname.split('.')[0];
    const fallbackUrl = `/api/v1/orders/${receiptOrder.id}/receipt?token=${token}&tenant=${tenant}`;

    const pReceipt = {
      storeName: 'Softivite POS',
      orderNumber: receiptOrder.order_number || receiptOrder.id.substring(0, 8),
      date: new Date(receiptOrder.created_at || Date.now()),
      items: (receiptOrder.items || []).map((i: any) => ({
        name: i.product_name || i.product?.name || 'Item',
        qty: i.quantity,
        price: i.unit_price,
        total: i.total
      })),
      subtotal: receiptOrder.subtotal,
      tax: receiptOrder.tax || 0,
      discount: receiptOrder.discount || 0,
      total: receiptOrder.total,
      amountPaid: receiptOrder.amount_paid,
      paymentMethod: receiptOrder.payment_method || (receiptOrder.payments && receiptOrder.payments.length ? receiptOrder.payments[0].method : 'cash'),
      change: Math.max(0, receiptOrder.amount_paid - receiptOrder.total),
    };

    await this.printer.printReceipt(pReceipt as any, fallbackUrl);
  }
}
