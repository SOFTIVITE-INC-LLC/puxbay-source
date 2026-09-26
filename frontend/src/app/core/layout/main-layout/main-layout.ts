import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../services/auth.service';
import { BranchService } from '../../services/branch.service';
import { HrService } from '../../services/hr.service';
import { ToastrService } from 'ngx-toastr';

import { Router, RouterModule, NavigationEnd } from '@angular/router';
import { filter } from 'rxjs/operators';
import { Sidebar } from '../sidebar/sidebar';
import { Topbar } from '../topbar/topbar';
import { ToastComponent } from '../toast/toast';
import { CommandPalette } from '../command-palette/command-palette';
import { Copilot } from '../copilot/copilot';

@Component({
  selector: 'app-main-layout',
  standalone: true,
  imports: [CommonModule, RouterModule, Sidebar, Topbar, ToastComponent, CommandPalette, Copilot],
  templateUrl: './main-layout.html',
  styleUrl: './main-layout.css',
})
export class MainLayout implements OnInit {
  authService = inject(AuthService);
  branchService = inject(BranchService);
  hrService = inject(HrService);
  toastr = inject(ToastrService);
  router = inject(Router);

  isSidebarCollapsed = typeof window !== 'undefined' ? window.innerWidth < 768 : false;

  isImpersonating = this.authService.isImpersonating;
  currentUser = this.authService.currentUser;

  hasCheckedAttendance = signal(false);
  isClockingIn = signal(false);

  isClockedIn = computed(() => this.hrService.myAttendance().is_clocked_in);

  isClockInRequired = computed(() => {
    const user = this.currentUser();
    if (!user || user.role === 'customer' || user.role === 'supplier' || this.isImpersonating()) {
      return false;
    }
    return this.hasCheckedAttendance() && !this.isClockedIn();
  });

  currentTime = signal(new Date());

  ngOnInit() {
    this.router.events.pipe(
      filter((event): event is NavigationEnd => event instanceof NavigationEnd)
    ).subscribe(() => {
      this.closeSidebar();
    });

    if (this.authService.isAuthenticated()) {
      this.hrService.getMyAttendanceStatus().subscribe({
        next: () => this.hasCheckedAttendance.set(true),
        error: () => this.hasCheckedAttendance.set(true)
      });
    }

    if (typeof window !== 'undefined') {
      setInterval(() => {
        this.currentTime.set(new Date());
      }, 1000);
    }
  }

  toggleSidebar() {
    this.isSidebarCollapsed = !this.isSidebarCollapsed;
  }

  closeSidebar() {
    if (typeof window !== 'undefined' && window.innerWidth < 768) {
      this.isSidebarCollapsed = true;
    }
  }

  exitImpersonation() {
    this.authService.logout();
    window.close(); // Close the support mode tab
  }

  clockInFromGateway() {
    this.isClockingIn.set(true);
    this.hrService.clockIn().subscribe({
      next: () => {
        this.isClockingIn.set(false);
        this.toastr.success('You have clocked in! Welcome to your shift.', 'Workspace Unlocked');
      },
      error: (err: any) => {
        this.isClockingIn.set(false);
        this.toastr.error(err?.error?.error || 'Failed to clock in. Please try again.');
      }
    });
  }

  logout() {
    this.authService.logout();
    this.router.navigate(['/auth/login']);
  }
}

