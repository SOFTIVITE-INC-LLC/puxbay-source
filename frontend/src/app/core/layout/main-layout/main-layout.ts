import { Component, inject, effect } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { BranchService } from '../../services/branch.service';

import { Router, RouterModule } from '@angular/router';
import { Sidebar } from '../sidebar/sidebar';
import { Topbar } from '../topbar/topbar';
import { ToastComponent } from '../toast/toast';
import { CommandPalette } from '../command-palette/command-palette';
import { Copilot } from '../copilot/copilot';
import { AlertComponent } from '../alert/alert';

@Component({
  selector: 'app-main-layout',
  standalone: true,
  imports: [RouterModule, Sidebar, Topbar, ToastComponent, CommandPalette, Copilot, AlertComponent],
  templateUrl: './main-layout.html',
  styleUrl: './main-layout.css',
})
export class MainLayout {
  authService = inject(AuthService);
  branchService = inject(BranchService);
  router = inject(Router);

  isSidebarCollapsed = typeof window !== 'undefined' ? window.innerWidth < 768 : false;

  isImpersonating = this.authService.isImpersonating;
  currentUser = this.authService.currentUser;

  constructor() {
  }

  toggleSidebar() {
    this.isSidebarCollapsed = !this.isSidebarCollapsed;
  }

  exitImpersonation() {
    this.authService.logout();
    window.close(); // Close the support mode tab
  }
}
