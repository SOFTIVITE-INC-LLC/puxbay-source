import { Injectable, inject } from '@angular/core';
import { ApiService } from './api.service';
import { Observable } from 'rxjs';
import { UserProfile } from '../models/user.models';

export interface StaffCreateInput {
  username: string;
  email: string;
  phone?: string;
  password?: string;
  first_name: string;
  last_name: string;
  role: string;
  branch_id?: string;
}

@Injectable({
  providedIn: 'root'
})
export class StaffService {
  private api = inject(ApiService);
  private readonly baseUrl = '/staff';

  listStaff(): Observable<UserProfile[]> {
    return this.api.get<UserProfile[]>(this.baseUrl);
  }

  createStaff(input: StaffCreateInput): Observable<UserProfile> {
    return this.api.post<UserProfile>(this.baseUrl, input);
  }

  getStaff(id: string): Observable<UserProfile> {
    return this.api.get<UserProfile>(`${this.baseUrl}/${id}`);
  }

  updateStaff(id: string, input: Partial<StaffCreateInput>): Observable<UserProfile> {
    return this.api.put<UserProfile>(`${this.baseUrl}/${id}`, input);
  }

  deleteStaff(id: string): Observable<void> {
    return this.api.delete<void>(`${this.baseUrl}/${id}`);
  }
}
