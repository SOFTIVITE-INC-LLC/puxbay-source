import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { StorefrontAuthService } from '../../../core/store/services/storefront-auth.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-store-account',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe, RouterModule],
  templateUrl: './account.component.html'
})
export class AccountComponent implements OnInit {
  authService = inject(StorefrontAuthService);
  
  activeTab = signal<'profile' | 'orders'>('profile');
  orders = signal<any[]>([]);
  isLoadingOrders = signal(false);

  profileData = {
    name: '',
    phone: '',
    address: ''
  };
  isUpdating = false;

  ngOnInit() {
    const customer = this.authService.currentUser();
    if (customer) {
      this.profileData = {
        name: customer.name || '',
        phone: customer.phone || '',
        address: customer.address || ''
      };
    }
    this.loadOrders();
  }

  loadOrders() {
    this.isLoadingOrders.set(true);
    this.authService.getOrders().subscribe({
      next: (data: any[]) => {
        this.orders.set(data);
        this.isLoadingOrders.set(false);
      },
      error: () => {
        this.isLoadingOrders.set(false);
      }
    });
  }

  updateProfile() {
    this.isUpdating = true;
    this.authService.updateMe(this.profileData).subscribe({
      next: () => {
        this.isUpdating = false;
      },
      error: () => {
        this.isUpdating = false;
      }
    });
  }

  logout() {
    this.authService.logout();
  }
}
