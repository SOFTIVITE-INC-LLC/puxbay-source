import { Component, OnDestroy, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

interface Slide {
  title: string;
  subtitle: string;
  ctaText: string;
  ctaLink: string;
  bgImage: string;
  theme: 'indigo' | 'rose' | 'emerald';
}

@Component({
  selector: 'app-hero-carousel',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './hero-carousel.component.html'
})
export class HeroCarouselComponent implements OnInit, OnDestroy {
  slides: Slide[] = [
    {
      title: 'Summer Tech Sale',
      subtitle: 'Up to 50% off on premium audio and accessories.',
      ctaText: 'Shop the Sale',
      ctaLink: '/store',
      bgImage: 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?q=80&w=1000&auto=format&fit=crop',
      theme: 'indigo'
    },
    {
      title: 'New Arrivals',
      subtitle: 'Discover the latest smart home innovations.',
      ctaText: 'Explore Now',
      ctaLink: '/store',
      bgImage: 'https://images.unsplash.com/photo-1558089687-f282ffcbc126?q=80&w=1000&auto=format&fit=crop',
      theme: 'rose'
    },
    {
      title: 'Workspace Essentials',
      subtitle: 'Elevate your productivity with our premium gear.',
      ctaText: 'View Collection',
      ctaLink: '/store',
      bgImage: 'https://images.unsplash.com/photo-1527443154391-507e9dc6c5cc?q=80&w=1000&auto=format&fit=crop',
      theme: 'emerald'
    }
  ];

  currentSlide = signal(0);
  private timer: any;

  ngOnInit() {
    this.startCarousel();
  }

  ngOnDestroy() {
    this.stopCarousel();
  }

  startCarousel() {
    this.timer = setInterval(() => {
      this.nextSlide();
    }, 5000);
  }

  stopCarousel() {
    if (this.timer) clearInterval(this.timer);
  }

  nextSlide() {
    this.currentSlide.set((this.currentSlide() + 1) % this.slides.length);
  }

  prevSlide() {
    this.currentSlide.set((this.currentSlide() - 1 + this.slides.length) % this.slides.length);
  }

  goToSlide(index: number) {
    this.currentSlide.set(index);
    this.stopCarousel();
    this.startCarousel();
  }
}
