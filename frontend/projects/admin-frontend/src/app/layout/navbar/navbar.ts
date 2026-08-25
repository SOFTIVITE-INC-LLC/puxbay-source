import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { LayoutService } from '../layout.service';

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './navbar.html'
})
export class NavbarComponent {
  private authService = inject(AuthService);
  private router = inject(Router);
  public themeService = inject(ThemeService);
  public layout = inject(LayoutService);

  logout() {
    this.authService.logout();
    this.router.navigate(['/login']);
  }
}
