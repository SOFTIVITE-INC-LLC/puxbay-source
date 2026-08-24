import { Component, ViewEncapsulation, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { Testimonials } from '../landing/components/testimonials/testimonials';

@Component({
  selector: 'app-storefront-product',
  standalone: true,
  imports: [RouterModule, CommonModule, Testimonials],
  templateUrl: './storefront-product.html',
  encapsulation: ViewEncapsulation.None,
})
export class StorefrontProduct {
  isScanning = signal(false);

  toggleScan() {
    this.isScanning.set(!this.isScanning());
  }
}
