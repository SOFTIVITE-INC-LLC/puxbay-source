import { Component, ViewEncapsulation } from '@angular/core';

import { RouterModule } from '@angular/router';
import { Testimonials } from '../landing/components/testimonials/testimonials';

@Component({
  selector: 'app-inventory-product',
  standalone: true,
  imports: [RouterModule, Testimonials],
  templateUrl: './inventory-product.html',
  encapsulation: ViewEncapsulation.None,
})
export class InventoryProduct {}
