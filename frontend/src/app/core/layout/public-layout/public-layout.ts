import { Component, OnInit, OnDestroy, NgZone, Inject, PLATFORM_ID, afterNextRender, signal } from '@angular/core';
import { isPlatformBrowser, DOCUMENT } from '@angular/common';
import { RouterModule, Router, NavigationStart } from '@angular/router';
import { Subscription } from 'rxjs';
import { filter } from 'rxjs/operators';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-public-layout',
  standalone: true,
  imports: [RouterModule],
  templateUrl: './public-layout.html',
})
export class PublicLayout implements OnInit, OnDestroy {
  private observer: IntersectionObserver | null = null;
  private intervalId: any;
  private routerSub: Subscription | null = null;

  mobileMenuOpen = signal(false);

  constructor(
    private ngZone: NgZone,
    private router: Router,
    @Inject(DOCUMENT) private document: Document,
    @Inject(PLATFORM_ID) private platformId: Object,
    public authService: AuthService
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
    }, { threshold: 0.1 });

    this.observeElements();

    this.ngZone.runOutsideAngular(() => {
      this.intervalId = setInterval(() => this.observeElements(), 1000);
    });
  }

  observeElements() {
    if (!this.observer) return;
    this.document.querySelectorAll('.reveal:not(.active)').forEach(el => {
      this.observer?.observe(el);
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
