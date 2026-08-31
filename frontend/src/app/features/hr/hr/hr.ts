import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HrService } from '../../../core/services/hr.service';
import { IntelligenceService } from '../../../core/services/intelligence.service';
import { Staff } from '../../../core/models/hr.models';
import { ToastrService } from 'ngx-toastr';
import { AlertService } from '../../../core/services/alert.service';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-hr',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './hr.html',
  styles: `
    .animated-gradient-text { color: #005b96; }
    .glass-panel {
      background: rgba(255, 255, 255, 0.7);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border: 1px solid rgba(255, 255, 255, 0.4);
    }
    :host-context(.dark) .glass-panel {
      background: rgba(24, 24, 27, 0.7);
      border: 1px solid rgba(255, 255, 255, 0.08);
    }
    .card-hover {
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }
    .card-hover:hover {
      transform: translateY(-4px) scale(1.01);
      box-shadow: 0 20px 40px -10px rgba(249, 115, 22, 0.15);
    }
    :host-context(.dark) .card-hover:hover {
      box-shadow: 0 20px 40px -10px rgba(0, 0, 0, 0.5);
    }
  `
})
export class Hr implements OnInit {
  createPayrollPeriod() {}
  private route = inject(ActivatedRoute);
  hrService = inject(HrService);
  intelligence = inject(IntelligenceService);
  private toastr = inject(ToastrService);
  private alertService = inject(AlertService);
  authService = inject(AuthService);

  canManagePayroll = this.authService.hasPermission(['hr:manage', 'payroll:manage']);

  activeTab = signal<'staff' | 'attendance' | 'leaves' | 'payroll' | 'roster'>('staff');

  isStaffModalOpen = signal(false);
  isLeaveModalOpen = signal(false);
  isShiftModalOpen = signal(false);
  
  shiftForm = signal({
    staff_id: '',
    start_time: '',
    end_time: '',
    role: '',
    notes: ''
  });
  isPayrollPeriodModalOpen = signal(false);

  searchQuery = signal('');
  currentStaff = signal<Partial<Staff>>({ is_active: true, role: 'cashier' });
  newLeave = signal<any>({ leave_type: 'annual', start_date: '', end_date: '', reason: '' });
  newPeriod = signal<any>({ name: '', start_date: '', end_date: '' });
  payrollPeriods = signal<any[]>([]);

  todayPresent = computed(() =>
    this.hrService.attendances().filter(a => {
      const today = new Date().toDateString();
      return new Date(a.clock_in).toDateString() === today && !a.clock_out;
    }).length
  );
  onLeave = computed(() =>
    this.hrService.leaveRequests().filter(l => {
      const now = new Date();
      return l.status === 'approved' && new Date(l.start_date) <= now && new Date(l.end_date) >= now;
    }).length
  );
  pendingLeaves = computed(() =>
    this.hrService.leaveRequests().filter(l => l.status === 'pending').length
  );

  // --- Filters ---
  attendanceFilters = signal({ date_from: '', date_to: '', staff_id: '' });

  ngOnInit() {
    this.route.paramMap.subscribe(params => {
      const tab = params.get('tab');
      if (tab && ['staff', 'attendance', 'leaves', 'payroll'].includes(tab)) {
        this.activeTab.set(tab as any);
        this.loadActiveTabData();
      }
    });

    this.hrService.getStaff().subscribe();
    this.hrService.listAttendance().subscribe();
    this.hrService.listLeaveRequests().subscribe();
    this.intelligence.getStaffLeaderboard(30).subscribe();
    if (this.canManagePayroll) {
      this.loadPayrollPeriods();
    }
  }

  loadActiveTabData() {
    switch (this.activeTab()) {
      case 'attendance': this.applyFilters(); break;
      case 'leaves': this.hrService.listLeaveRequests().subscribe(); break;
      case 'payroll': this.loadPayrollPeriods(); break;
      case 'roster': this.hrService.listShifts().subscribe(); break;
    }
  }

  applyFilters() {
    this.hrService.listAttendance(this.attendanceFilters()).subscribe();
  }

  clearFilters() {
    this.attendanceFilters.set({ date_from: '', date_to: '', staff_id: '' });
    this.applyFilters();
  }

  get totalFilteredHours() {
    return this.hrService.attendances().reduce((total, a) => {
      if (!a.clock_out) return total;
      const hours = (new Date(a.clock_out).getTime() - new Date(a.clock_in).getTime()) / 3600000;
      return total + Math.max(0, hours);
    }, 0);
  }

  get averageFilteredHours() {
    const attendances = this.hrService.attendances().filter(a => a.clock_out);
    if (!attendances.length) return 0;
    return this.totalFilteredHours / attendances.length;
  }

  get activeShifts() {
    return this.hrService.attendances().filter(a => !a.clock_out).length;
  }

  // ... (keep staff logic)
  loadPayrollPeriods() {
    if (!this.canManagePayroll) return;
    this.hrService.listPayrollPeriods().subscribe(p => this.payrollPeriods.set(p || []));
  }

  get filteredStaff() {
    const q = this.searchQuery().toLowerCase();
    return this.hrService.staff().filter(
      (s: any) =>
        (s.first_name || '').toLowerCase().includes(q) ||
        (s.last_name || '').toLowerCase().includes(q) ||
        (s.role || '').toLowerCase().includes(q)
    );
  }

  staffName(id: string): string {
    const s = this.hrService.staff().find((s: any) => s.id === id || s.user_id === id);
    if (!s) return id?.slice(0, 8) + '...';
    return `${s.first_name || ''} ${s.last_name || ''}`.trim() || s.role;
  }

  staffInitials(id: string): string {
    const s = this.hrService.staff().find((s: any) => s.id === id || s.user_id === id);
    if (!s) return '?';
    return `${s.first_name?.[0] || ''}${s.last_name?.[0] || ''}` || '?';
  }

  hoursWorked(clockIn: string, clockOut?: string): number {
    if (!clockOut) return 0;
    return (new Date(clockOut).getTime() - new Date(clockIn).getTime()) / 3600000;
  }

  hoursWorkedFormatted(clockIn: string, clockOut?: string): string {
    if (!clockOut) return 'Active';
    return `${this.hoursWorked(clockIn, clockOut).toFixed(1)}h`;
  }

  progressWidth(clockIn: string, clockOut?: string): string {
    const hours = this.hoursWorked(clockIn, clockOut);
    const target = 8;
    return `${Math.min((hours / target) * 100, 100)}%`;
  }

  openStaffModal(staff?: Staff) {
    if (staff) this.currentStaff.set({ ...staff });
    else this.currentStaff.set({ is_active: true, role: 'cashier' });
    this.isStaffModalOpen.set(true);
  }

  saveStaff() {
    const s = this.currentStaff();
    const action = s.id ? this.hrService.updateStaff(s.id, s) : this.hrService.createStaff(s);
    action.subscribe({
      next: () => {
        this.toastr.success(`Staff ${s.id ? 'updated' : 'created'}`);
        this.isStaffModalOpen.set(false);
        this.hrService.getStaff().subscribe();
      },
      error: () => this.toastr.error('Failed to save staff')
    });
  }

  clockIn() {
    this.hrService.clockIn().subscribe({
      next: () => { this.toastr.success('Clocked in!'); this.applyFilters(); },
      error: (e: any) => this.toastr.error(e?.error?.error || 'Clock in failed.')
    });
  }

  clockOut() {
    this.hrService.clockOut().subscribe({
      next: () => { this.toastr.success('Clocked out!'); this.applyFilters(); },
      error: (e: any) => this.toastr.error(e?.error?.error || 'Clock out failed.')
    });
  }

  manualCorrect(a: any) {
    const nowIso = new Date().toISOString().slice(0, 16);
    const corrected = prompt('Enter clock-out time (YYYY-MM-DDTHH:mm):', nowIso);
    if (!corrected) return;
    this.hrService.correctAttendance(a.id, corrected).subscribe({
      next: () => { this.toastr.success('Attendance corrected'); this.applyFilters(); },
      error: (e: any) => this.toastr.error(e?.error?.error || 'Correction failed')
    });
  }

  async deleteAttendance(a: any) {
    if (!(await this.alertService.confirm('Are you sure you want to delete this record?', 'Delete Record'))) return;
    this.hrService.deleteAttendance(a.id).subscribe({
      next: () => { this.toastr.success('Record deleted'); this.applyFilters(); },
      error: () => this.toastr.error('Failed to delete record')
    });
  }

  saveShift() {
    this.hrService.createShift(this.shiftForm()).subscribe({
      next: () => {
        this.toastr.success('Shift scheduled!');
        this.isShiftModalOpen.set(false);
        this.shiftForm.set({ staff_id: '', start_time: '', end_time: '', role: '', notes: '' });
      },
      error: () => this.toastr.error('Failed to schedule shift')
    });
  }

  exportTimesheets() {
    const attendances = this.hrService.attendances();
    if (!attendances.length) {
      this.toastr.warning('No records to export');
      return;
    }
    const headers = ['Staff Name', 'Clock In', 'Clock Out', 'Hours', 'Status'];
    const rows = attendances.map(a => [
      this.staffName(a.staff_id),
      new Date(a.clock_in).toLocaleString(),
      a.clock_out ? new Date(a.clock_out).toLocaleString() : '—',
      this.hoursWorkedFormatted(a.clock_in, a.clock_out),
      a.clock_out ? 'Completed' : 'Active'
    ]);
    const csvContent = "data:text/csv;charset=utf-8," 
      + [headers.join(','), ...rows.map(e => e.join(','))].join('\n');
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', `timesheets_${new Date().toISOString().split('T')[0]}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  submitLeave() {
    const l = this.newLeave();
    if (!l.start_date || !l.end_date) { this.toastr.warning('Please fill in all dates'); return; }
    this.hrService.createLeaveRequest(l).subscribe({
      next: () => {
        this.toastr.success('Leave request submitted');
        this.isLeaveModalOpen.set(false);
        this.newLeave.set({ leave_type: 'annual', start_date: '', end_date: '', reason: '' });
        this.hrService.listLeaveRequests().subscribe();
      },
      error: () => this.toastr.error('Submission failed')
    });
  }

  approveLeave(id: string) {
    this.hrService.approveLeaveRequest(id).subscribe({
      next: () => { this.toastr.success('Leave approved'); this.hrService.listLeaveRequests().subscribe(); },
      error: () => this.toastr.error('Failed to approve leave')
    });
  }

  rejectLeave(id: string) {
    this.hrService.rejectLeaveRequest(id).subscribe({
      next: () => { this.toastr.success('Leave rejected'); this.hrService.listLeaveRequests().subscribe(); },
      error: () => this.toastr.error('Failed to reject leave')
    });
  }

  processPayroll(periodId: string) {
    this.hrService.processPayroll(periodId).subscribe({
      next: () => { this.toastr.success('Payroll processed!'); this.loadPayrollPeriods(); },
      error: () => this.toastr.error('Payroll processing failed')
    });
  }
}
