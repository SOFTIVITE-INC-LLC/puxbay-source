import { RenderMode, ServerRoute } from '@angular/ssr';

export const serverRoutes: ServerRoute[] = [
  // Public marketing & legal pages
  { path: '', renderMode: RenderMode.Server },
  { path: 'features', renderMode: RenderMode.Server },
  { path: 'solutions', renderMode: RenderMode.Server },
  { path: 'pricing', renderMode: RenderMode.Server },
  { path: 'about', renderMode: RenderMode.Server },
  { path: 'careers', renderMode: RenderMode.Server },
  { path: 'blog', renderMode: RenderMode.Server },
  { path: 'contact', renderMode: RenderMode.Server },
  { path: 'privacy-policy', renderMode: RenderMode.Server },
  { path: 'terms', renderMode: RenderMode.Server },
  { path: 'cookie-policy', renderMode: RenderMode.Server },
  { path: 'product/**', renderMode: RenderMode.Server },
  { path: 'storefront/**', renderMode: RenderMode.Server },
  { path: 'store/**', renderMode: RenderMode.Server },
  // All internal dashboard, notifications, POS, auth, and protected routes use Client rendering
  {
    path: '**',
    renderMode: RenderMode.Client,
  },
];
