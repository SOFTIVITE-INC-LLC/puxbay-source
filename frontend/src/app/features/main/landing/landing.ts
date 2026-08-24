import { Component, ViewEncapsulation } from '@angular/core';

import { RouterModule } from '@angular/router';
import { Hero } from './components/hero/hero';
import { FeatureOverview } from './components/feature-overview/feature-overview';
import { Testimonials } from './components/testimonials/testimonials';

@Component({
  selector: 'app-landing',
  standalone: true,
  imports: [RouterModule, Hero, FeatureOverview, Testimonials],
  templateUrl: './landing.html',
  styleUrls: ['./landing.css'],
  encapsulation: ViewEncapsulation.None,
})
export class Landing {}
