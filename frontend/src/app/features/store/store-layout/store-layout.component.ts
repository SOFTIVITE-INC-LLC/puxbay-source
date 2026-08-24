import { Component, inject, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { CartService } from '../../../core/store/services/cart.service';
import { WishlistService } from '../../../core/store/services/wishlist.service';
import { ToastService } from '../../../core/store/services/toast.service';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';
import { StorefrontAuthService } from '../../../core/store/services/storefront-auth.service';
import { ToastComponent } from '../../../core/store/components/toast/toast.component';
import { SearchOverlayComponent } from '../search-overlay/search-overlay.component';
import { MiniCartComponent } from '../mini-cart/mini-cart.component';
import { SocialProofComponent } from '../catalog/components/social-proof/social-proof.component';
import { WelcomeDiscountComponent } from '../catalog/components/welcome-discount/welcome-discount.component';
import { CurrencyService } from '../../../core/store/services/currency.service';
import { ThemeService } from '../../../core/services/theme.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-store-layout',
  standalone: true,
  imports: [CommonModule, RouterModule, ToastComponent, FormsModule, SearchOverlayComponent, MiniCartComponent, SocialProofComponent, WelcomeDiscountComponent, AppCurrencyPipe],
  templateUrl: './store-layout.component.html'
})
export class StoreLayoutComponent implements OnInit {
  cartService = inject(CartService);
  toastService = inject(ToastService);
  settingsService = inject(StorefrontSettingsService);
  wishlistService = inject(WishlistService);
  authService = inject(StorefrontAuthService);
  currencyService = inject(CurrencyService);
  themeService = inject(ThemeService);
  http = inject(HttpClient);

  mobileMenuOpen = false;
  searchOpen = false;
  miniCartOpen = false;
  currentYear = new Date().getFullYear();

  newsletterEmail = '';
  isSubscribing = false;

  ngOnInit() {
    this.cartService.loadCart();
    this.settingsService.loadSettings().subscribe();
    this.currencyService.init();
  }

  toggleMobileMenu() {
    this.mobileMenuOpen = !this.mobileMenuOpen;
  }

  subscribeNewsletter() {
    if (!this.newsletterEmail || this.isSubscribing) return;
    
    this.isSubscribing = true;
    this.http.post('/api/v1/storefront/newsletter/subscribe', { email: this.newsletterEmail }).subscribe({
      next: () => {
        this.toastService.show('Successfully subscribed to newsletter!', 'success');
        this.newsletterEmail = '';
        this.isSubscribing = false;
      },
      error: () => {
        this.toastService.show('Failed to subscribe. Please try again.', 'error');
        this.isSubscribing = false;
      }
    });
  }
}
