import { Component } from '@angular/core';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-not-found',
  standalone: true,
  imports: [RouterModule],
  template: `
    <div style="display:flex; flex-direction:column; align-items:center; justify-content:center; min-height:100vh; background:#0f172a; color:#fff; font-family:'Outfit',sans-serif; text-align:center; padding: 24px;">
      <h1 style="font-size:120px; font-weight:900; line-height:1; margin:0; background:linear-gradient(135deg,#6366f1,#8b5cf6); -webkit-background-clip:text; -webkit-text-fill-color:transparent;">404</h1>
      <h2 style="font-size:24px; font-weight:700; margin:16px 0; letter-spacing:-0.02em;">Page Not Found</h2>
      <p style="font-size:16px; color:#94a3b8; max-width:400px; margin-bottom:32px; line-height:1.6;">
        The page you are looking for doesn't exist or has been moved. Let's get you back on track.
      </p>
      <a routerLink="/" style="display:inline-flex; align-items:center; justify-content:center; padding:14px 24px; background:#6366f1; color:#fff; text-decoration:none; font-weight:700; font-size:14px; border-radius:10px; transition:all 0.2s; gap:8px;">
        <svg style="width:16px; height:16px;" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M10 19l-7-7m0 0l7-7m-7 7h18"/></svg>
        Back to Home
      </a>
    </div>
  `
})
export class NotFoundComponent {}
