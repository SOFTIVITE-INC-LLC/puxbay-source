import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { CatalogService } from '../../../core/services/catalog.service';
import { StorefrontService } from '../../../core/services/storefront.service';

@Component({
  selector: 'app-public-storefront',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './public-storefront.html',
})
export class PublicStorefront implements OnInit {
  private route = inject(ActivatedRoute);
  catalogService = inject(CatalogService);
  storefrontService = inject(StorefrontService);

  slug = signal('');
  cart = signal<{product: any, qty: number}[]>([]);
  isCartOpen = signal(false);
  isCheckoutOpen = signal(false);
  checkoutStep = signal(1);

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
    const cats = new Set(this.catalogService.products().map(p => p.category));
    return ['All', ...Array.from(cats).filter(c => !!c)];
  });

  filteredProducts = computed(() => {
    let p = this.catalogService.products().filter(prod => prod.is_active);
    
    // Apply text search
    const q = this.searchQuery().toLowerCase();
    if (q) p = p.filter(prod => prod.name.toLowerCase().includes(q) || prod.sku?.toLowerCase().includes(q));
    
    // Apply category
    const c = this.selectedCategory();
    if (c !== 'All') p = p.filter(prod => prod.category === c);
    
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
      // newest (default fallback assuming higher id or created_at logic, here just reverse order as a mock for 'newest' if no dates exist)
      // or we can sort by id descending
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
    
    const all = this.catalogService.products().filter(prod => prod.is_active && prod.id !== p.id && prod.category === p.category);
    // Return max 4 related products
    return all.slice(0, 4);
  });

  recentlyViewedProducts = computed(() => {
    const ids = this.recentlyViewedIds();
    if (ids.length === 0) return [];
    
    const all = this.catalogService.products();
    // Maintain order of recentlyViewedIds
    return ids.map(id => all.find(p => p.id === id)).filter(p => !!p);
  });

  cartTotal = computed(() => {
    return this.cart().reduce((sum, item) => sum + (item.qty * (item.product.selling_price || 0)), 0);
  });

  ngOnInit() {
    this.slug.set(this.route.snapshot.paramMap.get('slug') || '');
    this.catalogService.getProducts().subscribe();
    this.storefrontService.getSettings().subscribe();
    
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
  }

  submitOrder() {
    this.orderTrackingNumber.set('TRK-' + Math.random().toString(36).substring(2, 10).toUpperCase());
    this.orderPlaced.set(true);
    this.cart.set([]);
  }
}
