import { Component, inject, OnInit, OnDestroy, signal, computed, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { RouterModule, Router, ActivatedRoute } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ProductService } from '../../../core/store/services/product.service';
import { CartService } from '../../../core/store/services/cart.service';
import { ToastService } from '../../../core/store/services/toast.service';
import { WishlistService } from '../../../core/store/services/wishlist.service';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { Product, Category } from '../../../core/store/models/product.model';
import { RecentlyViewedService } from '../../../core/store/services/recently-viewed.service';

import { EmptyStateComponent } from '../../../shared/components/empty-state/empty-state.component';

@Component({
  selector: 'app-catalog',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule, AppCurrencyPipe, EmptyStateComponent],
  templateUrl: './catalog.component.html'
})
export class CatalogComponent implements OnInit, OnDestroy {
  productService = inject(ProductService);
  cartService = inject(CartService);
  toastService = inject(ToastService);
  wishlistService = inject(WishlistService);
  settingsService = inject(StorefrontSettingsService);
  router = inject(Router);
  route = inject(ActivatedRoute);
  recentlyViewedService = inject(RecentlyViewedService);
  destroyRef = inject(DestroyRef);

  branchId = signal<string | null>(null);
  products = signal<Product[]>([]);
  categories = signal<Category[]>([]);
  searchQuery = signal('');
  selectedCategoryId = signal<string | null>(null);
  minPrice = signal<number | null>(null);
  maxPrice = signal<number | null>(null);
  minRating = signal<number | null>(null);
  inStockOnly = signal(false);
  sortBy = signal<string>('latest');
  isLoading = signal(true);
  viewMode = signal<'grid' | 'list'>('grid');
  isFiltersOpen = signal<boolean>(typeof window !== 'undefined' ? window.innerWidth > 768 : false);

  // Selected Category computed
  selectedCategory = computed(() => {
    const id = this.selectedCategoryId();
    if (!id) return null;
    return this.categories().find(c => c.id === id) || null;
  });

  // Active filter count
  activeFiltersCount = computed(() => {
    let count = 0;
    if (this.selectedCategoryId()) count++;
    if (this.minPrice() !== null || this.maxPrice() !== null) count++;
    if (this.inStockOnly()) count++;
    if (this.minRating() !== null) count++;
    if (this.searchQuery().trim()) count++;
    return count;
  });

  // Pagination
  currentPage = signal(1);
  totalPages = signal(1);
  totalProducts = signal(0);
  pageSize = 12;

  // Comparison State
  compareProducts = signal<Product[]>([]);
  isCompareModalOpen = signal(false);

  // Adding state
  addingProductId = signal<string | null>(null);

  // Quick View
  quickViewProduct = signal<Product | null>(null);

  // Recently Viewed & Related Products
  recentlyViewedProducts = computed(() => this.recentlyViewedService.recentlyViewedProducts());

  relatedProducts = computed(() => {
    const p = this.quickViewProduct();
    if (!p) return [];

    // Find other products with the same category
    const all = this.products(); // We only have the current page's products, but that's fine for now
    return all.filter(prod => prod.id !== p.id && prod.category?.id === p.category?.id).slice(0, 4);
  });

  // Flash Sale
  saleTimeLeft = signal({ hours: 0, minutes: 0, seconds: 0 });
  private timerInterval: any;

  // Back in stock notification
  notifyEmail = signal('');

  ngOnInit() {
    // Check both route params (:branchId) and query params (?branch_id=...)
    const paramBranch = this.route.snapshot.paramMap.get('branchId');
    const queryBranch = this.route.snapshot.queryParamMap.get('branch_id') || this.route.snapshot.queryParamMap.get('branchId');
    if (paramBranch) {
      this.branchId.set(paramBranch);
      if (typeof window !== 'undefined') localStorage.setItem('store_branch_id', paramBranch);
    } else if (queryBranch) {
      this.branchId.set(queryBranch);
      if (typeof window !== 'undefined') localStorage.setItem('store_branch_id', queryBranch);
    }

    this.route.paramMap.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(params => {
      const b = params.get('branchId');
      if (b && b !== this.branchId()) {
        this.branchId.set(b);
        if (typeof window !== 'undefined') localStorage.setItem('store_branch_id', b);
        this.loadCategories();
        this.loadProducts();
      }
    });

    this.route.queryParamMap.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(qp => {
      const b = qp.get('branch_id') || qp.get('branchId');
      if (b && b !== this.branchId()) {
        this.branchId.set(b);
        if (typeof window !== 'undefined') localStorage.setItem('store_branch_id', b);
        this.loadCategories();
        this.loadProducts();
      }
    });

    this.loadCategories();
    this.loadProducts();
    this.recentlyViewedService.loadRecentlyViewed();
    this.startCountdown();
  }

  ngOnDestroy() {
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
    }
  }

  startCountdown() {
    const updateTime = () => {
      let targetTime: Date;
      const settings = this.settingsService.settings();
      if (settings?.flash_sale_end_time) {
        targetTime = new Date(settings.flash_sale_end_time);
      } else {
        // Fallback to midnight if not configured
        const now = new Date();
        targetTime = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1);
      }

      const diff = targetTime.getTime() - new Date().getTime();
      if (diff <= 0) {
        this.saleTimeLeft.set({ hours: 0, minutes: 0, seconds: 0 });
        return;
      }

      this.saleTimeLeft.set({
        hours: Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)),
        minutes: Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60)),
        seconds: Math.floor((diff % (1000 * 60)) / 1000)
      });
    };

    updateTime();
    this.timerInterval = setInterval(updateTime, 1000);
  }

  submitNotify() {
    if (!this.notifyEmail() || !this.notifyEmail().includes('@')) {
      this.toastService.show('Please enter a valid email.', 'error');
      return;
    }

    const product = this.quickViewProduct();
    if (!product) return;

    this.productService.notifyRestock(product.id, this.notifyEmail()).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: () => {
        this.toastService.show('You will be notified when this is back in stock!', 'success');
        this.notifyEmail.set('');
      },
      error: () => {
        this.toastService.show('Failed to subscribe for notifications.', 'error');
      }
    });
  }

  loadCategories() {
    this.productService.getCategories(this.branchId() || undefined).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (cats) => this.categories.set(cats || []),
      error: (err) => console.error('Failed to load categories', err)
    });
  }

  loadProducts(append = false) {
    if (!append) {
      this.isLoading.set(true);
      this.currentPage.set(1);
    }

    let params: any = {
      page: this.currentPage(),
      page_size: 12
    };

    if (this.branchId()) {
      params.branch_id = this.branchId();
    }
    if (this.selectedCategoryId()) {
      params.category_id = this.selectedCategoryId();
    }
    if (this.searchQuery()) {
      params.search = this.searchQuery();
    }
    if (this.minPrice() !== null) {
      params.min_price = this.minPrice();
    }
    if (this.maxPrice() !== null) {
      params.max_price = this.maxPrice();
    }
    if (this.inStockOnly()) {
      params.in_stock = 'true';
    }
    if (this.sortBy()) {
      params.sort = this.sortBy();
    }

    this.productService.getProducts(params).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (res) => {
        if (append) {
          this.products.update(current => [...current, ...(res.products || [])]);
        } else {
          this.products.set(res.products || []);
        }
        this.totalPages.set(res.total_pages || 1);
        this.totalProducts.set(res.total || 0);
        this.isLoading.set(false);
      },
      error: () => this.isLoading.set(false)
    });
  }

  loadMore() {
    if (this.currentPage() < this.totalPages()) {
      this.currentPage.update(p => p + 1);
      this.loadProducts(true);
    }
  }

  calculatePoints(price: number): number {
    return Math.floor(price * 10);
  }

  onSearch(query: string) {
    this.searchQuery.set(query);
    this.currentPage.set(1);
    this.loadProducts();
  }

  applyFilters() {
    this.currentPage.set(1);
    this.loadProducts();
  }

  setPricePreset(min: number | null, max: number | null) {
    this.minPrice.set(min);
    this.maxPrice.set(max);
    this.applyFilters();
  }

  setMinRating(rating: number | null) {
    this.minRating.set(this.minRating() === rating ? null : rating);
    this.applyFilters();
  }

  clearFilter(filter: 'category' | 'price' | 'stock' | 'rating' | 'search') {
    if (filter === 'category') this.selectedCategoryId.set(null);
    if (filter === 'price') { this.minPrice.set(null); this.maxPrice.set(null); }
    if (filter === 'stock') this.inStockOnly.set(false);
    if (filter === 'rating') this.minRating.set(null);
    if (filter === 'search') this.searchQuery.set('');
    this.applyFilters();
  }

  resetFilters() {
    this.selectedCategoryId.set(null);
    this.minPrice.set(null);
    this.maxPrice.set(null);
    this.minRating.set(null);
    this.inStockOnly.set(false);
    this.searchQuery.set('');
    this.sortBy.set('latest');
    this.applyFilters();
  }

  openQuickView(product: Product, event: Event) {
    event.preventDefault();
    event.stopPropagation();
    this.quickViewProduct.set(product);

    // Update recently viewed
    this.recentlyViewedService.addProduct(product.id);
    this.recentlyViewedService.loadRecentlyViewed();
  }

  navigateToProduct(id: string) {
    // using inject(Router)
    this.router.navigate(['/store/product', id]);
  }

  closeQuickView() {
    this.quickViewProduct.set(null);
  }

  selectCategory(categoryId: string | null) {
    if (this.selectedCategoryId() === categoryId) return;
    this.selectedCategoryId.set(categoryId);
    this.currentPage.set(1);
    this.loadProducts();
  }

  onSortChange(sort: string) {
    this.sortBy.set(sort);
    this.loadProducts();
  }

  goToPage(page: number) {
    if (page < 1 || page > this.totalPages()) return;
    this.currentPage.set(page);
    this.loadProducts();
  }

  quickAddToCart(product: Product, event: Event) {
    event.stopPropagation();
    event.preventDefault();
    this.addingProductId.set(product.id);
    this.cartService.addToCart({ product_id: product.id, quantity: 1 }).subscribe({
      next: () => {
        this.toastService.show(`Added ${product.name} to cart`, 'success');
        setTimeout(() => this.addingProductId.set(null), 800);
      },
      error: () => {
        this.toastService.show('Failed to add to cart', 'error');
        this.addingProductId.set(null);
      }
    });
  }

  get pages(): number[] {
    const total = this.totalPages();
    const current = this.currentPage();
    const pages: number[] = [];
    const start = Math.max(1, current - 2);
    const end = Math.min(total, current + 2);
    for (let i = start; i <= end; i++) pages.push(i);
    return pages;
  }

  // Comparison Actions
  toggleCompare(product: Product, event: Event) {
    event.stopPropagation();
    event.preventDefault();
    const current = this.compareProducts();
    const exists = current.find(p => p.id === product.id);

    if (exists) {
      this.compareProducts.set(current.filter(p => p.id !== product.id));
    } else {
      if (current.length >= 3) {
        this.toastService.show('You can only compare up to 3 products at a time', 'info');
        return;
      }
      this.compareProducts.set([...current, product]);
    }
  }

  openCompareModal() {
    if (this.compareProducts().length < 2) {
      this.toastService.show('Select at least 2 products to compare', 'info');
      return;
    }
    this.isCompareModalOpen.set(true);
  }

  closeCompareModal() {
    this.isCompareModalOpen.set(false);
  }

  removeCompareProduct(id: string) {
    this.compareProducts.set(this.compareProducts().filter(p => p.id !== id));
    if (this.compareProducts().length < 2) {
      this.closeCompareModal();
    }
  }
}
