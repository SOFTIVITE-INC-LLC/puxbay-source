import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import { delay } from 'rxjs/operators';

export interface FeatureFlags {
  enable_new_dashboard: boolean;
  enable_pos_beta: boolean;
  maintenance_mode: boolean;
  enable_api_keys: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class SettingsService {
  
  // Mock state since backend KV store is pending
  private flags: FeatureFlags = {
    enable_new_dashboard: true,
    enable_pos_beta: false,
    maintenance_mode: false,
    enable_api_keys: true
  };

  getFeatureFlags(): Observable<FeatureFlags> {
    return of({...this.flags}).pipe(delay(500));
  }

  updateFeatureFlags(flags: FeatureFlags): Observable<any> {
    this.flags = {...flags};
    return of({ status: 'updated' }).pipe(delay(800));
  }
}
