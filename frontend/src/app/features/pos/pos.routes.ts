import { Routes } from '@angular/router';

export const posRoutes: Routes = [
  {
    path: '',
    loadComponent: () => import('./pos/pos').then(m => m.Pos)
  }
];
