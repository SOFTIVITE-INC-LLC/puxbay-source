import { Component, OnInit, ViewEncapsulation, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { BlogService, BlogPost } from '../../../core/services/blog.service';

@Component({
  selector: 'app-blog',
  standalone: true,
  imports: [RouterModule, CommonModule],
  templateUrl: './blog.html',
  encapsulation: ViewEncapsulation.None,
})
export class Blog implements OnInit {
  private blogService = inject(BlogService);

  posts = signal<BlogPost[]>([]);
  isLoading = signal(true);
  hasError = signal(false);

  ngOnInit() {
    this.blogService.getBlogPosts().subscribe({
      next: (res) => {
        this.posts.set(res.posts || []);
        this.isLoading.set(false);
      },
      error: () => {
        this.hasError.set(true);
        this.isLoading.set(false);
      }
    });
  }

  formatDate(dateStr?: string): string {
    if (!dateStr) return '';
    return new Date(dateStr).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
  }
}
