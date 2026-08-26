import { Injectable, inject } from '@angular/core';
import { ApiService } from './api.service';
import { Observable } from 'rxjs';

export interface Device {
  id: string;
  branch_id?: string;
  name: string;
  device_type: string;
  ip_address: string;
  mac_address: string;
  status: string;
  config: any;
  created_at: string;
  updated_at: string;
}

export interface DeviceCreateInput {
  branch_id?: string;
  name: string;
  device_type: string;
  ip_address?: string;
  mac_address?: string;
  config?: string; // JSON string
}

@Injectable({
  providedIn: 'root'
})
export class DeviceService {
  private api = inject(ApiService);

  listDevices(branchId?: string): Observable<Device[]> {
    const options = branchId ? { params: { branch_id: branchId } as Record<string, string> } : undefined;
    return this.api.get<Device[]>('/devices', options);
  }

  getDevice(id: string): Observable<Device> {
    return this.api.get<Device>(`/devices/${id}`);
  }

  createDevice(input: DeviceCreateInput): Observable<Device> {
    return this.api.post<Device>('/devices', input);
  }

  updateDevice(id: string, input: DeviceCreateInput): Observable<Device> {
    return this.api.put<Device>(`/devices/${id}`, input);
  }

  deleteDevice(id: string): Observable<void> {
    return this.api.delete<void>(`/devices/${id}`);
  }
}
