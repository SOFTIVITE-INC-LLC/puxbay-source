import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { ActivatedRoute } from '@angular/router';
import { ProductService } from '../../../core/store/services/product.service';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';

@Component({
  selector: 'app-public-storefront',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './public-storefront.html',
})
export class PublicStorefront implements OnInit {
  private route = inject(ActivatedRoute);
  private http = inject(HttpClient);
  productService = inject(ProductService);
  storefrontService = inject(StorefrontSettingsService);

  slug = signal('');
  productsList = signal<any[]>([]);
  cart = signal<{product: any, qty: number}[]>([]);
  isCartOpen = signal(false);
  isCheckoutOpen = signal(false);
  checkoutStep = signal(1);
  paymentMethod = signal<'online' | 'cash'>('online');
  isProcessing = signal(false);
  checkoutError = signal('');

  darkMode = signal(false);
  quickViewProduct = signal<any>(null);

  searchQuery = signal('');
  selectedCategory = signal('All');
  
  // Advanced Filtering
  priceRange = signal<{min: number, max: number | null}>({min: 0, max: null});
  inStockOnly = signal(false);
  showWishlistOnly = signal(false);
  sortBy = signal('newest'); // newest, price_low_high, price_high_low

  // Phase 3 States
  wishlist = signal<Set<string>>(new Set());
  recentlyViewedIds = signal<string[]>([]);

  guestDetails = signal({ name: '', email: '', phone: '', address: '' });
  orderPlaced = signal(false);
  orderTrackingNumber = signal('');

  categories = computed(() => {
    const cats = new Set(this.productsList().map(p => p.category?.name || p.category || ''));
    return ['All', ...Array.from(cats).filter(c => !!c)];
  });

  filteredProducts = computed(() => {
    let p = [...this.productsList()].filter(prod => prod.is_active !== false);
    
    // Apply text search
    const q = this.searchQuery().toLowerCase();
    if (q) p = p.filter(prod => prod.name?.toLowerCase().includes(q) || prod.sku?.toLowerCase().includes(q));
    
    // Apply category
    const c = this.selectedCategory();
    if (c !== 'All') p = p.filter(prod => (prod.category?.name || prod.category) === c);
    
    // Apply stock filter
    if (this.inStockOnly()) p = p.filter(prod => (prod.current_stock || 0) > 0);
    
    // Apply wishlist filter
    if (this.showWishlistOnly()) p = p.filter(prod => this.wishlist().has(prod.id));
    
    // Apply price filter
    const minP = this.priceRange().min;
    const maxP = this.priceRange().max;
    if (minP > 0) p = p.filter(prod => (prod.selling_price || 0) >= minP);
    if (maxP !== null) p = p.filter(prod => (prod.selling_price || 0) <= maxP);
    
    // Apply sorting
    const sort = this.sortBy();
    if (sort === 'price_low_high') {
      p.sort((a, b) => (a.selling_price || 0) - (b.selling_price || 0));
    } else if (sort === 'price_high_low') {
      p.sort((a, b) => (b.selling_price || 0) - (a.selling_price || 0));
    } else {
      p.sort((a, b) => {
        if (a.created_at && b.created_at) return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
        return (b.id > a.id ? 1 : -1);
      });
    }

    return p;
  });

  relatedProducts = computed(() => {
    const p = this.quickViewProduct();
    if (!p) return [];
    
    const catName = p.category?.name || p.category;
    const all = this.productsList().filter(prod => (prod.category?.name || prod.category) === catName && prod.id !== p.id);
    return all.slice(0, 4);
  });

  recentlyViewedProducts = computed(() => {
    const ids = this.recentlyViewedIds();
    if (ids.length === 0) return [];
    
    const all = this.productsList();
    return ids.map(id => all.find(p => p.id === id)).filter(p => !!p);
  });

  cartTotal = computed(() => {
    return this.cart().reduce((sum, item) => sum + (item.qty * (item.product.selling_price || 0)), 0);
  });

  ngOnInit() {
    this.slug.set(this.route.snapshot.paramMap.get('slug') || '');
    const branchId = this.route.snapshot.queryParamMap.get('branch_id') ||
                     this.route.snapshot.queryParamMap.get('branchId') ||
                     this.route.snapshot.paramMap.get('branchId');
    const productParams: any = { page_size: 100 };
    if (branchId) {
      productParams.branch_id = branchId;
    }
    this.productService.getProducts(productParams).subscribe({
      next: (res) => this.productsList.set(res.products || [])
    });
    this.storefrontService.loadSettings().subscribe();

    // Load Paystack Inline JS Script
    if (typeof document !== 'undefined') {
      const script = document.createElement('script');
      script.src = 'https://js.paystack.co/v1/inline.js';
      document.head.appendChild(script);
    }
    
    // Load from local storage
    try {
      const savedWishlist = localStorage.getItem(`puxbay_wishlist_${this.slug()}`);
      if (savedWishlist) {
        this.wishlist.set(new Set(JSON.parse(savedWishlist)));
      }
      const savedRecentlyViewed = localStorage.getItem(`puxbay_recently_viewed_${this.slug()}`);
      if (savedRecentlyViewed) {
        this.recentlyViewedIds.set(JSON.parse(savedRecentlyViewed));
      }
    } catch(e) {}
  }
  
  toggleWishlist(productId: string) {
    const current = new Set(this.wishlist());
    if (current.has(productId)) {
      current.delete(productId);
    } else {
      current.add(productId);
    }
    this.wishlist.set(current);
    localStorage.setItem(`puxbay_wishlist_${this.slug()}`, JSON.stringify(Array.from(current)));
  }
  
  openQuickView(product: any) {
    this.quickViewProduct.set(product);
    
    // Update recently viewed
    this.recentlyViewedIds.update(ids => {
      const newIds = [product.id, ...ids.filter(id => id !== product.id)].slice(0, 6);
      localStorage.setItem(`puxbay_recently_viewed_${this.slug()}`, JSON.stringify(newIds));
      return newIds;
    });
  }

  addToCart(product: any) {
    const existing = this.cart().find(i => i.product.id === product.id);
    if (existing) {
      this.cart.update(c => c.map(i => i.product.id === product.id ? { ...i, qty: i.qty + 1 } : i));
    } else {
      this.cart.update(c => [...c, { product, qty: 1 }]);
    }
    this.isCartOpen.set(true);
  }

  updateQty(productId: string, delta: number) {
    this.cart.update(c => {
      return c.map(i => {
        if (i.product.id === productId) {
          const newQty = i.qty + delta;
          return { ...i, qty: newQty };
        }
        return i;
      }).filter(i => i.qty > 0);
    });
  }

  startCheckout() {
    this.isCartOpen.set(false);
    this.isCheckoutOpen.set(true);
    this.checkoutStep.set(1);
    this.checkoutError.set('');
  }

  submitOrder() {
    if (this.isProcessing()) return;
    this.checkoutError.set('');

    const settings = this.storefrontService.settings();
    const guest = this.guestDetails();

    if (!guest.name || !guest.email || !guest.phone) {
      this.checkoutError.set('Please provide your name, email, and phone number.');
      return;
    }

    const payload: any = {
      total: this.cartTotal(),
      delivery_method: guest.address ? 'delivery' : 'pickup',
      delivery_address: guest.address || '',
      payment_method: this.paymentMethod(),
      items: this.cart().map(i => ({
        product_id: i.product.id,
        quantity: i.qty
      }))
    };

    if (this.paymentMethod() === 'cash') {
      this.isProcessing.set(true);
      this.http.post<any>('/api/v1/storefront/checkout/verify', payload).subscribe({
        next: (res) => {
          this.orderTrackingNumber.set(res?.order?.order_number || ('TRK-' + Math.random().toString(36).substring(2, 10).toUpperCase()));
          this.orderPlaced.set(true);
          this.cart.set([]);
          this.isProcessing.set(false);
        },
        error: (err) => {
          this.checkoutError.set(err.error?.error || 'Order placement failed. Please try again.');
          this.isProcessing.set(false);
        }
      });
      return;
    }

    // Paystack / Mobile Money Checkout with Subaccount
    const pubKey = settings?.paystack_public_key;
    if (!pubKey) {
      this.checkoutError.set('Payment gateway is not configured for this store.');
      return;
    }

    this.isProcessing.set(true);
    const amountInKobo = Math.round(this.cartTotal() * 100);

    const setupConfig: any = {
      key: pubKey,
      email: guest.email,
      amount: amountInKobo,
      currency: 'GHS',
      channels: ['mobile_money', 'card', 'bank', 'ussd', 'qr'],
      callback: (response: any) => {
        payload['payment_method'] = 'paystack';
        payload['reference'] = response.reference;
        this.http.post<any>('/api/v1/storefront/checkout/verify', payload).subscribe({
          next: (res) => {
            this.orderTrackingNumber.set(res?.order?.order_number || ('TRK-' + Math.random().toString(36).substring(2, 10).toUpperCase()));
            this.orderPlaced.set(true);
            this.cart.set([]);
            this.isProcessing.set(false);
          },
          error: (err) => {
            this.checkoutError.set(err.error?.error || 'Payment verification failed. Please contact support.');
            this.isProcessing.set(false);
          }
        });
      },
      onClose: () => {
        this.checkoutError.set('Payment was cancelled.');
        this.isProcessing.set(false);
      }
    };

    // Route directly to tenant's Paystack subaccount
    if (settings?.paystack_subaccount_code) {
      setupConfig.subaccount = settings.paystack_subaccount_code;
    }

    if ((window as any).PaystackPop) {
      const handler = (window as any).PaystackPop.setup(setupConfig);
      handler.openIframe();
    } else {
      this.checkoutError.set('Paystack script is still loading. Please try again in a few seconds.');
      this.isProcessing.set(false);
    }
  }
}
