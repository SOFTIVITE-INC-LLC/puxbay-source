import { Injectable, signal } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class LayoutService {
  isMobileSidebarOpen = signal(false);

  toggleSidebar() {
    this.isMobileSidebarOpen.set(!this.isMobileSidebarOpen());
  }

  closeSidebar() {
    this.isMobileSidebarOpen.set(false);
  }
}
