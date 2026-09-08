import { Component, OnDestroy, signal, afterNextRender } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';

export interface HeroSlide {
  id: number;
  image: string;
  badge: string;
  headline: string;
  headlineAccent: string;
  subtext: string;
  ctaLabel: string;
  ctaLink: string;
  secondaryCtaLabel?: string;
  accentColor: string;
  overlayGradient: string;
}

@Component({
  selector: 'app-hero',
  standalone: true,
  imports: [RouterModule, CommonModule],
  templateUrl: './hero.html',
  styleUrl: './hero.css',
})
export class Hero implements OnDestroy {
  readonly slides: HeroSlide[] = [
    {
      id: 0,
      image: '/hero-slide-1.jpg',
      badge: 'Your Retail Operating System',
      headline: 'Run Your Store.',
      headlineAccent: 'Grow Your Empire.',
      subtext: 'The all-in-one POS, inventory, and storefront platform trusted by 2,000+ businesses worldwide.',
      ctaLabel: 'Start Free Trial',
      ctaLink: '/login',
      secondaryCtaLabel: 'Watch Demo',
      accentColor: '#005b96',
      overlayGradient: 'linear-gradient(135deg, rgba(1,31,75,0.82) 0%, rgba(3,57,108,0.65) 50%, rgba(0,0,0,0.25) 100%)',
    },
    {
      id: 1,
      image: '/hero-slide-2.jpg',
      badge: 'Smart Inventory',
      headline: 'Always in Stock.',
      headlineAccent: 'Never Out of Control.',
      subtext: 'AI-powered forecasting and multi-location inventory management — so you are always stocked before customers ask.',
      ctaLabel: 'Explore Inventory',
      ctaLink: '/product/inventory',
      secondaryCtaLabel: 'See Features',
      accentColor: '#03396c',
      overlayGradient: 'linear-gradient(135deg, rgba(3,57,108,0.85) 0%, rgba(1,31,75,0.65) 50%, rgba(0,0,0,0.3) 100%)',
    },
    {
      id: 2,
      image: '/hero-slide-3.jpg',
      badge: 'Zero-Setup Ecommerce',
      headline: 'Launch Your',
      headlineAccent: 'Online Store Today.',
      subtext: 'Beautiful storefronts, seamless checkout, and integrated inventory — live in minutes with no technical expertise required.',
      ctaLabel: 'Create Your Store',
      ctaLink: '/product/storefront',
      secondaryCtaLabel: 'See Examples',
      accentColor: '#005b96',
      overlayGradient: 'linear-gradient(135deg, rgba(1,31,75,0.75) 0%, rgba(100,151,177,0.55) 60%, rgba(0,0,0,0.2) 100%)',
    },
    {
      id: 3,
      image: '/hero-slide-4.jpg',
      badge: 'AI-Powered Analytics',
      headline: 'Decisions Backed',
      headlineAccent: 'by Real Data.',
      subtext: 'Real-time revenue dashboards, customer insights, and AI forecasting — every decision backed by live data.',
      ctaLabel: 'Explore Analytics',
      ctaLink: '/product/analytics',
      secondaryCtaLabel: 'See Reports',
      accentColor: '#011f4b',
      overlayGradient: 'linear-gradient(135deg, rgba(1,31,75,0.90) 0%, rgba(3,57,108,0.60) 50%, rgba(0,0,0,0.15) 100%)',
    },
  ];

  activeSlide = signal(0);
  isTransitioning = signal(false);
  private intervalId: ReturnType<typeof setInterval> | null = null;

  constructor() {
    afterNextRender(() => {
      this.startAutoplay();
    });
  }

  startAutoplay(): void {
    this.intervalId = setInterval(() => {
      this.goToNext();
    }, 5500);
  }

  goTo(index: number): void {
    if (index === this.activeSlide() || this.isTransitioning()) return;
    this.isTransitioning.set(true);
    this.activeSlide.set(index);
    // Reset timer
    if (this.intervalId) {
      clearInterval(this.intervalId);
    }
    setTimeout(() => {
      this.isTransitioning.set(false);
      this.startAutoplay();
    }, 800);
  }

  goToNext(): void {
    const next = (this.activeSlide() + 1) % this.slides.length;
    this.activeSlide.set(next);
  }

  goToPrev(): void {
    const prev = (this.activeSlide() - 1 + this.slides.length) % this.slides.length;
    this.goTo(prev);
  }

  ngOnDestroy(): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
    }
  }
}
