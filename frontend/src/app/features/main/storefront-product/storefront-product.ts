import { Component, ViewEncapsulation, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { Testimonials } from '../landing/components/testimonials/testimonials';
import { SettingsService } from '../../../core/services/settings.service';

@Component({
  selector: 'app-storefront-product',
  standalone: true,
  imports: [RouterModule, CommonModule, Testimonials],
  templateUrl: './storefront-product.html',
  encapsulation: ViewEncapsulation.None,
})
export class StorefrontProduct {
  settingsService = inject(SettingsService);
  isScanning = signal(false);

  toggleScan() {
    this.isScanning.set(!this.isScanning());
  }
}
