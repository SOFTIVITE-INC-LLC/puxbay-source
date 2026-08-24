import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { WishlistService } from '../../../core/store/services/wishlist.service';
import { CartService } from '../../../core/store/services/cart.service';
import { ToastService } from '../../../core/store/services/toast.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { Product } from '../../../core/store/models/product.model';

@Component({
  selector: 'app-wishlist',
  standalone: true,
  imports: [CommonModule, RouterModule, AppCurrencyPipe],
  templateUrl: './wishlist.component.html'
})
export class WishlistComponent implements OnInit {
  wishlistService = inject(WishlistService);
  cartService = inject(CartService);
  toastService = inject(ToastService);

  ngOnInit() {
    this.wishlistService.loadWishlistProducts();
  }

  addToCart(product: Product) {
    this.cartService.addToCart({ product_id: product.id, quantity: 1 }).subscribe({
      next: () => {
        this.toastService.show(`Added ${product.name} to cart`, 'success');
        this.wishlistService.toggleWishlist(product.id); // Remove from wishlist after adding to cart
        this.wishlistService.loadWishlistProducts(); // Reload to refresh list
      },
      error: () => this.toastService.show('Failed to add to cart', 'error')
    });
  }
}
