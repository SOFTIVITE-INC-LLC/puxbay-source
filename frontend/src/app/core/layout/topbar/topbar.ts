import { Component, EventEmitter, inject, Output, signal } from '@angular/core';
import { Router } from '@angular/router';
import { CommandPaletteService } from '../../services/command-palette.service';
import { ThemeService } from '../../services/theme.service';
import { OfflineService } from '../../services/offline/offline';
import { NotificationService } from '../../services/notification.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-topbar',
  standalone: true,
  imports: [],
  templateUrl: './topbar.html',
  styleUrl: './topbar.css',
})
export class Topbar {
  @Output() toggleSidebar = new EventEmitter<void>();
  private cmdPalette = inject(CommandPaletteService);
  private themeService = inject(ThemeService);
  public offlineService = inject(OfflineService);
  public notificationService = inject(NotificationService);
  public authService = inject(AuthService);
  private router = inject(Router);

  get username() {
    const user = this.authService.currentUser();
    if (!user) return 'Loading...';
    return user.first_name && user.last_name ? `${user.first_name} ${user.last_name}` : user.username;
  }

  get role() {
    const user = this.authService.currentUser();
    return user?.role ? user.role.charAt(0).toUpperCase() + user.role.slice(1) : 'User';
  }
  isFullscreen = signal(false);

  toggleFullscreen() {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().then(() => {
        this.isFullscreen.set(true);
      });
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen().then(() => {
          this.isFullscreen.set(false);
        });
      }
    }
  }

  onToggleSidebar() {
    this.toggleSidebar.emit();
  }

  openCommandPalette() {
    this.cmdPalette.open();
  }

  toggleTheme() {
    this.themeService.toggleTheme();
  }

  goToNotifications() {
    this.router.navigate(['/notifications']);
  }
}
