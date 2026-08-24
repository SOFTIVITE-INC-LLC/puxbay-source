import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from './api.service';
import { Observable, tap } from 'rxjs';
import { Service, Appointment } from '../models/services.models';

export interface ServiceCreateInput {
  name: string;
  description?: string;
  price: number;
  duration_min: number;
}

export interface AppointmentCreateInput {
  service_id: string;
  customer_id: string;
  staff_member_id: string;
  start_time: string; // ISO string
}

@Injectable({
  providedIn: 'root'
})
export class ServicesService {
  private api = inject(ApiService);
  
  services = signal<Service[]>([]);
  appointments = signal<Appointment[]>([]);
  loading = signal<boolean>(false);

  getServices(): Observable<Service[]> {
    this.loading.set(true);
    return this.api.get<Service[]>('/services').pipe(
      tap(res => {
        this.services.set(res || []);
        this.loading.set(false);
      })
    );
  }

  createService(input: ServiceCreateInput): Observable<Service> {
    return this.api.post<Service>('/services', input).pipe(
      tap(s => this.services.update(list => [...list, s]))
    );
  }

  getAppointments(): Observable<Appointment[]> {
    return this.api.get<Appointment[]>('/services/appointments').pipe(
      tap(res => {
        this.appointments.set(res || []);
      })
    );
  }

  createAppointment(input: AppointmentCreateInput): Observable<Appointment> {
    return this.api.post<Appointment>('/services/appointments', input).pipe(
      tap(a => this.appointments.update(list => [...list, a]))
    );
  }

  listCommissions(): Observable<any[]> { return this.api.get<any[]>('/services/commissions'); }
  markCommissionsPaid(data: any): Observable<any> { return this.api.post<any>('/services/commissions/mark-paid', data); }
}
