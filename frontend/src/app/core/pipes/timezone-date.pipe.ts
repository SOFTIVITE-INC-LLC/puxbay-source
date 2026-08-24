import { Pipe, PipeTransform, inject } from '@angular/core';
import { DatePipe } from '@angular/common';
import { SettingsService } from '../services/settings.service';

@Pipe({
  name: 'timezoneDate',
  standalone: true
})
export class TimezoneDatePipe implements PipeTransform {
  private settingsService = inject(SettingsService, { optional: true });
  constructor(private datePipe: DatePipe) {}

  transform(value: string | Date | number | null | undefined, format: string = 'medium', timezone?: string): string | null {
    if (!value) return null;
    
    // Fetch timezone from user settings/tenant store or default to local browser timezone
    let tz = timezone;
    if (!tz) {
      const settings = this.settingsService?.settings();
      tz = settings?.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone;
    }
    
    return this.datePipe.transform(value, format, tz);
  }
}
