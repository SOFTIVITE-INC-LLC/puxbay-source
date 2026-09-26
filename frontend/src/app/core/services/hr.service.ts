import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, map, tap } from 'rxjs';
import { Staff, Attendance, LeaveRequest, PayrollPeriod, PayrollRecord, CommissionRule, StaffAchievement, ShiftSwapRequest, StaffShift, LeaveSummary } from '../models/hr.models';

export interface LeaveCreateInput {
  staff_id?: string;
  leave_type: string;
  start_date: string;
  end_date: string;
  reason?: string;
}

@Injectable({
  providedIn: 'root'
})
export class HrService {
  private api = inject(ApiService);
  
  attendances = signal<Attendance[]>([]);
  myAttendance = signal<{ is_clocked_in: boolean; attendance?: Attendance; today_hours: number }>({ is_clocked_in: false, today_hours: 0 });
  leaveRequests = signal<LeaveRequest[]>([]);
  leaveSummary = signal<LeaveSummary>({ total_quota: 21, used_days: 0, pending_days: 0, remaining_days: 21 });
  loading = signal<boolean>(false);
  staff = signal<Staff[]>([]);
  payrollPeriods = signal<PayrollPeriod[]>([]);
  commissionRules = signal<CommissionRule[]>([]);
  achievements = signal<StaffAchievement[]>([]);
  shiftSwaps = signal<ShiftSwapRequest[]>([]);
  shifts = signal<StaffShift[]>([]);

  // --- Staff Directory ---
  getStaff(branchId?: string): Observable<Staff[]> { 
    return this.api.get<Staff[]>('/staff', { params: branchId ? { branch_id: branchId } : undefined }).pipe(
      tap(res => this.staff.set(res || []))
    ); 
  }
  createStaff(s: any): Observable<any> { return this.api.post('/staff', s); }
  updateStaff(id: string, s: any): Observable<any> { return this.api.put('/staff/' + id, s); }
  deleteStaff(id: string): Observable<any> { return this.api.delete('/staff/' + id); }

  // --- Attendance & Timesheets ---
  getMyAttendanceStatus(): Observable<{ is_clocked_in: boolean; attendance?: Attendance; today_hours: number }> {
    return this.api.get<{ is_clocked_in: boolean; attendance?: Attendance; today_hours: number }>('/hr/attendance/my-status').pipe(
      tap(res => {
        if (res) {
          this.myAttendance.set(res);
        }
      })
    );
  }

  listAttendance(params?: { date_from?: string; date_to?: string; staff_id?: string }): Observable<Attendance[]> {
    this.loading.set(true);
    return this.api.get<Attendance[]>('/hr/attendance', { params }).pipe(
      tap(res => {
        this.attendances.set(res || []);
        this.loading.set(false);
      })
    );
  }

  clockIn(): Observable<Attendance> {
    return this.api.post<Attendance>('/hr/attendance/clock_in', {}).pipe(
      tap(a => {
        this.attendances.update(list => [a, ...list]);
        this.myAttendance.update(prev => ({ ...prev, is_clocked_in: true, attendance: a }));
      })
    );
  }

  clockOut(): Observable<Attendance> {
    return this.api.post<Attendance>('/hr/attendance/clock_out', {}).pipe(
      tap(a => {
        this.attendances.update(list => list.map(item => item.staff_id === a.staff_id && !item.clock_out ? a : item));
        this.myAttendance.update(prev => ({ ...prev, is_clocked_in: false, attendance: undefined }));
      })
    );
  }

  correctAttendance(id: string, clockOut: string, notes?: string): Observable<Attendance> {
    return this.api.patch<Attendance>(`/hr/attendance/${id}/correct`, { clock_out: clockOut, notes }).pipe(
      tap(a => this.attendances.update(list => list.map(item => item.id === a.id ? a : item)))
    );
  }

  deleteAttendance(id: string): Observable<any> {
    return this.api.delete<any>(`/hr/attendance/${id}`).pipe(
      tap(() => this.attendances.update(list => list.filter(item => item.id !== id)))
    );
  }

  // --- Leaves & Time Off ---
  listLeaveRequests(): Observable<LeaveRequest[]> {
    this.loading.set(true);
    return this.api.get<LeaveRequest[]>('/hr/leave-requests').pipe(
      tap(res => {
        this.leaveRequests.set(res || []);
        this.loading.set(false);
      })
    );
  }

  getLeaveSummary(staffId?: string): Observable<LeaveSummary> {
    return this.api.get<LeaveSummary>('/hr/leaves/summary', { params: staffId ? { staff_id: staffId } : undefined }).pipe(
      tap(res => {
        if (res) this.leaveSummary.set(res);
      })
    );
  }

  createLeaveRequest(input: LeaveCreateInput): Observable<LeaveRequest> {
    return this.api.post<LeaveRequest>('/hr/leave-requests', input).pipe(
      tap(l => this.leaveRequests.update(list => [l, ...list]))
    );
  }

  approveLeaveRequest(id: string, managerNotes?: string): Observable<any> { 
    return this.api.put(`/hr/leave-requests/${id}/approve`, { manager_notes: managerNotes }); 
  }

  rejectLeaveRequest(id: string, managerNotes?: string): Observable<any> { 
    return this.api.put(`/hr/leave-requests/${id}/reject`, { manager_notes: managerNotes }); 
  }

  // --- Payroll Cycles & Payslips ---
  listPayrollPeriods(): Observable<PayrollPeriod[]> {
    return this.api.get<{ periods: PayrollPeriod[] }>('/hr/payroll/periods').pipe(
      map(res => res?.periods || []),
      tap(periods => this.payrollPeriods.set(periods))
    );
  }

  createPayrollPeriod(period: { name: string; start_date: string; end_date: string }): Observable<PayrollPeriod> {
    return this.api.post<PayrollPeriod>('/hr/payroll/periods', period).pipe(
      tap(p => this.payrollPeriods.update(list => [p, ...list]))
    );
  }

  getPayrollPeriod(id: string): Observable<PayrollPeriod> { 
    return this.api.get<PayrollPeriod>(`/hr/payroll/periods/${id}`); 
  }

  processPayroll(id: string): Observable<{ status: string; period: PayrollPeriod }> { 
    return this.api.post<{ status: string; period: PayrollPeriod }>(`/hr/payroll/periods/${id}/process`, {}); 
  }

  getPayslip(id: string): Observable<PayrollRecord> { 
    return this.api.get<PayrollRecord>(`/hr/payslips/${id}`); 
  }

  markPayslipPaid(id: string, paymentMethod: string, paymentReference: string): Observable<PayrollRecord> {
    return this.api.post<PayrollRecord>(`/hr/payslips/${id}/pay`, { payment_method: paymentMethod, payment_reference: paymentReference });
  }

  // --- Commission Rules ---
  listCommissionRules(): Observable<CommissionRule[]> {
    return this.api.get<CommissionRule[]>('/hr/commission-rules').pipe(
      tap(res => this.commissionRules.set(res || []))
    );
  }

  createCommissionRule(rule: CommissionRule): Observable<CommissionRule> {
    return this.api.post<CommissionRule>('/hr/commission-rules', rule).pipe(
      tap(r => this.commissionRules.update(list => [r, ...list]))
    );
  }

  deleteCommissionRule(id: string): Observable<any> {
    return this.api.delete<any>(`/hr/commission-rules/${id}`).pipe(
      tap(() => this.commissionRules.update(list => list.filter(r => r.id !== id)))
    );
  }

  // --- Staff Achievements & Gamification ---
  listAchievements(): Observable<StaffAchievement[]> {
    return this.api.get<StaffAchievement[]>('/hr/achievements').pipe(
      tap(res => this.achievements.set(res || []))
    );
  }

  createAchievement(a: StaffAchievement): Observable<StaffAchievement> {
    return this.api.post<StaffAchievement>('/hr/achievements', a).pipe(
      tap(r => this.achievements.update(list => [r, ...list]))
    );
  }

  deleteAchievement(id: string): Observable<any> {
    return this.api.delete<any>(`/hr/achievements/${id}`).pipe(
      tap(() => this.achievements.update(list => list.filter(a => a.id !== id)))
    );
  }

  // --- Shift Swapping ---
  listShiftSwaps(): Observable<ShiftSwapRequest[]> {
    return this.api.get<ShiftSwapRequest[]>('/hr/shift-swaps').pipe(
      tap(res => this.shiftSwaps.set(res || []))
    );
  }

  createShiftSwap(swap: ShiftSwapRequest): Observable<ShiftSwapRequest> {
    return this.api.post<ShiftSwapRequest>('/hr/shift-swaps', swap).pipe(
      tap(r => this.shiftSwaps.update(list => [r, ...list]))
    );
  }

  approveShiftSwap(id: string): Observable<any> {
    return this.api.put(`/hr/shift-swaps/${id}/approve`, {});
  }

  rejectShiftSwap(id: string): Observable<any> {
    return this.api.put(`/hr/shift-swaps/${id}/reject`, {});
  }

  // --- Shift Roster ---
  listShifts(): Observable<StaffShift[]> {
    return this.api.get<StaffShift[]>('/hr/roster').pipe(
      tap(res => this.shifts.set(res || []))
    );
  }

  createShift(shift: StaffShift): Observable<StaffShift> {
    return this.api.post<StaffShift>('/hr/roster', shift).pipe(
      tap(r => this.shifts.update(list => [...list, r]))
    );
  }
}

