import { Routes } from '@angular/router';

export const storeRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./store-layout/store-layout.component').then(m => m.StoreLayoutComponent),
    children: [
      {
        path: '',
        loadComponent: () => import('./catalog/catalog.component').then(m => m.CatalogComponent)
      },
      {
        path: 'product/:id',
        loadComponent: () => import('./product-detail/product-detail.component').then(m => m.ProductDetailComponent)
      },
      {
        path: 'cart',
        loadComponent: () => import('./cart/cart.component').then(m => m.CartComponent)
      },
      {
        path: 'checkout',
        loadComponent: () => import('./checkout/checkout.component').then(m => m.CheckoutComponent)
      },
      {
        path: 'wishlist',
        loadComponent: () => import('./wishlist/wishlist.component').then(m => m.WishlistComponent)
      },
      {
        path: 'track-order',
        loadComponent: () => import('./track-order/track-order.component').then(m => m.TrackOrderComponent)
      },
      {
        path: 'order-confirmation',
        loadComponent: () => import('./order-confirmation/order-confirmation.component').then(m => m.OrderConfirmationComponent)
      },
      {
        path: 'login',
        loadComponent: () => import('./auth/login/login.component').then(m => m.LoginComponent)
      },
      {
        path: 'register',
        loadComponent: () => import('./auth/register/register.component').then(m => m.RegisterComponent)
      },
      {
        path: 'account',
        loadComponent: () => import('./account/account.component').then(m => m.AccountComponent)
      }
    ]
  }
];
