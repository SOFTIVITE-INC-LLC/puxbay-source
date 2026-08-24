import { Component, inject, OnInit, signal } from '@angular/core';
import { Subject } from 'rxjs';
import { debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

import { FormsModule } from '@angular/forms';
import {
  PortalService,
  PortalConfig,
  PublicProduct,
  Category,
} from '../../../core/services/portal.service';
import { ThemeService } from '../../../core/services/theme.service';

@Component({
  selector: 'app-portal',
  standalone: true,
  imports: [FormsModule, AppCurrencyPipe],
  templateUrl: './portal.html',
})
export class Portal implements OnInit {
  portalService = inject(PortalService);
  private themeService = inject(ThemeService);

  // State from Service
  storeSettings = this.portalService.config;
  products = this.portalService.products;
  categories = this.portalService.categories;
  availableBrands = this.portalService.availableBrands;

  // Local State
  searchQuery = signal<string>('');
  private searchSubject = new Subject<string>();
  cartCount = signal<number>(0);
  branch = signal<{ name: string }>({ name: 'Main Branch' });

  filters = signal({
    min_price: null as number | null,
    max_price: null as number | null,
    brands: [] as string[],
    sort: 'latest',
  });

  ngOnInit() {
    // Load config and initial products
    this.portalService.getConfig().subscribe();
    this.applyFilters();

    // Setup debounced search
    this.searchSubject.pipe(
      debounceTime(300),
      distinctUntilChanged()
    ).subscribe(query => {
      this.searchQuery.set(query);
      this.applyFilters();
    });
  }

  toggleTheme() {
    this.themeService.toggleTheme();
  }

  onSearchInput(event: Event) {
    const target = event.target as HTMLInputElement;
    this.searchSubject.next(target.value);
  }

  toggleBrand(brand: string) {
    const current = this.filters().brands;
    const next = current.includes(brand) ? current.filter((b) => b !== brand) : [...current, brand];

    this.filters.update((f) => ({ ...f, brands: next }));
    this.applyFilters();
  }

  getSubdomain(): string {
    const host = window.location.hostname;
    const parts = host.split('.');
    return parts.length > 2 ? parts[0] : 'demo';
  }

  applyFilters() {
    this.portalService.loadStore(this.getSubdomain());
  }
}
