import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { BlogService, BlogPost } from '../../services/blog.service';
import { AlertService } from '../../services/alert.service';

@Component({
  selector: 'app-blog-list',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './blog-list.html',
})
export class BlogListComponent implements OnInit {
  private blogService = inject(BlogService);
  private alert = inject(AlertService);

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

  async deletePost(id: string) {
    const confirmed = await this.alert.confirm({
      title: 'Delete Blog Post',
      message: 'Are you sure you want to permanently delete this blog post? This cannot be undone.',
      confirmText: 'Delete Post',
      cancelText: 'Cancel',
      type: 'danger'
    });
    if (confirmed) {
      this.blogService.deleteBlogPost(id).subscribe({
        next: () => {
          this.alert.success('Blog post deleted.');
          this.loadPosts();
        },
        error: () => this.alert.error('Failed to delete blog post.')
      });
    }
  }

  formatDate(dateStr?: string): string {
    if (!dateStr) return 'Not Published';
    return new Date(dateStr).toLocaleDateString();
  }
}
