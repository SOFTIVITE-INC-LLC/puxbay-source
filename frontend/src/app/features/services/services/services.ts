import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ServicesService } from '../../../core/services/services.service';
import { StaffService } from '../../../core/services/staff.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { Service, Appointment } from '../../../core/models/services.models';
import { UserProfile } from '../../../core/models/user.models';
import { forkJoin } from 'rxjs';
import { ToastrService } from 'ngx-toastr';

type CalendarView = 'month' | 'week' | 'day';
type ModalMode = 'booking' | 'service';

interface CalendarDay {
  date: Date;
  isCurrentMonth: boolean;
  isToday: boolean;
  appointments: Appointment[];
}

@Component({
  selector: 'app-services',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './services.html',
})
export class Services implements OnInit {
  private servicesService = inject(ServicesService);
  private staffService = inject(StaffService);
  private toastr = inject(ToastrService);

  // View state
  view = signal<CalendarView>('month');
  currentDate = signal<Date>(new Date());
  loading = signal(false);
  activeTab = signal<'calendar' | 'services' | 'commissions'>('calendar');

  // Data
  services = signal<Service[]>([]);
  appointments = signal<Appointment[]>([]);
  staff = signal<UserProfile[]>([]);

  // Modals
  showBookingModal = signal(false);
  showServiceModal = signal(false);
  savingBooking = signal(false);
  savingService = signal(false);

  newBooking = signal<any>({
    service_id: '',
    staff_member_id: '',
    start_time: '',
    customer_name: '',
    customer_phone: '',
    customer_email: '',
    notes: '',
  });

  newService = signal<any>({
    name: '',
    description: '',
    price: 0,
    duration_minutes: 30,
    is_active: true,
  });

  // Computed calendar properties
  currentMonthLabel = computed(() => {
    return this.currentDate().toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
  });

  calendarDays = computed((): CalendarDay[] => {
    const date = this.currentDate();
    const year = date.getFullYear();
    const month = date.getMonth();

    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const days: CalendarDay[] = [];

    // Fill leading days from previous month
    const startPad = firstDay.getDay();
    for (let i = startPad - 1; i >= 0; i--) {
      const d = new Date(year, month, -i);
      days.push({ date: d, isCurrentMonth: false, isToday: false, appointments: [] });
    }

    // Current month days
    for (let d = 1; d <= lastDay.getDate(); d++) {
      const dayDate = new Date(year, month, d);
      const isToday = dayDate.getTime() === today.getTime();
      const appts = this.appointments().filter(a => {
        const aDate = new Date(a.start_time);
        return aDate.getFullYear() === year && aDate.getMonth() === month && aDate.getDate() === d;
      });
      days.push({ date: dayDate, isCurrentMonth: true, isToday, appointments: appts });
    }

    // Fill trailing days
    const remaining = 42 - days.length;
    for (let d = 1; d <= remaining; d++) {
      const dayDate = new Date(year, month + 1, d);
      days.push({ date: dayDate, isCurrentMonth: false, isToday: false, appointments: [] });
    }

    return days;
  });

  weekDays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  statusColors: Record<string, string> = {
    scheduled: 'bg-blue-500',
    confirmed: 'bg-emerald-500',
    in_progress: 'bg-amber-500',
    completed: 'bg-zinc-400',
    cancelled: 'bg-red-500',
    no_show: 'bg-orange-500',
  };

  ngOnInit() {
    this.loadData();
  }

  loadData() {
    this.loading.set(true);
    forkJoin({
      services: this.servicesService.getServices(),
      appointments: this.servicesService.getAppointments(),
      staff: this.staffService.listStaff(),
    }).subscribe({
      next: ({ services, appointments, staff }) => {
        this.services.set(services || []);
        this.appointments.set(appointments || []);
        this.staff.set(staff || []);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
      }
    });
  }

  prevMonth() {
    const d = new Date(this.currentDate());
    d.setMonth(d.getMonth() - 1);
    this.currentDate.set(d);
  }

  nextMonth() {
    const d = new Date(this.currentDate());
    d.setMonth(d.getMonth() + 1);
    this.currentDate.set(d);
  }

  goToToday() {
    this.currentDate.set(new Date());
  }

  openBookingModal(date?: Date) {
    const startTime = date
      ? new Date(date.getFullYear(), date.getMonth(), date.getDate(), 9, 0)
      : new Date();
    this.newBooking.set({
      service_id: '',
      staff_member_id: '',
      start_time: this.toLocalDatetimeString(startTime),
      customer_name: '',
      customer_phone: '',
      customer_email: '',
      notes: '',
    });
    this.showBookingModal.set(true);
  }

  closeBookingModal() {
    this.showBookingModal.set(false);
  }

  saveBooking() {
    const b = this.newBooking();
    if (!b.service_id || !b.start_time) {
      this.toastr.warning('Please fill in the required fields', 'Missing Info');
      return;
    }
    this.savingBooking.set(true);

    const service = this.services().find(s => (s as any).id === b.service_id);
    const durationMs = (service?.duration_minutes || 60) * 60 * 1000;
    const startMs = new Date(b.start_time).getTime();
    const endTime = new Date(startMs + durationMs).toISOString();

    const payload = {
      service_id: b.service_id,
      staff_member_id: b.staff_member_id || undefined,
      start_time: new Date(b.start_time).toISOString(),
      end_time: endTime,
      customer_name: b.customer_name,
      customer_phone: b.customer_phone,
      customer_email: b.customer_email,
      notes: b.notes,
      status: 'scheduled',
    };

    this.servicesService.createAppointment(payload as any).subscribe({
      next: () => {
        this.toastr.success('Appointment booked successfully');
        this.closeBookingModal();
        this.servicesService.getAppointments().subscribe(a => this.appointments.set(a || []));
        this.savingBooking.set(false);
      },
      error: (e) => {
        this.toastr.error(e.error?.error || 'Failed to save booking');
        this.savingBooking.set(false);
      }
    });
  }

  openServiceModal() {
    this.newService.set({ name: '', description: '', price: 0, duration_minutes: 30, is_active: true });
    this.showServiceModal.set(true);
  }

  closeServiceModal() {
    this.showServiceModal.set(false);
  }

  saveService() {
    const s = this.newService();
    if (!s.name) {
      this.toastr.warning('Service name is required');
      return;
    }
    this.savingService.set(true);
    this.servicesService.createService(s).subscribe({
      next: () => {
        this.toastr.success('Service created successfully');
        this.closeServiceModal();
        this.servicesService.getServices().subscribe(sv => this.services.set(sv || []));
        this.savingService.set(false);
      },
      error: (e) => {
        this.toastr.error(e.error?.error || 'Failed to create service');
        this.savingService.set(false);
      }
    });
  }

  getStaffName(id: string): string {
    const s = this.staff().find(m => (m as any).id === id || (m as any).user_id === id);
    return s ? `${(s as any).first_name || ''} ${(s as any).last_name || ''}`.trim() : 'Unassigned';
  }

  getServiceName(id: string): string {
    const s = this.services().find(sv => (sv as any).id === id);
    return s?.name || '—';
  }

  formatTime(iso: string): string {
    return new Date(iso).toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true });
  }

  formatDate(iso: string): string {
    return new Date(iso).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
  }

  getAppointmentHour(apt: Appointment): number {
    return new Date(apt.start_time).getHours();
  }

  private toLocalDatetimeString(d: Date): string {
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }

  upcomingAppointments = computed(() => {
    const now = new Date();
    return this.appointments()
      .filter(a => new Date(a.start_time) >= now && a.status !== 'cancelled')
      .sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime())
      .slice(0, 5);
  });

  todaysAppointments = computed(() => {
    const today = new Date();
    return this.appointments().filter(a => {
      const d = new Date(a.start_time);
      return d.toDateString() === today.toDateString();
    });
  });
}
