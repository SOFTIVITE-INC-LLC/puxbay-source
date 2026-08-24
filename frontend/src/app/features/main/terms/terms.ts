import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { LegalService, LegalDocument } from '../../../core/services/legal.service';

@Component({
  selector: 'app-terms',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './terms.html',
})
export class Terms implements OnInit {
  private legalService = inject(LegalService);

  doc = signal<LegalDocument | null>(null);
  isLoading = signal(true);
  hasError = signal(false);

  ngOnInit() {
    this.legalService.getLegalDocument('terms').subscribe({
      next: (d) => { this.doc.set(d); this.isLoading.set(false); },
      error: () => { this.hasError.set(true); this.isLoading.set(false); }
    });
  }

  formatDate(dateStr?: string): string {
    if (!dateStr) return '';
    return new Date(dateStr).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
  }
}
