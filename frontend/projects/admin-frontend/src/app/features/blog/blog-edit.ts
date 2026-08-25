import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule, ActivatedRoute, Router } from '@angular/router';
import { BlogService, BlogPost } from '../../services/blog.service';

@Component({
  selector: 'app-blog-edit',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './blog-edit.html',
})
export class BlogEditComponent implements OnInit {
  private blogService = inject(BlogService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);

  isNew = signal(true);
  isLoading = signal(false);
  isSaving = signal(false);
  error = signal<string | null>(null);

  post = signal<BlogPost>({
    title: '',
    slug: '',
    category: '',
    category_color: 'cyan',
    excerpt: '',
    content: '',
    status: 'draft'
  });

  categoryColors = [
    { value: 'cyan', label: 'Cyan', bg: '#06b6d4' },
    { value: 'purple', label: 'Purple', bg: '#a855f7' },
    { value: 'green', label: 'Green', bg: '#22c55e' },
    { value: 'slate', label: 'Slate', bg: '#64748b' },
    { value: 'rose', label: 'Rose', bg: '#f43f5e' },
    { value: 'orange', label: 'Orange', bg: '#f97316' },
    { value: 'indigo', label: 'Indigo', bg: '#6366f1' },
    { value: 'amber', label: 'Amber', bg: '#f59e0b' },
  ];

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    if (id && id !== 'new') {
      this.isNew.set(false);
      this.loadPost(id);
    }
  }

  loadPost(id: string) {
    this.isLoading.set(true);
    this.blogService.getBlogPost(id).subscribe({
      next: (res) => {
        this.post.set(res);
        this.isLoading.set(false);
      },
      error: () => {
        this.error.set('Failed to load blog post.');
        this.isLoading.set(false);
      }
    });
  }

  generateSlug() {
    if (this.post().title && !this.post().id) {
      const slug = this.post().title
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/(^-|-$)+/g, '');
      this.updateField('slug', slug);
    }
  }

  updateField(field: keyof BlogPost, value: string) {
    this.post.update(p => ({ ...p, [field]: value }));
  }

  save() {
    if (!this.post().title || !this.post().slug || !this.post().content) {
      this.error.set('Title, Slug, and Content are required.');
      return;
    }

    this.isSaving.set(true);
    this.error.set(null);

    const obs = this.isNew() 
      ? this.blogService.createBlogPost(this.post())
      : this.blogService.updateBlogPost(this.post().id!, this.post());

    obs.subscribe({
      next: () => {
        this.router.navigate(['/blog']);
      },
      error: (err) => {
        this.isSaving.set(false);
        this.error.set(err?.error?.error || 'Failed to save blog post.');
      }
    });
  }
}
