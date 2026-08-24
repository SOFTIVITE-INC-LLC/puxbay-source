import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { LegalService, LegalDocument, LegalDocType } from '../../services/legal.service';

type Tab = { type: LegalDocType; label: string; icon: string; description: string };

@Component({
  selector: 'app-legal-documents',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './legal-documents.html',
})
export class LegalDocumentsComponent implements OnInit {
  private service = inject(LegalService);

  isLoading = signal(true);
  isSaving = signal(false);
  saveSuccess = signal<LegalDocType | null>(null);
  saveError = signal<string | null>(null);

  tabs: Tab[] = [
    { type: 'terms',   label: 'Terms of Service', icon: 'gavel',    description: 'Governs use of the Puxbay platform.' },
    { type: 'privacy', label: 'Privacy Policy',   icon: 'lock',     description: 'Explains how user data is collected and used.' },
    { type: 'cookie',  label: 'Cookie Policy',    icon: 'cookie',   description: 'Describes the cookies used on the site.' },
  ];

  activeTab = signal<LegalDocType>('terms');

  docs = signal<Record<LegalDocType, LegalDocument>>({
    terms:   { type: 'terms',   title: 'Terms of Service', content: '', version: '1.0' },
    privacy: { type: 'privacy', title: 'Privacy Policy',   content: '', version: '1.0' },
    cookie:  { type: 'cookie',  title: 'Cookie Policy',    content: '', version: '1.0' },
  });

  currentDoc = computed(() => this.docs()[this.activeTab()]);

  charCount = computed(() => this.currentDoc().content.length);

  ngOnInit() {
    this.loadDocuments();
  }

  loadDocuments() {
    this.isLoading.set(true);
    this.service.getLegalDocuments().subscribe({
      next: (res) => {
        const map = { ...this.docs() };
        (res.documents || []).forEach(doc => {
          if (doc.type in map) {
            map[doc.type] = { ...map[doc.type], ...doc };
          }
        });
        this.docs.set(map);
        this.isLoading.set(false);
      },
      error: () => {
        this.isLoading.set(false);
      }
    });
  }

  setTab(type: LegalDocType) {
    this.activeTab.set(type);
    this.saveSuccess.set(null);
    this.saveError.set(null);
  }

  updateField(field: keyof LegalDocument, value: string) {
    const type = this.activeTab();
    this.docs.update(all => ({
      ...all,
      [type]: { ...all[type], [field]: value }
    }));
  }

  save() {
    const type = this.activeTab();
    const doc = this.currentDoc();

    if (!doc.content.trim()) {
      this.saveError.set('Content cannot be empty.');
      return;
    }

    this.isSaving.set(true);
    this.saveError.set(null);

    const payload = {
      title: doc.title,
      content: doc.content,
      version: doc.version,
      effective_date: doc.effective_date ? doc.effective_date.split('T')[0] : undefined,
    };

    this.service.upsertLegalDocument(type, payload).subscribe({
      next: (updated) => {
        this.docs.update(all => ({ ...all, [type]: { ...all[type], ...updated } }));
        this.isSaving.set(false);
        this.saveSuccess.set(type);
        setTimeout(() => this.saveSuccess.set(null), 3000);
      },
      error: (err) => {
        this.isSaving.set(false);
        this.saveError.set(err?.error?.error || 'Failed to save. Please try again.');
      }
    });
  }

  formatDate(dateStr?: string): string {
    if (!dateStr) return 'Never saved';
    return new Date(dateStr).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
  }

  getLastUpdated(type: LegalDocType): string {
    return this.formatDate(this.docs()[type]?.updated_at);
  }

  hasContent(type: LegalDocType): boolean {
    return !!this.docs()[type]?.content?.trim();
  }
}
