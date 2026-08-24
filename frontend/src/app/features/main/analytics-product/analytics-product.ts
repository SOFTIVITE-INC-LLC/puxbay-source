import { Component, ViewEncapsulation } from '@angular/core';

import { RouterModule } from '@angular/router';
import { Testimonials } from '../landing/components/testimonials/testimonials';

@Component({
  selector: 'app-analytics-product',
  standalone: true,
  imports: [RouterModule, Testimonials],
  templateUrl: './analytics-product.html',
  encapsulation: ViewEncapsulation.None,
})
export class AnalyticsProduct {}
