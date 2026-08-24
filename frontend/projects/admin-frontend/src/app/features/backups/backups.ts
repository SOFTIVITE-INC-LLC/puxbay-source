import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { BackupService, DatabaseBackup } from '../../services/backup.service';

@Component({
  selector: 'app-backups',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './backups.html',
})
export class BackupsComponent implements OnInit {
  private service = inject(BackupService);

  backups = signal<DatabaseBackup[]>([]);
  isLoading = signal(true);

  ngOnInit() {
    this.loadBackups();
  }

  loadBackups() {
    this.isLoading.set(true);
    this.service.getBackups().subscribe({
      next: (data) => {
        this.backups.set(data || []);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load backups', err);
        this.isLoading.set(false);
      }
    });
  }

  formatBytes(bytes: number, decimals = 2) {
    if (!+bytes) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
  }

  triggerBackup() {
    this.isLoading.set(true);
    this.service.triggerBackup().subscribe({
      next: () => {
        this.loadBackups();
      },
      error: (err) => {
        console.error('Failed to trigger backup', err);
        this.isLoading.set(false);
      }
    });
  }

  downloadBackup(backup: DatabaseBackup) {
    this.service.downloadBackup(backup.id).subscribe({
      next: (blob) => {
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = backup.filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
      },
      error: (err) => {
        console.error('Failed to download backup', err);
      }
    });
  }
}
