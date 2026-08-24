import { Injectable, inject } from '@angular/core';
import { ApiService } from './api.service';
import { Observable } from 'rxjs';

export interface Permission {
  id: string;
  code: string;
  description: string;
  module: string;
}

export interface Role {
  id: string;
  name: string;
  description: string;
  is_system: boolean;
  permissions?: Permission[];
}

@Injectable({
  providedIn: 'root'
})
export class RolesService {
  private api = inject(ApiService);

  getRoles(): Observable<Role[]> {
    return this.api.get<Role[]>('/roles');
  }

  getPermissions(): Observable<Permission[]> {
    return this.api.get<Permission[]>('/permissions');
  }

  createRole(data: any): Observable<Role> {
    return this.api.post<Role>('/roles', data);
  }

  updateRole(id: string, data: any): Observable<Role> {
    return this.api.put<Role>(`/roles/${id}`, data);
  }

  deleteRole(id: string): Observable<any> {
    return this.api.delete(`/roles/${id}`);
  }
}
