import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CategoryService, CategoryCreateInput } from '../../../core/services/category.service';
import { Category } from '../../../core/models/product.models';

// Curated swatch palette
export const COLOR_SWATCHES = [
  '#ef4444', // red
  '#f97316', // orange
  '#f59e0b', // amber
  '#84cc16', // lime
  '#22c55e', // green
  '#14b8a6', // teal
  '#06b6d4', // cyan
  '#3b82f6', // blue
  '#6366f1', // indigo
  '#8b5cf6', // violet
  '#ec4899', // pink
  '#64748b', // slate
];

@Component({
  selector: 'app-categories',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './categories.html',
})
export class Categories implements OnInit {
  categoryService = inject(CategoryService);

  isDrawerOpen = signal(false);
  isDeleteModalOpen = signal(false);
  saving = signal(false);

  editingId = signal<string | null>(null);
  categoryForm = signal<CategoryCreateInput>({ name: '', description: '', color: '#6366f1' });
  categoryToDelete = signal<Category | null>(null);

  searchQuery = signal('');

  readonly swatches = COLOR_SWATCHES;

  filteredCategories = computed(() => {
    const q = this.searchQuery().toLowerCase().trim();
    const cats = this.categoryService.categories();
    if (!q) return cats;
    return cats.filter((c: any) =>
      c.name.toLowerCase().includes(q) || (c.description || '').toLowerCase().includes(q)
    );
  });

  ngOnInit() {
    this.categoryService.getCategories().subscribe();
  }

  openCreateDrawer() {
    this.editingId.set(null);
    this.categoryForm.set({ name: '', description: '', color: '#6366f1' });
    this.isDrawerOpen.set(true);
  }

  openEditDrawer(cat: any) {
    this.editingId.set(cat.id);
    this.categoryForm.set({ name: cat.name, description: cat.description || '', color: cat.color || '#6366f1' });
    this.isDrawerOpen.set(true);
  }

  closeDrawer() {
    this.isDrawerOpen.set(false);
  }

  pickColor(hex: string) {
    this.categoryForm.update(f => ({ ...f, color: hex }));
  }

  saveCategory() {
    const form = this.categoryForm();
    if (!form.name.trim()) return;

    this.saving.set(true);
    const id = this.editingId();

    const request$ = id
      ? this.categoryService.updateCategory(id, form)
      : this.categoryService.createCategory(form);

    request$.subscribe({
      next: () => {
        this.saving.set(false);
        this.closeDrawer();
      },
      error: () => this.saving.set(false)
    });
  }

  confirmDelete(cat: Category) {
    this.categoryToDelete.set(cat);
    this.isDeleteModalOpen.set(true);
  }

  deleteCategory() {
    const cat = this.categoryToDelete();
    if (!cat) return;

    this.saving.set(true);
    this.categoryService.deleteCategory(cat.id).subscribe({
      next: () => {
        this.saving.set(false);
        this.isDeleteModalOpen.set(false);
        this.categoryToDelete.set(null);
      },
      error: () => this.saving.set(false)
    });
  }

  // Derive a readable foreground color (white/dark) for a given hex bg
  textColorFor(hex: string): string {
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    // Perceived luminance
    const lum = 0.299 * r + 0.587 * g + 0.114 * b;
    return lum > 160 ? '#1e293b' : '#ffffff';
  }
}
