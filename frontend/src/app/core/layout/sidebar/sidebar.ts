import { Component, Input, Output, EventEmitter, inject, computed } from '@angular/core';

import { RouterModule, Router } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { BranchService } from '../../../core/services/branch.service';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [RouterModule],
  templateUrl: './sidebar.html',
  styleUrl: './sidebar.css',
})
export class Sidebar {
  authService = inject(AuthService);
  router = inject(Router);
  branchService = inject(BranchService);

  @Input() isCollapsed = false;
  @Input() storeName = 'Puxbay OS';

  userRole = computed(() => {
    return (this.authService.currentUser()?.role || 'admin').toLowerCase();
  });

  @Output() toggleCollapse = new EventEmitter<void>();

  onToggle() {
    this.toggleCollapse.emit();
  }

  logout() {
    this.authService.logout();
    this.router.navigate(['/login']);
  }

  exitBranch() {
    this.branchService.setActiveBranch(null);
    this.router.navigate(['/branches']);
  }
}
