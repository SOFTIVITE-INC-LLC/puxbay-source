import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface BlogPost {
  id?: string;
  title: string;
  slug: string;
  content: string;
  excerpt?: string;
  category: string;
  category_color: string;
  status: string;
  published_at?: string;
  created_at?: string;
}

export interface BlogListResponse {
  posts: BlogPost[];
}

@Injectable({ providedIn: 'root' })
export class BlogService {
  private http = inject(HttpClient);
  private base = '/api/v1/admin/blog';

  getBlogPosts(): Observable<BlogListResponse> {
    return this.http.get<BlogListResponse>(this.base);
  }

  getBlogPost(id: string): Observable<BlogPost> {
    return this.http.get<BlogPost>(`${this.base}/${id}`);
  }

  createBlogPost(post: BlogPost): Observable<BlogPost> {
    return this.http.post<BlogPost>(this.base, post);
  }

  updateBlogPost(id: string, post: BlogPost): Observable<BlogPost> {
    return this.http.put<BlogPost>(`${this.base}/${id}`, post);
  }

  deleteBlogPost(id: string): Observable<any> {
    return this.http.delete(`${this.base}/${id}`);
  }
}
