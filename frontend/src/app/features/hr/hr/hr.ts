import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HrService } from '../../../core/services/hr.service';
import { IntelligenceService } from '../../../core/services/intelligence.service';
import { Staff, PayrollPeriod, PayrollRecord, CommissionRule, StaffAchievement, ShiftSwapRequest, StaffShift } from '../../../core/models/hr.models';
import { ToastrService } from 'ngx-toastr';
import { AlertService } from '../../../core/services/alert.service';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-hr',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './hr.html',
  styles: `
    .animated-gradient-text { 
      background: linear-gradient(135deg, #4f46e5 0%, #06b6d4 100%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }
    .glass-panel {
      background: rgba(255, 255, 255, 0.75);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border: 1px solid rgba(255, 255, 255, 0.5);
    }
    :host-context(.dark) .glass-panel {
      background: rgba(24, 24, 27, 0.75);
      border: 1px solid rgba(255, 255, 255, 0.08);
    }
    .card-hover {
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }
    .card-hover:hover {
      transform: translateY(-3px);
      box-shadow: 0 20px 30px -10px rgba(79, 70, 229, 0.12);
    }
    :host-context(.dark) .card-hover:hover {
      box-shadow: 0 20px 30px -10px rgba(0, 0, 0, 0.6);
    }
  `
})
export class Hr implements OnInit {
  private route = inject(ActivatedRoute);
  hrService = inject(HrService);
  intelligence = inject(IntelligenceService);
  private toastr = inject(ToastrService);
  private alertService = inject(AlertService);
  authService = inject(AuthService);

  canManagePayroll = this.authService.hasPermission(['hr:manage', 'payroll:manage']);

  activeTab = signal<'staff' | 'attendance' | 'leaves' | 'payroll' | 'roster' | 'shift_swaps' | 'commissions' | 'achievements'>('staff');

  // Modals state
  isStaffModalOpen = signal(false);
  isLeaveModalOpen = signal(false);
  isShiftModalOpen = signal(false);
  isPayrollPeriodModalOpen = signal(false);
  isCommissionRuleModalOpen = signal(false);
  isAchievementModalOpen = signal(false);
  isShiftSwapModalOpen = signal(false);
  isPayslipModalOpen = signal(false);
  isAttendanceCorrectionModalOpen = signal(false);
  isLeaveActionModalOpen = signal(false);

  // Active records & forms
  searchQuery = signal('');
  currentStaff = signal<Partial<Staff>>({ is_active: true, role: 'cashier', payment_method: 'bank_transfer', employment_type: 'full_time' });
  newLeave = signal<any>({ leave_type: 'annual', start_date: '', end_date: '', reason: '' });
  newPeriod = signal<any>({ name: '', start_date: '', end_date: '' });
  newCommissionRule = signal<CommissionRule>({ name: '', rule_type: 'percentage', min_sales_amount: 0, commission_percentage: 5, flat_bonus: 0, is_active: true });
  newAchievement = signal<StaffAchievement>({ staff_id: '', badge_name: 'Top Performer', badge_icon: 'emoji_events', description: '' });
  newShiftSwap = signal<any>({ requesting_staff_id: '', target_staff_id: '', original_shift_id: '', notes: '' });

  shiftForm = signal({
    staff_id: '',
    start_time: '',
    end_time: '',
    role: '',
    notes: ''
  });

  // Selected for modals
  selectedPeriod = signal<PayrollPeriod | null>(null);
  selectedPayslip = signal<PayrollRecord | null>(null);
  selectedAttendance = signal<any | null>(null);
  correctionClockOut = signal('');
  correctionNotes = signal('');

  selectedLeaveForAction = signal<any | null>(null);
  leaveActionType = signal<'approve' | 'reject'>('approve');
  managerLeaveNote = signal('');

  payPaymentMethod = signal('bank_transfer');
  payPaymentReference = signal('');

  // Overview metrics computed from signals
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

  pendingShiftSwaps = computed(() =>
    this.hrService.shiftSwaps().filter(s => s.status === 'pending').length
  );

  // Filters
  attendanceFilters = signal({ date_from: '', date_to: '', staff_id: '' });
  staffRoleFilter = signal<string>('all');

  ngOnInit() {
    this.route.paramMap.subscribe(params => {
      const tab = params.get('tab');
      if (tab && ['staff', 'attendance', 'leaves', 'payroll', 'roster', 'shift_swaps', 'commissions', 'achievements'].includes(tab)) {
        this.activeTab.set(tab as any);
      }
      this.loadActiveTabData();
    });

    this.hrService.getStaff().subscribe();
    this.hrService.listAttendance().subscribe();
    this.hrService.listLeaveRequests().subscribe();
    this.hrService.getLeaveSummary().subscribe();
    this.intelligence.getStaffLeaderboard(30).subscribe();
    if (this.canManagePayroll) {
      this.loadPayrollPeriods();
      this.hrService.listCommissionRules().subscribe();
    }
    this.hrService.listAchievements().subscribe();
    this.hrService.listShiftSwaps().subscribe();
    this.hrService.listShifts().subscribe();
  }

  loadActiveTabData() {
    switch (this.activeTab()) {
      case 'attendance': this.applyFilters(); break;
      case 'leaves': 
        this.hrService.listLeaveRequests().subscribe(); 
        this.hrService.getLeaveSummary().subscribe();
        break;
      case 'payroll': this.loadPayrollPeriods(); break;
      case 'roster': this.hrService.listShifts().subscribe(); break;
      case 'shift_swaps': this.hrService.listShiftSwaps().subscribe(); break;
      case 'commissions': this.hrService.listCommissionRules().subscribe(); break;
      case 'achievements': this.hrService.listAchievements().subscribe(); break;
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
      if (a.hours_worked) return total + a.hours_worked;
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

  loadPayrollPeriods() {
    if (!this.canManagePayroll) return;
    this.hrService.listPayrollPeriods().subscribe();
  }

  get filteredStaff() {
    const q = this.searchQuery().toLowerCase();
    const role = this.staffRoleFilter();
    return this.hrService.staff().filter(
      (s: any) => {
        const matchesQuery = (s.first_name || '').toLowerCase().includes(q) ||
          (s.last_name || '').toLowerCase().includes(q) ||
          (s.role || '').toLowerCase().includes(q) ||
          (s.department || '').toLowerCase().includes(q) ||
          (s.email || '').toLowerCase().includes(q);
        const matchesRole = role === 'all' || s.role === role;
        return matchesQuery && matchesRole;
      }
    );
  }

  staffName(id: string): string {
    const s = this.hrService.staff().find((s: any) => s.id === id || s.user_id === id);
    if (!s) return id?.slice(0, 8) + '...';
    return `${s.first_name || ''} ${s.last_name || ''}`.trim() || s.role || 'Staff Member';
  }

  staffInitials(id: string): string {
    const s = this.hrService.staff().find((s: any) => s.id === id || s.user_id === id);
    if (!s) return '?';
    return `${s.first_name?.[0] || ''}${s.last_name?.[0] || ''}`.toUpperCase() || '?';
  }

  hoursWorked(clockIn: string, clockOut?: string): number {
    if (!clockOut) return 0;
    return (new Date(clockOut).getTime() - new Date(clockIn).getTime()) / 3600000;
  }

  hoursWorkedFormatted(clockIn: string, clockOut?: string, hoursWorkedDirect?: number): string {
    if (hoursWorkedDirect !== undefined && hoursWorkedDirect > 0) return `${hoursWorkedDirect.toFixed(1)}h`;
    if (!clockOut) return 'Active';
    return `${this.hoursWorked(clockIn, clockOut).toFixed(1)}h`;
  }

  progressWidth(clockIn: string, clockOut?: string, hoursWorkedDirect?: number): string {
    const hours = hoursWorkedDirect !== undefined && hoursWorkedDirect > 0 ? hoursWorkedDirect : this.hoursWorked(clockIn, clockOut);
    const target = 8;
    return `${Math.min((hours / target) * 100, 100)}%`;
  }

  // --- Staff Actions ---
  openStaffModal(staff?: Staff) {
    if (staff) {
      this.currentStaff.set({ ...staff });
    } else {
      this.currentStaff.set({ 
        is_active: true, 
        role: 'cashier', 
        payment_method: 'bank_transfer',
        employment_type: 'full_time',
        base_salary: 0,
        hourly_rate: 0
      });
    }
    this.isStaffModalOpen.set(true);
  }

  saveStaff() {
    const s = this.currentStaff();
    const action = s.id ? this.hrService.updateStaff(s.id, s) : this.hrService.createStaff(s);
    action.subscribe({
      next: () => {
        this.toastr.success(`Staff ${s.id ? 'updated' : 'created'} successfully`);
        this.isStaffModalOpen.set(false);
        this.hrService.getStaff().subscribe();
      },
      error: (err: any) => this.toastr.error(err?.error?.error || 'Failed to save staff profile')
    });
  }

  // --- Time & Attendance Actions ---
  clockIn() {
    this.hrService.clockIn().subscribe({
      next: () => { this.toastr.success('Clocked in successfully!'); this.applyFilters(); },
      error: (e: any) => this.toastr.error(e?.error?.error || 'Clock in failed.')
    });
  }

  clockOut() {
    this.hrService.clockOut().subscribe({
      next: () => { this.toastr.success('Clocked out successfully!'); this.applyFilters(); },
      error: (e: any) => this.toastr.error(e?.error?.error || 'Clock out failed.')
    });
  }

  openAttendanceCorrectionModal(a: any) {
    this.selectedAttendance.set(a);
    const nowIso = new Date().toISOString().slice(0, 16);
    this.correctionClockOut.set(nowIso);
    this.correctionNotes.set(a.notes || '');
    this.isAttendanceCorrectionModalOpen.set(true);
  }

  submitAttendanceCorrection() {
    const a = this.selectedAttendance();
    if (!a || !this.correctionClockOut()) return;

    this.hrService.correctAttendance(a.id, this.correctionClockOut(), this.correctionNotes()).subscribe({
      next: () => {
        this.toastr.success('Attendance record corrected');
        this.isAttendanceCorrectionModalOpen.set(false);
        this.applyFilters();
      },
      error: (e: any) => this.toastr.error(e?.error?.error || 'Correction failed')
    });
  }

  async deleteAttendance(a: any) {
    if (!(await this.alertService.confirm('Are you sure you want to delete this attendance record?', 'Delete Record'))) return;
    this.hrService.deleteAttendance(a.id).subscribe({
      next: () => { this.toastr.success('Record deleted'); this.applyFilters(); },
      error: () => this.toastr.error('Failed to delete record')
    });
  }

  exportTimesheets() {
    const attendances = this.hrService.attendances();
    if (!attendances.length) {
      this.toastr.warning('No records to export');
      return;
    }
    const headers = ['Staff Name', 'Clock In', 'Clock Out', 'Hours Worked', 'Status', 'Notes'];
    const rows = attendances.map(a => [
      `"${this.staffName(a.staff_id)}"`,
      `"${new Date(a.clock_in).toLocaleString()}"`,
      a.clock_out ? `"${new Date(a.clock_out).toLocaleString()}"` : '"—"',
      `"${this.hoursWorkedFormatted(a.clock_in, a.clock_out, a.hours_worked)}"`,
      `"${a.clock_out ? 'Completed' : 'Active'}"`,
      `"${a.notes || ''}"`
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

  // --- Leaves Actions ---
  submitLeave() {
    const l = this.newLeave();
    if (!l.start_date || !l.end_date) { this.toastr.warning('Please select start and end dates'); return; }
    this.hrService.createLeaveRequest(l).subscribe({
      next: () => {
        this.toastr.success('Leave request submitted successfully');
        this.isLeaveModalOpen.set(false);
        this.newLeave.set({ leave_type: 'annual', start_date: '', end_date: '', reason: '' });
        this.hrService.listLeaveRequests().subscribe();
        this.hrService.getLeaveSummary().subscribe();
      },
      error: (err: any) => this.toastr.error(err?.error?.error || 'Submission failed')
    });
  }

  openLeaveActionModal(leave: any, type: 'approve' | 'reject') {
    this.selectedLeaveForAction.set(leave);
    this.leaveActionType.set(type);
    this.managerLeaveNote.set('');
    this.isLeaveActionModalOpen.set(true);
  }

  confirmLeaveAction() {
    const leave = this.selectedLeaveForAction();
    if (!leave) return;

    const action = this.leaveActionType() === 'approve'
      ? this.hrService.approveLeaveRequest(leave.id, this.managerLeaveNote())
      : this.hrService.rejectLeaveRequest(leave.id, this.managerLeaveNote());

    action.subscribe({
      next: () => {
        this.toastr.success(`Leave request ${this.leaveActionType()}d`);
        this.isLeaveActionModalOpen.set(false);
        this.hrService.listLeaveRequests().subscribe();
        this.hrService.getLeaveSummary().subscribe();
      },
      error: (err: any) => this.toastr.error(err?.error?.error || `Failed to ${this.leaveActionType()} leave`)
    });
  }

  // --- Payroll Engine Actions ---
  createPayrollPeriod() {
    const p = this.newPeriod();
    if (!p.name || !p.start_date || !p.end_date) {
      this.toastr.warning('Please fill in all period details');
      return;
    }
    this.hrService.createPayrollPeriod(p).subscribe({
      next: () => {
        this.toastr.success('Payroll cycle created!');
        this.isPayrollPeriodModalOpen.set(false);
        this.newPeriod.set({ name: '', start_date: '', end_date: '' });
        this.loadPayrollPeriods();
      },
      error: (err: any) => this.toastr.error(err?.error?.error || 'Failed to create period')
    });
  }

  processPayroll(periodId: string) {
    this.hrService.processPayroll(periodId).subscribe({
      next: (res) => {
        this.toastr.success('Payroll calculated and finalized!');
        this.loadPayrollPeriods();
        if (this.selectedPeriod()?.id === periodId) {
          this.viewPeriodBreakdown(res.period);
        }
      },
      error: (err: any) => this.toastr.error(err?.error?.error || 'Payroll calculation failed')
    });
  }

  viewPeriodBreakdown(period: PayrollPeriod) {
    this.hrService.getPayrollPeriod(period.id).subscribe({
      next: (fullPeriod) => {
        this.selectedPeriod.set(fullPeriod);
      },
      error: () => this.selectedPeriod.set(period)
    });
  }

  openPayslipModal(record: PayrollRecord) {
    this.selectedPayslip.set(record);
    this.payPaymentMethod.set(record.payment_method || 'bank_transfer');
    this.payPaymentReference.set(record.payment_reference || '');
    this.isPayslipModalOpen.set(true);
  }

  markPayslipPaid() {
    const payslip = this.selectedPayslip();
    if (!payslip) return;

    this.hrService.markPayslipPaid(payslip.id, this.payPaymentMethod(), this.payPaymentReference()).subscribe({
      next: (updated) => {
        this.toastr.success('Payslip marked as paid!');
        this.selectedPayslip.set(updated);
        // Refresh period
        if (this.selectedPeriod()) {
          this.viewPeriodBreakdown(this.selectedPeriod()!);
        }
      },
      error: (err: any) => this.toastr.error(err?.error?.error || 'Failed to update payment status')
    });
  }

  printPayslip() {
    window.print();
  }

  // --- Commission Rules Actions ---
  openCommissionRuleModal() {
    this.newCommissionRule.set({
      name: '',
      rule_type: 'percentage',
      min_sales_amount: 500,
      commission_percentage: 5,
      flat_bonus: 0,
      is_active: true
    });
    this.isCommissionRuleModalOpen.set(true);
  }

  saveCommissionRule() {
    const r = this.newCommissionRule();
    if (!r.name) {
      this.toastr.warning('Please enter a rule name');
      return;
    }
    this.hrService.createCommissionRule(r).subscribe({
      next: () => {
        this.toastr.success('Commission rule created!');
        this.isCommissionRuleModalOpen.set(false);
        this.hrService.listCommissionRules().subscribe();
      },
      error: (err: any) => this.toastr.error(err?.error?.error || 'Failed to create commission rule')
    });
  }

  deleteCommissionRule(id: string) {
    this.hrService.deleteCommissionRule(id).subscribe({
      next: () => {
        this.toastr.success('Commission rule removed');
        this.hrService.listCommissionRules().subscribe();
      },
      error: () => this.toastr.error('Failed to remove commission rule')
    });
  }

  // --- Gamification & Badges ---
  openAchievementModal() {
    this.newAchievement.set({
      staff_id: this.hrService.staff()[0]?.id || '',
      badge_name: 'Sales Star',
      badge_icon: 'stars',
      description: 'Outstanding performance and sales dedication'
    });
    this.isAchievementModalOpen.set(true);
  }

  saveAchievement() {
    const a = this.newAchievement();
    if (!a.staff_id || !a.badge_name) {
      this.toastr.warning('Please select staff and enter badge name');
      return;
    }
    this.hrService.createAchievement(a).subscribe({
      next: () => {
        this.toastr.success('Achievement badge awarded!');
        this.isAchievementModalOpen.set(false);
        this.hrService.listAchievements().subscribe();
      },
      error: () => this.toastr.error('Failed to award badge')
    });
  }

  deleteAchievement(id: string) {
    this.hrService.deleteAchievement(id).subscribe({
      next: () => {
        this.toastr.success('Badge deleted');
        this.hrService.listAchievements().subscribe();
      },
      error: () => this.toastr.error('Failed to delete badge')
    });
  }

  // --- Shift Roster Actions ---
  saveShift() {
    this.hrService.createShift(this.shiftForm()).subscribe({
      next: () => {
        this.toastr.success('Shift scheduled!');
        this.isShiftModalOpen.set(false);
        this.shiftForm.set({ staff_id: '', start_time: '', end_time: '', role: '', notes: '' });
        this.hrService.listShifts().subscribe();
      },
      error: () => this.toastr.error('Failed to schedule shift')
    });
  }

  // --- Shift Swapping ---
  openShiftSwapModal() {
    this.newShiftSwap.set({
      requesting_staff_id: this.hrService.staff()[0]?.id || '',
      target_staff_id: '',
      original_shift_id: this.hrService.shifts()[0]?.id || '',
      notes: ''
    });
    this.isShiftSwapModalOpen.set(true);
  }

  submitShiftSwap() {
    const swap = this.newShiftSwap();
    if (!swap.requesting_staff_id || !swap.original_shift_id) {
      this.toastr.warning('Please specify shift details');
      return;
    }
    this.hrService.createShiftSwap(swap).subscribe({
      next: () => {
        this.toastr.success('Shift swap request posted!');
        this.isShiftSwapModalOpen.set(false);
        this.hrService.listShiftSwaps().subscribe();
      },
      error: () => this.toastr.error('Failed to submit swap request')
    });
  }

  approveShiftSwap(id: string) {
    this.hrService.approveShiftSwap(id).subscribe({
      next: () => {
        this.toastr.success('Shift swap approved!');
        this.hrService.listShiftSwaps().subscribe();
      },
      error: () => this.toastr.error('Failed to approve swap')
    });
  }

  rejectShiftSwap(id: string) {
    this.hrService.rejectShiftSwap(id).subscribe({
      next: () => {
        this.toastr.success('Shift swap rejected');
        this.hrService.listShiftSwaps().subscribe();
      },
      error: () => this.toastr.error('Failed to reject swap')
    });
  }
}
