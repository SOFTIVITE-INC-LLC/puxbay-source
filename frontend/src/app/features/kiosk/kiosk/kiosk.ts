import { ToastService } from '../../../core/services/toast';
import { Component, OnInit, OnDestroy, computed, inject, signal, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { CatalogService } from '../../../core/services/catalog.service';
import { StorefrontService } from '../../../core/services/storefront.service';
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
  catalogService = inject(CatalogService);
  storefrontService = inject(StorefrontService);
  toastService = inject(ToastService);
  private api = inject(ApiService);

  branchId = '';
  isIdle = signal(true);
  isCartOpen = signal(false);
  isProcessing = signal(false);
  idleTimer: any;
  IDLE_TIMEOUT = 60000; // 60 seconds

  // Customer info
  isCustomerStep = signal(false);
  customerName = '';
  customerPhone = '';
  kioskCustomer = signal<any>(null);
  isRegisteringCustomer = signal(false);

  cart = signal<{product: any, qty: number}[]>([]);
  orderPlaced = signal(false);
  orderNumber = signal<string | null>(null);
  countdown = signal(5);
  countdownTimer: any;

  // Modals
  selectedProduct = signal<any>(null);
  modalQty = 1;
  isPaymentModalOpen = signal(false);

  // Categories
  categories = computed(() => this.catalogService.categories());
  activeCategoryId = signal<string | null>(null);

  products = computed(() => {
    let prods = this.catalogService.products().filter(p => p.is_active);
    if (this.activeCategoryId()) {
      prods = prods.filter(p => p.category_id === this.activeCategoryId());
    }
    return prods;
  });

  cartTotal = computed(() => 
    this.cart().reduce((sum, item) => sum + (item.qty * (item.product.selling_price || 0)), 0)
  );

  ngOnInit() {
    this.branchId = this.route.snapshot.paramMap.get('branchId') || 'default';
    
    const params = this.branchId !== 'default' ? { branch_id: this.branchId } : undefined;
    this.catalogService.getProducts(params).subscribe();
    this.catalogService.getCategories().subscribe();
    this.storefrontService.getSettings().subscribe();
    
    // Setup idle detection
    if (isPlatformBrowser(this.platformId)) {
      document.addEventListener('touchstart', this.resetIdleTimer.bind(this));
      document.addEventListener('click', this.resetIdleTimer.bind(this));
    }
    this.resetIdleTimer();
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
    this.isCustomerStep.set(false);
    this.kioskCustomer.set(null);
    this.customerName = '';
    this.customerPhone = '';
    this.isIdle.set(true);
    clearInterval(this.countdownTimer);
  }

  setCategory(id: string | null) {
    this.activeCategoryId.set(id);
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

  processPayment(method: string) {
    this.payNow(method);
  }

  payNow(method: string = 'card') {
    if (this.cart().length === 0 || this.isProcessing()) return;
    
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
