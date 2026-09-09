import { Component, OnInit, OnDestroy, NgZone, Inject, PLATFORM_ID, afterNextRender, signal, HostListener } from '@angular/core';
import { isPlatformBrowser, DOCUMENT, NgClass } from '@angular/common';
import { RouterModule, Router, NavigationStart, NavigationEnd } from '@angular/router';
import { Subscription } from 'rxjs';
import { filter } from 'rxjs/operators';
import { AuthService } from '../../services/auth.service';
import { CookieConsentService } from '../../services/cookie-consent.service';

@Component({
  selector: 'app-public-layout',
  standalone: true,
  imports: [RouterModule, NgClass],
  templateUrl: './public-layout.html',
})
export class PublicLayout implements OnInit, OnDestroy {
  private observer: IntersectionObserver | null = null;
  private intervalId: any;
  private routerSub: Subscription | null = null;

  mobileMenuOpen = signal(false);
  scrolled = signal(false);

  constructor(
    private ngZone: NgZone,
    private router: Router,
    @Inject(DOCUMENT) private document: Document,
    @Inject(PLATFORM_ID) private platformId: Object,
    public authService: AuthService,
    public cookieConsentService: CookieConsentService
  ) {
    afterNextRender(() => {
      this.initObserver();
    });
  }

  ngOnInit() {
    if (isPlatformBrowser(this.platformId)) {
      this.document.documentElement.classList.remove('dark');
    }

    // Close mobile menu on route change
    this.routerSub = this.router.events
      .pipe(filter(e => e instanceof NavigationStart))
      .subscribe(() => this.mobileMenuOpen.set(false));
  }

  @HostListener('window:scroll', [])
  onScroll() {
    if (isPlatformBrowser(this.platformId)) {
      this.scrolled.set(window.scrollY > 60);
    }
  }

  toggleMobileMenu() {
    this.mobileMenuOpen.update(v => !v);
  }

  closeMobileMenu() {
    this.mobileMenuOpen.set(false);
  }

  initObserver() {
    this.observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          entry.target.classList.add('active');
          this.observer?.unobserve(entry.target);
        }
      });
    }, { 
      threshold: 0.08,
      rootMargin: '0px 0px -40px 0px'
    });

    this.observeElements();

    // Re-observe on route transitions
    this.router.events
      .pipe(filter(e => e instanceof NavigationEnd))
      .subscribe(() => {
        setTimeout(() => this.observeElements(), 50);
      });

    this.ngZone.runOutsideAngular(() => {
      this.intervalId = setInterval(() => this.observeElements(), 800);
    });
  }

  observeElements() {
    if (!this.observer) return;
    const selector = '.reveal:not(.active), .reveal-up:not(.active), .reveal-down:not(.active), .reveal-left:not(.active), .reveal-right:not(.active), .reveal-scale:not(.active), .reveal-rotate-left:not(.active), .reveal-rotate-right:not(.active)';
    this.document.querySelectorAll(selector).forEach(el => {
      // If already in top viewport, activate directly with smooth entrance
      const rect = el.getBoundingClientRect();
      if (rect.top < window.innerHeight && rect.bottom > 0) {
        el.classList.add('active');
      } else {
        this.observer?.observe(el);
      }
    });
  }

  ngOnDestroy() {
    if (this.observer) {
      this.observer.disconnect();
    }

    if (this.intervalId) {
      clearInterval(this.intervalId);
    }

    if (this.routerSub) {
      this.routerSub.unsubscribe();
    }

    if (isPlatformBrowser(this.platformId)) {
      const savedTheme = localStorage.getItem('theme');
      if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
        this.document.documentElement.classList.add('dark');
      }
    }
  }

  getDashboardUrl(): string {
    if (!isPlatformBrowser(this.platformId)) return '/dashboard';
    
    const user = this.authService.currentUser();
    const subdomain = user?.subdomain;
    if (!subdomain) return '/dashboard';

    const hostname = window.location.hostname;
    const port = window.location.port ? `:${window.location.port}` : '';
    const protocol = window.location.protocol;

    // If we are already on the subdomain, use relative path
    if (hostname.startsWith(`${subdomain}.`)) {
      return '/dashboard';
    }

    // Otherwise, redirect to the subdomain absolute URL
    // Special case: if hostname is "localhost", we don't want "tenant.localhost", 
    // actually we might want "tenant.localhost" if supported by the proxy.
    // If we are on puxbay.com, it goes to tenant.puxbay.com
    let baseHost = hostname;
    if (hostname.startsWith('www.')) {
      baseHost = hostname.substring(4);
    }
    
    return `${protocol}//${subdomain}.${baseHost}${port}/dashboard`;
  }
}
