import { Component, Input, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-skeleton',
  standalone: true,
  imports: [CommonModule],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div [ngClass]="['animate-pulse bg-slate-200 dark:bg-zinc-800 rounded', widthClass, heightClass, customClass]"></div>
  `
})
export class SkeletonComponent {
  @Input() widthClass: string = 'w-full';
  @Input() heightClass: string = 'h-4';
  @Input() customClass: string = '';
}
