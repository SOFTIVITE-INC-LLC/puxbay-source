import { Injectable, effect, signal } from '@angular/core';

export type ThemeMode = 'light' | 'dark' | 'system';

@Injectable({
  providedIn: 'root'
})
export class ThemeService {
  theme = signal<ThemeMode>('system');
  isDark = signal<boolean>(false);

  constructor() {
    // Load saved theme
    const saved = localStorage.getItem('admin-theme') as ThemeMode;
    if (saved) {
      this.theme.set(saved);
    }

    // Effect to apply theme to HTML document
    effect(() => {
      const current = this.theme();
      localStorage.setItem('admin-theme', current);
      
      let applyDark = false;
      if (current === 'system') {
        applyDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      } else {
        applyDark = current === 'dark';
      }

      this.isDark.set(applyDark);

      if (applyDark) {
        document.documentElement.classList.add('dark');
        document.documentElement.setAttribute('data-theme', 'dark'); // For DaisyUI
      } else {
        document.documentElement.classList.remove('dark');
        document.documentElement.setAttribute('data-theme', 'light');
      }
    });

    // Listen for system theme changes
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      if (this.theme() === 'system') {
        const applyDark = e.matches;
        this.isDark.set(applyDark);
        if (applyDark) {
          document.documentElement.classList.add('dark');
          document.documentElement.setAttribute('data-theme', 'dark');
        } else {
          document.documentElement.classList.remove('dark');
          document.documentElement.setAttribute('data-theme', 'light');
        }
      }
    });
  }

  setTheme(mode: ThemeMode) {
    this.theme.set(mode);
  }

  toggleTheme() {
    if (this.isDark()) {
      this.theme.set('light');
    } else {
      this.theme.set('dark');
    }
  }
}
