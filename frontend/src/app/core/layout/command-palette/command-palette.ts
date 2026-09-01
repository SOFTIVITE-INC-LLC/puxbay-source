import { Component, HostListener, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { CommandPaletteService } from '../../services/command-palette.service';

interface Command {
  name: string;
  description: string;
  icon: string;
  route: string;
  category: string;
}

@Component({
  selector: 'app-command-palette',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './command-palette.html'
})
export class CommandPalette {
  private cmdService = inject(CommandPaletteService);
  isOpen = this.cmdService.isOpen;
  searchQuery = '';

  readonly allCommands: Command[] = [
    // Navigation
    { name: 'Dashboard', description: 'View your business overview', icon: 'dashboard', route: '/dashboard', category: 'Navigation' },
    { name: 'Orders', description: 'Manage all customer orders', icon: 'receipt_long', route: '/orders', category: 'Navigation' },
    { name: 'Inventory', description: 'Track products and stock levels', icon: 'inventory_2', route: '/inventory', category: 'Navigation' },
    { name: 'Purchase Orders', description: 'Manage supplier orders', icon: 'local_shipping', route: '/inventory/supply-chain', category: 'Navigation' },
    { name: 'Customers', description: 'CRM and customer management', icon: 'group', route: '/customers', category: 'Navigation' },
    { name: 'Financial', description: 'Revenue, expenses, and accounting', icon: 'account_balance', route: '/financial', category: 'Navigation' },
    { name: 'Staff & HR', description: 'Manage your team and payroll', icon: 'badge', route: '/hr', category: 'Navigation' },
    { name: 'Intelligence', description: 'AI-powered analytics and insights', icon: 'insights', route: '/intelligence', category: 'Navigation' },
    { name: 'Reports', description: 'Generate business reports', icon: 'bar_chart', route: '/reports', category: 'Navigation' },
    { name: 'Branches', description: 'Manage multi-location branches', icon: 'store', route: '/branches', category: 'Navigation' },
    { name: 'Marketing', description: 'Campaigns and promotions', icon: 'campaign', route: '/marketing', category: 'Navigation' },
    { name: 'Point of Sale', description: 'Launch the POS terminal', icon: 'point_of_sale', route: '/pos', category: 'Navigation' },
    { name: 'Gift Cards', description: 'Manage gift card programs', icon: 'card_giftcard', route: '/gift-cards', category: 'Navigation' },
    { name: 'Suppliers', description: 'Manage your supplier network', icon: 'handshake', route: '/suppliers', category: 'Navigation' },
    { name: 'Returns', description: 'Process order returns', icon: 'assignment_return', route: '/returns', category: 'Navigation' },
    { name: 'Storefront', description: 'Manage your online store', icon: 'storefront', route: '/storefront', category: 'Navigation' },
    // Settings
    { name: 'My Profile', description: 'Update your profile and preferences', icon: 'manage_accounts', route: '/profile', category: 'Settings' },
    { name: 'Branch Settings', description: 'Configure branch details', icon: 'settings', route: '/branch-settings', category: 'Settings' },
    { name: 'Integrations', description: 'Connect third-party services', icon: 'extension', route: '/integrations', category: 'Settings' },
    { name: 'Security', description: 'Manage security settings', icon: 'security', route: '/security', category: 'Settings' },
    { name: 'Webhooks', description: 'Configure webhook endpoints', icon: 'webhook', route: '/webhooks', category: 'Settings' },
  ];

  filteredCommands = signal<Command[]>(this.allCommands);

  constructor(private router: Router) {}

  @HostListener('window:keydown', ['$event'])
  handleKeyboardEvent(event: KeyboardEvent) {
    if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
      event.preventDefault();
      this.togglePalette();
    }
    if (event.key === 'Escape' && this.isOpen()) {
      this.closePalette();
    }
  }

  togglePalette() {
    this.cmdService.toggle();
    if (this.isOpen()) {
      this.searchQuery = '';
      this.filteredCommands.set(this.allCommands);
      setTimeout(() => {
        document.getElementById('cmd-palette-input')?.focus();
      }, 50);
    }
  }

  closePalette() {
    this.cmdService.close();
  }

  onSearchChange() {
    const q = this.searchQuery.toLowerCase().trim();
    if (!q) {
      this.filteredCommands.set(this.allCommands);
      return;
    }
    this.filteredCommands.set(
      this.allCommands.filter(c =>
        c.name.toLowerCase().includes(q) ||
        c.description.toLowerCase().includes(q) ||
        c.category.toLowerCase().includes(q)
      )
    );
  }

  groupedCategories(): string[] {
    const cats = [...new Set(this.filteredCommands().map(c => c.category))];
    return cats;
  }

  commandsByCategory(category: string): Command[] {
    return this.filteredCommands().filter(c => c.category === category);
  }

  executeCommand(route: string) {
    this.closePalette();
    this.router.navigate([route]);
  }
}

