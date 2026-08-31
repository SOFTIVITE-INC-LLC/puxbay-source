import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { RouterModule } from '@angular/router';

interface OnboardingStep {
  id: string;
  title: string;
  description: string;
  completed: boolean;
  action_url: string;
  action_text: string;
}

interface OnboardingStatus {
  progress_percent: number;
  steps: OnboardingStep[];
}

@Component({
  selector: 'app-onboarding-widget',
  standalone: true,
  imports: [CommonModule, RouterModule],
  template: `
    <div class="rounded-2xl bg-white dark:bg-[#1a1f2e] border border-slate-200 dark:border-white/[0.08] shadow-sm mb-8 overflow-hidden" *ngIf="status() && status()!.progress_percent < 100">
      <div class="p-6">
        <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-5">
          <div>
            <h2 class="text-xl font-black text-slate-900 dark:text-white tracking-tight">Welcome to Puxbay! Let's get you set up.</h2>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1 font-medium">Complete these steps to unlock the full potential of your store.</p>
          </div>
          <span class="px-3 py-1.5 rounded-lg bg-indigo-50 dark:bg-indigo-500/10 border border-indigo-100 dark:border-indigo-500/20 text-indigo-600 dark:text-indigo-400 text-xs font-bold tracking-wide">
            {{ status()?.progress_percent }}% Complete
          </span>
        </div>
        
        <div class="w-full h-2 bg-slate-100 dark:bg-white/[0.06] rounded-full overflow-hidden mb-8">
          <div class="h-full bg-indigo-500 rounded-full transition-all duration-700 ease-out" [style.width.%]="status()?.progress_percent"></div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div *ngFor="let step of status()?.steps; let i = index" 
               class="relative p-5 rounded-2xl border transition-all duration-300"
               [ngClass]="step.completed 
                 ? 'bg-emerald-50/50 dark:bg-emerald-500/5 border-emerald-200 dark:border-emerald-500/20' 
                 : 'bg-slate-50 dark:bg-white/[0.02] border-slate-200 dark:border-white/[0.06] hover:border-indigo-300 dark:hover:border-white/20 hover:shadow-md'">
            
            <div class="flex flex-col h-full">
              <div class="flex items-center gap-3 mb-3">
                <div *ngIf="step.completed" class="w-8 h-8 rounded-full bg-emerald-100 dark:bg-emerald-500/20 flex items-center justify-center text-emerald-600 dark:text-emerald-400 shrink-0">
                  <span class="material-symbols-outlined !text-[18px] font-bold">check</span>
                </div>
                <div *ngIf="!step.completed" class="w-8 h-8 rounded-full bg-white dark:bg-[#1a1f2e] border border-slate-300 dark:border-white/20 flex items-center justify-center text-slate-500 dark:text-slate-400 text-xs font-black shrink-0">
                  {{ i + 1 }}
                </div>
                <h3 class="font-bold text-sm leading-tight" 
                    [ngClass]="step.completed ? 'text-slate-500 dark:text-slate-400 line-through decoration-slate-300 dark:decoration-slate-600' : 'text-slate-900 dark:text-white'">
                  {{ step.title }}
                </h3>
              </div>
              
              <p class="text-xs text-slate-500 dark:text-slate-400 mb-4 flex-grow pr-2"
                 [ngClass]="{'opacity-60': step.completed}">
                {{ step.description }}
              </p>
              
              <div class="mt-auto">
                <a *ngIf="!step.completed" [routerLink]="step.action_url" 
                   class="inline-flex w-full justify-center items-center gap-1.5 px-4 py-2.5 rounded-xl bg-white dark:bg-white/[0.06] border border-slate-200 dark:border-white/10 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-white/[0.1] transition-all font-bold text-xs shadow-sm hover:text-indigo-600 dark:hover:text-indigo-400">
                  {{ step.action_text }}
                  <span class="material-symbols-outlined !text-[14px]">arrow_forward</span>
                </a>
                <div *ngIf="step.completed" class="w-full flex items-center justify-center py-2.5 text-xs font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-100/50 dark:bg-emerald-500/10 rounded-xl">
                  Completed
                </div>
              </div>
            </div>
            
          </div>
        </div>
      </div>
    </div>
  `
})
export class OnboardingWidgetComponent implements OnInit {
  private http = inject(HttpClient);
  status = signal<OnboardingStatus | null>(null);

  ngOnInit() {
    this.http.get<OnboardingStatus>('/api/v1/onboarding/status').subscribe({
      next: (data) => this.status.set(data),
      error: (err) => console.error('Failed to load onboarding status', err)
    });
  }
}
