import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, map, tap } from 'rxjs';
import { Staff, Attendance, LeaveRequest } from '../models/hr.models';

export interface LeaveCreateInput {
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
  leaveRequests = signal<LeaveRequest[]>([]);
  loading = signal<boolean>(false);
  staff = signal<Staff[]>([]);
  getStaff(): Observable<Staff[]> { return this.api.get<Staff[]>('/staff').pipe(tap(res => this.staff.set(res || []))); }
  createStaff(s: any): Observable<any> { return this.api.post('/staff', s); }
  updateStaff(id: string, s: any): Observable<any> { return this.api.put('/staff/'+id, s); }

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
      tap(a => this.attendances.update(list => [a, ...list]))
    );
  }

  clockOut(): Observable<Attendance> {
    return this.api.post<Attendance>('/hr/attendance/clock_out', {}).pipe(
      tap(a => this.attendances.update(list => list.map(item => item.staff_id === a.staff_id ? a : item)))
    );
  }

  correctAttendance(id: string, clockOut: string): Observable<Attendance> {
    return this.api.patch<Attendance>(`/hr/attendance/${id}/correct`, { clock_out: clockOut }).pipe(
      tap(a => this.attendances.update(list => list.map(item => item.id === a.id ? a : item)))
    );
  }

  deleteAttendance(id: string): Observable<void> {
    return this.api.delete<void>(`/hr/attendance/${id}`).pipe(
      tap(() => this.attendances.update(list => list.filter(item => item.id !== id)))
    );
  }


  listLeaveRequests(): Observable<LeaveRequest[]> {
    this.loading.set(true);
    return this.api.get<LeaveRequest[]>('/hr/leave-requests').pipe(
      tap(res => {
        this.leaveRequests.set(res || []);
        this.loading.set(false);
      })
    );
  }

  createLeaveRequest(input: LeaveCreateInput): Observable<LeaveRequest> {
    return this.api.post<LeaveRequest>('/hr/leave-requests', input).pipe(
      tap(l => this.leaveRequests.update(list => [l, ...list]))
    );
  }

  approveLeaveRequest(id: string): Observable<any> { return this.api.put(`/hr/leave-requests/${id}/approve`, {}); }
  rejectLeaveRequest(id: string): Observable<any> { return this.api.put(`/hr/leave-requests/${id}/reject`, {}); }

  listPayrollPeriods(): Observable<any[]> {
    return this.api.get<{ periods: any[] }>('/hr/payroll/periods').pipe(
      map(res => res?.periods || [])
    );
  }
  getPayrollPeriod(id: string): Observable<any> { return this.api.get<any>(`/hr/payroll/periods/${id}`); }
  processPayroll(id: string): Observable<any> { return this.api.post<any>(`/hr/payroll/periods/${id}/process`, {}); }
  getPayslip(id: string): Observable<any> { return this.api.get<any>(`/hr/payslips/${id}`); }

  // --- Commission Rules ---
  commissionRules = signal<any[]>([]);
  listCommissionRules(): Observable<any[]> {
    return this.api.get<any[]>('/hr/commission-rules').pipe(
      tap(res => this.commissionRules.set(res || []))
    );
  }
  createCommissionRule(rule: any): Observable<any> {
    return this.api.post<any>('/hr/commission-rules', rule).pipe(
      tap(r => this.commissionRules.update(list => [r, ...list]))
    );
  }

  // --- Staff Achievements ---
  achievements = signal<any[]>([]);
  listAchievements(): Observable<any[]> {
    return this.api.get<any[]>('/hr/achievements').pipe(
      tap(res => this.achievements.set(res || []))
    );
  }
  createAchievement(a: any): Observable<any> {
    return this.api.post<any>('/hr/achievements', a).pipe(
      tap(r => this.achievements.update(list => [r, ...list]))
    );
  }

  // --- Shift Swap Requests ---
  shiftSwaps = signal<any[]>([]);
  listShiftSwaps(): Observable<any[]> {
    return this.api.get<any[]>('/hr/shift-swaps').pipe(
      tap(res => this.shiftSwaps.set(res || []))
    );
  }
  createShiftSwap(swap: any): Observable<any> {
    return this.api.post<any>('/hr/shift-swaps', swap).pipe(
      tap(r => this.shiftSwaps.update(list => [r, ...list]))
    );
  }
  // --- Shift Roster ---
  shifts = signal<any[]>([]);
  listShifts(): Observable<any[]> {
    return this.api.get<any[]>('/hr/roster').pipe(
      tap(res => this.shifts.set(res || []))
    );
  }
  createShift(shift: any): Observable<any> {
    return this.api.post<any>('/hr/roster', shift).pipe(
      tap(r => this.shifts.update(list => [...list, r]))
    );
  }
}
