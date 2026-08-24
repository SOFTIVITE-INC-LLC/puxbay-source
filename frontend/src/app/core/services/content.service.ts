import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';

export interface ContentPage {
  id: string;
  title: string;
  slug: string;
  status: string; // draft, published
  last_edited: string;
}

@Injectable({
  providedIn: 'root'
})
export class ContentService {
  private api = inject(ApiService);
  
  pages = signal<ContentPage[]>([]);
  loading = signal<boolean>(false);

  getPages(): Observable<{pages: ContentPage[]}> {
    this.loading.set(true);
    return this.api.get<{pages: ContentPage[]}>('/content').pipe(
      tap(res => {
        this.pages.set(res.pages || []);
        this.loading.set(false);
      })
    );
  }
}
