import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy } from '@angular/core';

@Component({
  selector: 'app-empty-state',
  standalone: true,
  imports: [],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div
      class="flex flex-col items-center justify-center p-8 text-center bg-white dark:bg-zinc-900 rounded-3xl border border-slate-100 dark:border-zinc-800 shadow-sm"
    >
      <div
        class="w-16 h-16 rounded-2xl bg-primary-50 dark:bg-primary-500/10 text-primary-500 flex items-center justify-center mb-4"
      >
        <span class="material-symbols-outlined !text-[32px]">{{ icon }}</span>
      </div>
      <h3 class="text-lg font-bold text-slate-900 dark:text-white mb-2">{{ title }}</h3>
      <p class="text-sm text-slate-500 mb-6 max-w-sm">{{ description }}</p>

      @if (actionLabel) {
        <button
          (click)="action.emit()"
          class="px-5 py-2.5 rounded-xl bg-primary-600 text-white font-bold text-sm hover:bg-primary-700 transition-colors shadow-lg shadow-primary-500/20"
        >
          {{ actionLabel }}
        </button>
      }
    </div>
  `,
})
export class EmptyStateComponent {
  @Input() icon: string = 'inbox';
  @Input() title: string = 'No Data Found';
  @Input() description: string = 'There is currently no data to display here.';
  @Input() actionLabel?: string;
  @Output() action = new EventEmitter<void>();
}
