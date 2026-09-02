import { Pipe, PipeTransform, inject, PLATFORM_ID } from '@angular/core';
import { environment } from '../../../environments/environment';

export const DEFAULT_PRODUCT_PLACEHOLDER = `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="400" height="400" viewBox="0 0 400 400" fill="none"><rect width="400" height="400" fill="%23f8fafc"/><g opacity="0.45" transform="translate(140, 130)"><rect x="10" y="25" width="100" height="80" rx="14" fill="%2394a3b8"/><path d="M60 25 L60 105 M10 58 L110 58" stroke="%23f1f5f9" stroke-width="3.5" stroke-dasharray="4 3"/><path d="M32 25 L48 10 L72 10 L88 25" fill="%2364748b"/></g><text x="200" y="260" text-anchor="middle" fill="%2394a3b8" font-family="system-ui, -apple-system, sans-serif" font-size="13" font-weight="700" letter-spacing="1">NO IMAGE</text></svg>`;

/**
 * Resolves a product/logo image URL to an absolute URL or sleek default placeholder SVG.
 *
 * - If the URL is empty/null, returns the default crisp vector placeholder SVG.
 * - If the URL is already absolute (http/https/data:), it is returned unchanged.
 * - If the URL is relative (e.g. `/uploads/products/...`), it is resolved
 *   against the API origin in production, or left relative in development
 *   (the Angular dev proxy forwards /uploads → localhost:5000).
 */
@Pipe({
  name: 'imageUrl',
  standalone: true,
})
export class ImageUrlPipe implements PipeTransform {
  transform(value: string | null | undefined, fallback: boolean = true): string {
    if (!value || value.trim() === '') {
      return fallback ? DEFAULT_PRODUCT_PLACEHOLDER : '';
    }
    const cleanVal = value.trim();

    // Already absolute or data URI — return as-is
    if (cleanVal.startsWith('http://') || cleanVal.startsWith('https://') || cleanVal.startsWith('data:')) {
      return cleanVal;
    }

    if (environment.production) {
      // In production the API origin is different from the app origin.
      // Strip /api/v1 suffix to get the bare API host.
      const apiOrigin = environment.apiUrl.replace(/\/api\/v1$/, '');
      const path = cleanVal.startsWith('/') ? cleanVal : '/' + cleanVal;
      return apiOrigin + path;
    }

    // In development the Angular dev server proxy forwards /uploads to the backend.
    return cleanVal.startsWith('/') ? cleanVal : '/' + cleanVal;
  }
}

/**
 * Standalone utility function — same logic as ImageUrlPipe.transform().
 * Use this in component TypeScript code instead of injecting the pipe.
 */
export function resolveImageUrl(value: string | null | undefined, fallback = false): string {
  if (!value || value.trim() === '') {
    return fallback ? DEFAULT_PRODUCT_PLACEHOLDER : '';
  }
  const cleanVal = value.trim();
  if (cleanVal.startsWith('http://') || cleanVal.startsWith('https://') || cleanVal.startsWith('data:')) {
    return cleanVal;
  }
  if (environment.production) {
    const apiOrigin = environment.apiUrl.replace(/\/api\/v1$/, '');
    const path = cleanVal.startsWith('/') ? cleanVal : '/' + cleanVal;
    return apiOrigin + path;
  }
  return cleanVal.startsWith('/') ? cleanVal : '/' + cleanVal;
}
