import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { BlogService, BlogPost } from '../../services/blog.service';

@Component({
  selector: 'app-blog-list',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './blog-list.html',
})
export class BlogListComponent implements OnInit {
  private blogService = inject(BlogService);

  posts = signal<BlogPost[]>([]);
  isLoading = signal(true);
  error = signal<string | null>(null);

  ngOnInit() {
    this.loadPosts();
  }

  loadPosts() {
    this.isLoading.set(true);
    this.blogService.getBlogPosts().subscribe({
      next: (res) => {
        this.posts.set(res.posts || []);
        this.isLoading.set(false);
      },
      error: () => {
        this.error.set('Failed to load blog posts.');
        this.isLoading.set(false);
      }
    });
  }

  deletePost(id: string) {
    if (confirm('Are you sure you want to delete this blog post?')) {
      this.blogService.deleteBlogPost(id).subscribe({
        next: () => {
          this.loadPosts();
        },
        error: () => {
          alert('Failed to delete post.');
        }
      });
    }
  }

  formatDate(dateStr?: string): string {
    if (!dateStr) return 'Not Published';
    return new Date(dateStr).toLocaleDateString();
  }
}
