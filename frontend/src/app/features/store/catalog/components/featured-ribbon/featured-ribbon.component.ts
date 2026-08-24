import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { ProductService } from '../../../../../core/store/services/product.service';
import { CartService } from '../../../../../core/store/services/cart.service';
import { WishlistService } from '../../../../../core/store/services/wishlist.service';
import { AppCurrencyPipe } from '../../../../../core/pipes/app-currency.pipe';
import { Product } from '../../../../../core/store/models/product.model';

@Component({
  selector: 'app-featured-ribbon',
  standalone: true,
  imports: [CommonModule, RouterModule, AppCurrencyPipe],
  templateUrl: './featured-ribbon.component.html'
})
export class FeaturedRibbonComponent implements OnInit {
  productService = inject(ProductService);
  cartService = inject(CartService);
  wishlistService = inject(WishlistService);

  products = signal<Product[]>([]);
  isLoading = signal(true);
  addingProductId = signal<string | null>(null);

  ngOnInit() {
    this.loadFeatured();
  }

  loadFeatured() {
    // We'll just fetch latest products and use them as featured
    this.productService.searchProducts('page=1&page_size=8&sort=latest').subscribe({
      next: (res) => {
        this.products.set(res.products || []);
        this.isLoading.set(false);
      },
      error: () => this.isLoading.set(false)
    });
  }

  quickAddToCart(product: Product, event: Event) {
    event.preventDefault();
    event.stopPropagation();
    if (product.stock_quantity <= 0 || this.addingProductId()) return;

    this.addingProductId.set(product.id);
    this.cartService.addToCart({ product_id: product.id, quantity: 1 }).subscribe({
      next: () => {
        setTimeout(() => this.addingProductId.set(null), 1000);
      },
      error: () => this.addingProductId.set(null)
    });
  }
}
