import { Injectable, signal } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ThemeService {
  isDark = signal<boolean>(false);

  constructor() {
    this.initTheme();
  }

  private initTheme() {
    if (typeof document === 'undefined') return;
    // The initial theme is set by the inline script in index.html to avoid FOUC
    const isDark = document.documentElement.classList.contains('dark');
    this.isDark.set(isDark);
  }

  toggleTheme() {
    if (typeof document === 'undefined') return;
    const isDark = document.documentElement.classList.toggle('dark');
    this.isDark.set(isDark);
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('theme', isDark ? 'dark' : 'light');
    }
  }
}
