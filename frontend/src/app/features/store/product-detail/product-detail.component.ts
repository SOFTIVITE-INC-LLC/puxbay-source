import { Component, computed, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule, Router } from '@angular/router';
import { ProductService } from '../../../core/store/services/product.service';
import { CartService } from '../../../core/store/services/cart.service';
import { ToastService } from '../../../core/store/services/toast.service';
import { RecentlyViewedService } from '../../../core/store/services/recently-viewed.service';
import { WishlistService } from '../../../core/store/services/wishlist.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { Product, ProductReview } from '../../../core/store/models/product.model';
import { Title, Meta } from '@angular/platform-browser';

import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-product-detail',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule, AppCurrencyPipe],
  templateUrl: './product-detail.component.html',
  styleUrls: ['./product-detail.component.css']
})
export class ProductDetailComponent implements OnInit {
  navigateToProduct(product: any) { }
  route = inject(ActivatedRoute);
  router = inject(Router);
  productService = inject(ProductService);
  cartService = inject(CartService);
  toastService = inject(ToastService);
  recentlyViewedService = inject(RecentlyViewedService);
  wishlistService = inject(WishlistService);
  titleService = inject(Title);
  metaService = inject(Meta);

  product = signal<Product | null>(null);
  images = signal<{ image_url: string }[]>([]);
  activeImage = signal<string | null>(null);
  reviews = signal<ProductReview[]>([]);
  avgRating = signal(0);
  relatedProducts = signal<Product[]>([]);
  isLoading = signal(true);
  quantity = signal(1);
  isAdding = signal(false);
  addedSuccess = signal(false);

  // Reviews state
  isReviewModalOpen = signal(false);
  newReviewRating = signal(0);
  hoverRating = signal(0);
  newReviewComment = signal('');
  newReviewAuthor = signal('');
  isSubmittingReview = signal(false);
  selectedReviewFilter = signal<number | null>(null);

  // Sticky Mobile Bar State
  showStickyBar = signal(false);

  // Bundle state
  isAddingBundle = signal(false);

  // Computed Rating Breakdown (5 to 1 stars)
  ratingBreakdown = computed(() => {
    const list = this.reviews();
    const total = list.length;
    const counts: { [star: number]: { count: number, percent: number } } = {
      5: { count: 0, percent: 0 },
      4: { count: 0, percent: 0 },
      3: { count: 0, percent: 0 },
      2: { count: 0, percent: 0 },
      1: { count: 0, percent: 0 },
    };

    if (total === 0) return counts;

    list.forEach(r => {
      const star = Math.min(5, Math.max(1, Math.round(r.rating)));
      if (counts[star]) counts[star].count++;
    });

    for (let i = 1; i <= 5; i++) {
      counts[i].percent = Math.round((counts[i].count / total) * 100);
    }
    return counts;
  });

  // Filtered reviews
  filteredReviews = computed(() => {
    const filter = this.selectedReviewFilter();
    const list = this.reviews();
    if (!filter) return list;
    return list.filter(r => Math.round(r.rating) === filter);
  });

  ngOnInit() {
    this.route.paramMap.subscribe(params => {
      const id = params.get('id');
      if (id) {
        this.loadProduct(id);
      }
    });
    this.recentlyViewedService.loadRecentlyViewed();

    if (typeof window !== 'undefined') {
      window.addEventListener('scroll', this.onWindowScroll);
    }
  }

  ngOnDestroy() {
    if (typeof window !== 'undefined') {
      window.removeEventListener('scroll', this.onWindowScroll);
    }
  }

  onWindowScroll = () => {
    if (typeof window !== 'undefined') {
      this.showStickyBar.set(window.scrollY > 350);
    }
  };

  loadProduct(id: string) {
    this.isLoading.set(true);
    this.productService.getProduct(id).subscribe({
      next: (res) => {
        this.product.set(res.product);
        this.images.set(res.images || []);

        // Set active image (main product image or first gallery image)
        if (res.product.image_url) {
          this.activeImage.set(res.product.image_url);
        } else if (res.images && res.images.length > 0) {
          this.activeImage.set(res.images[0].image_url);
        }

        this.reviews.set(res.reviews || []);
        this.avgRating.set(res.avg_rating || 0);
        this.relatedProducts.set(res.related_products || []);
        this.isLoading.set(false);
        this.recentlyViewedService.addProduct(id);
        this.updateSEO(res.product);
      },
      error: () => this.isLoading.set(false)
    });
  }

  updateSEO(product: Product) {
    this.titleService.setTitle(`${product.name} | Puxbay`);
    this.metaService.updateTag({ name: 'description', content: product.description || '' });

    this.metaService.updateTag({ property: 'og:title', content: product.name });
    this.metaService.updateTag({ property: 'og:description', content: product.description || '' });
    if (product.image_url) {
      this.metaService.updateTag({ property: 'og:image', content: product.image_url });
    }
  }

  increment() { this.quantity.update(q => q + 1); }
  decrement() { this.quantity.update(q => q > 1 ? q - 1 : 1); }

  addToCart() {
    const p = this.product();
    if (!p || this.isAdding()) return;

    this.isAdding.set(true);
    this.cartService.addToCart({
      product_id: p.id,
      quantity: this.quantity()
    }).subscribe({
      next: () => {
        this.isAdding.set(false);
        this.addedSuccess.set(true);
        setTimeout(() => this.addedSuccess.set(false), 4000);
      },
      error: () => this.isAdding.set(false)
    });
  }

  goToCart() {
    this.router.navigate(['/store/cart']);
  }

  readonly starArray = [1, 2, 3, 4, 5];

  getStarIcon(star: number, rating: number): string {
    if (rating >= star) return 'star';
    if (rating >= star - 0.5) return 'star_half';
    return 'star_border';
  }

  shareProduct() {
    const url = window.location.href;
    if (navigator.share) {
      navigator.share({
        title: this.product()?.name || 'Check this out',
        text: this.product()?.description || 'I found this product on the store',
        url: url
      }).catch(err => {
        // user cancelled or error
      });
    } else {
      navigator.clipboard.writeText(url).then(() => {
        this.toastService.show('Link copied to clipboard', 'info');
      });
    }
  }

  // Reviews
  openReviewModal() {
    this.isReviewModalOpen.set(true);
    this.newReviewRating.set(0);
    this.newReviewComment.set('');
  }

  closeReviewModal() {
    this.isReviewModalOpen.set(false);
  }

  submitReview() {
    const p = this.product();
    if (!p || this.newReviewRating() === 0) return;

    this.isSubmittingReview.set(true);

    // Using mock customer ID since we don't have auth state yet
    const reviewData = {
      customer_id: '123e4567-e89b-12d3-a456-426614174000',
      rating: this.newReviewRating(),
      comment: this.newReviewComment()
    };

    this.productService.submitReview(p.id, reviewData).subscribe({
      next: (review) => {
        // Optimistically update the UI
        this.reviews.update(current => [review, ...current]);

        // Recalculate average rating
        const currentReviews = this.reviews();
        const sum = currentReviews.reduce((acc, r) => acc + r.rating, 0);
        this.avgRating.set(sum / currentReviews.length);

        this.toastService.show('Review submitted successfully!', 'success');
        this.isSubmittingReview.set(false);
        this.closeReviewModal();
      },
      error: () => {
        this.toastService.show('Failed to submit review.', 'error');
        this.isSubmittingReview.set(false);
      }
    });
  }

  // Bundles
  addAllToCart() {
    const p = this.product();
    const related = this.relatedProducts();

    if (!p || related.length === 0 || this.isAddingBundle()) return;

    this.isAddingBundle.set(true);

    // Add current product
    this.cartService.addToCart({
      product_id: p.id,
      quantity: 1
    }).subscribe({
      next: () => {
        // Then add the related product
        this.cartService.addToCart({
          product_id: related[0].id,
          quantity: 1
        }).subscribe({
          next: () => {
            this.isAddingBundle.set(false);
            this.addedSuccess.set(true);
            this.toastService.show('Bundle added to cart!', 'success');
            setTimeout(() => this.addedSuccess.set(false), 4000);
          },
          error: () => {
            this.isAddingBundle.set(false);
            this.toastService.show('Failed to add the related item.', 'error');
          }
        });
      },
      error: () => {
        this.isAddingBundle.set(false);
        this.toastService.show('Failed to add the primary item.', 'error');
      }
    });
  }
}
