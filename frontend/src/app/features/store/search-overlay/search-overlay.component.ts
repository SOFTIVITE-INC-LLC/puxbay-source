import { Component, EventEmitter, inject, Output, signal, ViewChild, ElementRef, AfterViewInit, ChangeDetectionStrategy, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, debounceTime, distinctUntilChanged, switchMap, filter, catchError, of, tap } from 'rxjs';
import { ProductService } from '../../../core/store/services/product.service';
import { Product } from '../../../core/store/models/product.model';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-search-overlay',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe],
  templateUrl: './search-overlay.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class SearchOverlayComponent implements AfterViewInit {
  @Output() closeOverlay = new EventEmitter<void>();
  @ViewChild('searchInput') searchInput!: ElementRef<HTMLInputElement>;
  
  private productService = inject(ProductService);
  private destroyRef = inject(DestroyRef);

  searchQuery = signal('');
  results = signal<Product[]>([]);
  isLoading = signal(false);
  hasSearched = signal(false);

  private searchSubject = new Subject<string>();

  constructor() {
    this.searchSubject.pipe(
      debounceTime(300),
      distinctUntilChanged(),
      tap(query => {
        if (!query.trim()) {
          this.results.set([]);
          this.hasSearched.set(false);
          this.isLoading.set(false);
        } else {
          this.isLoading.set(true);
          this.hasSearched.set(true);
        }
      }),
      filter(query => !!query.trim()),
      switchMap(query => this.productService.searchProducts(`search=${query}&page_size=6`).pipe(
        catchError(() => of({ products: [] }))
      )),
      takeUntilDestroyed(this.destroyRef)
    ).subscribe(res => {
      this.results.set(res.products || []);
      this.isLoading.set(false);
    });
  }

  ngAfterViewInit() {
    setTimeout(() => {
      this.searchInput.nativeElement.focus();
    }, 100);
  }

  onSearch(query: string) {
    this.searchQuery.set(query);
    this.searchSubject.next(query);
  }

  close() {
    this.closeOverlay.emit();
  }
}
