import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuditService, AuditLog } from '../../services/audit.service';

@Component({
  selector: 'app-audit-logs',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './audit-logs.html',
})
export class AuditLogsComponent implements OnInit {
  private service = inject(AuditService);

  logs = signal<AuditLog[]>([]);
  stats = signal<any>(null);
  isLoading = signal(true);
  filterType = signal<string>('all');

  ngOnInit() {
    this.loadLogs();
  }

  get filteredLogs() {
    const f = this.filterType();
    if (f === 'all') return this.logs();
    if (f === 'critical') return this.logs().filter(l => l.severity === 'error' || l.severity === 'critical');
    if (f === 'high_risk') return this.logs().filter(l => {
      const a = (l.action || l.action_type || '').toLowerCase();
      return a.includes('delete') || a.includes('override') || a.includes('force');
    });
    return this.logs().filter(l => (l.action_type || '').includes(f));
  }

  setFilter(f: string) {
    this.filterType.set(f);
  }

  getSeverityClass(log: AuditLog): string {
    const action = (log.action_type || '').toLowerCase();
    if (action.includes('error') || action.includes('fail') || action.includes('denied')) return 'error';
    if (action.includes('delete') || action.includes('suspend') || action.includes('force')) return 'danger';
    if (action.includes('update') || action.includes('override')) return 'warning';
    if (action.includes('create')) return 'success';
    if (action.includes('login') || action.includes('auth')) return 'info';
    return 'neutral';
  }

  loadLogs() {
    this.isLoading.set(true);
    this.service.getAuditLogs().subscribe({
      next: (res) => {
        this.logs.set(res.data || []);
        this.stats.set(res.stats || null);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load audit logs', err);
        this.isLoading.set(false);
      }
    });
  }

  exportCSV() {
    const data = this.filteredLogs;
    const headers = ['ID', 'Actor', 'Action', 'Target Model', 'IP Address', 'Date'];
    const rows = data.map(l => [
      l.id, l.actor_id, l.action_type, l.target_model, l.ip_address, l.created_at
    ]);
    const csv = [headers, ...rows].map(r => r.join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `audit_logs_${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }
}
