import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SecurityService } from '../../services/security.service';

@Component({
  selector: 'app-api-keys',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './api-keys.html',
})
export class ApiKeysComponent implements OnInit {
  private service = inject(SecurityService);
  keys = signal<any[]>([]);
  isLoading = signal(true);
  isCreating = signal(false);

  newKeyName = signal('');
  isModalOpen = signal(false);
  newlyCreatedKey = signal<string | null>(null);

  ngOnInit() {
    this.loadKeys();
  }

  loadKeys() {
    this.isLoading.set(true);
    this.service.getAPIKeys().subscribe({
      next: (res) => { this.keys.set(res.data || []); this.isLoading.set(false); },
      error: () => this.isLoading.set(false)
    });
  }

  openCreateModal() {
    this.newKeyName.set('');
    this.newlyCreatedKey.set(null);
    this.isModalOpen.set(true);
  }

  closeCreateModal() {
    this.isModalOpen.set(false);
    this.newlyCreatedKey.set(null);
  }

  createKey() {
    if (!this.newKeyName()) return;
    this.isCreating.set(true);
    this.service.createAPIKey(this.newKeyName()).subscribe({
      next: (res) => {
        this.newlyCreatedKey.set(res.key);
        this.isCreating.set(false);
        this.loadKeys();
      },
      error: () => this.isCreating.set(false)
    });
  }

  revokeKey(id: string) {
    if (confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) {
      this.service.revokeAPIKey(id).subscribe({
        next: () => this.loadKeys()
      });
    }
  }

  copyToClipboard(val: string) {
    navigator.clipboard.writeText(val);
    alert('Copied to clipboard!');
  }
}
