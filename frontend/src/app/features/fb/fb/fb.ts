import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { FNBService, DiningTable, KDSTicket } from '../../../core/services/fnb.service';
import { CatalogService } from '../../../core/services/catalog.service';
import { CategoryService } from '../../../core/services/category.service';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-fb',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './fb.html',
})
export class Fb implements OnInit {
  fnbService = inject(FNBService);
  catalogService = inject(CatalogService);
  categoryService = inject(CategoryService);
  private toastr = inject(ToastrService);

  activeTab = signal<'tables' | 'kds' | 'menu'>('tables');
  isModalOpen = signal(false);
  isMenuModalOpen = signal(false);
  isCategoryModalOpen = signal(false);
  isUpdatingStatus = signal<string | null>(null);

  currentTable = signal<Partial<DiningTable>>({ capacity: 4, status: 'available' });
  currentMenuItem = signal<any>({ name: '', selling_price: 0, track_inventory: false, current_stock: 0, stock_unit: 'plate' });
  currentCategory = signal<any>({ name: '', color: 'orange' });

  ngOnInit() {
    this.fnbService.getTables().subscribe();
    this.fnbService.getKDS().subscribe();
    this.categoryService.getCategories().subscribe();
    this.catalogService.getProducts().subscribe();
  }

  openAddModal() {
    this.currentTable.set({ capacity: 4, status: 'available' });
    this.isModalOpen.set(true);
  }

  closeModal() { this.isModalOpen.set(false); }

  openMenuCategoryModal() {
    this.currentCategory.set({ name: '', color: 'orange' });
    this.isCategoryModalOpen.set(true);
  }
  closeCategoryModal() { this.isCategoryModalOpen.set(false); }

  openMenuItemModal() {
    this.currentMenuItem.set({ name: '', selling_price: 0, track_inventory: false, current_stock: 0, stock_unit: 'plate', sku: 'FNB-' + Math.floor(Math.random()*10000) });
    this.isMenuModalOpen.set(true);
  }
  closeMenuModal() { this.isMenuModalOpen.set(false); }

  saveCategory() {
    const c = this.currentCategory();
    if (!c.name) { this.toastr.error('Category name is required'); return; }
    this.categoryService.createCategory(c).subscribe({
      next: () => { this.toastr.success('Menu Category added!'); this.closeCategoryModal(); },
      error: () => this.toastr.error('Failed to add category')
    });
  }

  saveMenuItem() {
    const item = this.currentMenuItem();
    if (!item.name || !item.selling_price) { this.toastr.error('Name and price are required'); return; }
    this.catalogService.createProduct(item).subscribe({
      next: () => {
        this.toastr.success('Menu Item added!');
        this.closeMenuModal();
        this.catalogService.getProducts().subscribe();
      },
      error: () => this.toastr.error('Failed to add menu item')
    });
  }

  saveTable() {
    const t = this.currentTable();
    if (!t.name) { this.toastr.error('Table name is required'); return; }
    this.fnbService.createTable(t).subscribe({
      next: () => {
        this.toastr.success('Table added!');
        this.closeModal();
        this.fnbService.getTables().subscribe();
      },
      error: () => this.toastr.error('Failed to add table')
    });
  }

  setTableStatus(table: DiningTable, status: string) {
    this.isUpdatingStatus.set(table.id);
    this.fnbService.updateTableStatus(table.id, status).subscribe({
      next: () => { this.isUpdatingStatus.set(null); this.toastr.success('Status updated'); },
      error: () => { this.isUpdatingStatus.set(null); this.toastr.error('Failed to update status'); }
    });
  }

  advanceTicket(ticket: KDSTicket) {
    this.fnbService.advanceTicketStatus(ticket.id).subscribe({
      next: () => this.toastr.success('Ticket advanced!'),
      error: () => this.toastr.error('Failed to advance ticket')
    });
  }

  statusColor(status: string): string {
    return { available: 'bg-emerald-500', occupied: 'bg-rose-500', reserved: 'bg-amber-500', cleaning: 'bg-sky-500' }[status] ?? 'bg-zinc-400';
  }

  statusBadge(status: string): string {
    return {
      available: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
      occupied: 'bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400',
      reserved: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
      cleaning: 'bg-sky-100 text-sky-700 dark:bg-sky-900/30 dark:text-sky-400',
    }[status] ?? 'bg-zinc-100 text-zinc-600';
  }

  ticketStatusBadge(status: string): string {
    return {
      pending: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
      preparing: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
      ready: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300',
      served: 'bg-zinc-100 text-zinc-500 dark:bg-zinc-800 dark:text-zinc-400',
    }[status] ?? 'bg-zinc-100 text-zinc-500';
  }

  nextAction(status: string): string {
    return { pending: 'Start Preparing', preparing: 'Mark Ready', ready: 'Mark Served' }[status] ?? '';
  }

  kdsTicketsByStatus(status: string): KDSTicket[] {
    return this.fnbService.tickets().filter(t => t.status === status);
  }
}
