import { Component, inject, OnInit, computed, signal, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TelemetryService } from '../../services/telemetry.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-telemetry',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './telemetry.html'
})
export class TelemetryComponent implements OnInit, OnDestroy {
  telemetryService = inject(TelemetryService);
  private authService = inject(AuthService);

  currentPage = signal(1);
  pageSize = signal(50);
  totalPages = computed(() => Math.max(1, Math.ceil(this.telemetryService.totalLogs() / this.pageSize())));

  ngOnInit() {
    this.loadLogs();
  }

  ngOnDestroy() {
    if (this.telemetryService.liveMode()) {
      this.telemetryService.toggleLiveMode(''); // disconnect
    }
  }

  loadLogs() {
    this.telemetryService.getTelemetryLogs(this.currentPage(), this.pageSize()).subscribe();
  }

  changePage(delta: number) {
    const newPage = this.currentPage() + delta;
    if (newPage > 0 && newPage <= this.totalPages()) {
      this.currentPage.set(newPage);
      this.loadLogs();
    }
  }

  toggleLiveMode() {
    const token = this.authService.getToken();
    if (!token) return;
    this.telemetryService.toggleLiveMode(token);
    
    // If we turned it off, reload page 1 to sync back with historical
    if (!this.telemetryService.liveMode()) {
      this.currentPage.set(1);
      this.loadLogs();
    }
  }

  getStatusClass(status: string) {
    switch (status) {
      case 'Ok': return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400';
      case 'Error': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400';
      default: return 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300';
    }
  }
}
