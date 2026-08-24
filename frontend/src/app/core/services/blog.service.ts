import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface BlogPost {
  id: string;
  title: string;
  slug: string;
  content: string;
  excerpt?: string;
  category: string;
  category_color: string;
  status: string;
  published_at?: string;
}

export interface BlogListResponse {
  posts: BlogPost[];
}

@Injectable({ providedIn: 'root' })
export class BlogService {
  private http = inject(HttpClient);
  private base = `${environment.apiUrl}/public/blog`;

  getBlogPosts(): Observable<BlogListResponse> {
    return this.http.get<BlogListResponse>(this.base);
  }

  getBlogPost(slug: string): Observable<BlogPost> {
    return this.http.get<BlogPost>(`${this.base}/${slug}`);
  }
}
