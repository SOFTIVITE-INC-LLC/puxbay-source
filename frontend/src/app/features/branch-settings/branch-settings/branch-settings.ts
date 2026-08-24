import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { BranchService } from '../../../core/services/branch.service';
import { Branch } from '../../../core/models/branch.models';

type SettingsTab = 'general' | 'appearance' | 'pos' | 'inventory';

@Component({
  selector: 'app-branch-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './branch-settings.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.7);
      backdrop-filter: blur(24px);
      -webkit-backdrop-filter: blur(24px);
      border: 1px solid rgba(255, 255, 255, 0.8);
      box-shadow: 0 4px 30px rgba(0, 0, 0, 0.05);
    }
    :host-context(.dark) .glass-panel {
      background: rgba(24, 24, 27, 0.6);
      border: 1px solid rgba(255, 255, 255, 0.08);
      box-shadow: 0 4px 30px rgba(0, 0, 0, 0.3);
    }
    .input-premium {
      @apply w-full pl-11 pr-4 py-3.5 rounded-2xl bg-white/60 dark:bg-black/30 text-zinc-900 dark:text-white border border-zinc-200/80 dark:border-zinc-700/50 focus:outline-none focus:border-indigo-500 focus:ring-4 focus:ring-indigo-500/20 font-bold transition-all shadow-sm backdrop-blur-sm;
    }
    .select-premium {
      @apply w-full pl-11 pr-10 py-3.5 rounded-2xl bg-white/60 dark:bg-black/30 text-zinc-900 dark:text-white border border-zinc-200/80 dark:border-zinc-700/50 focus:outline-none focus:border-indigo-500 focus:ring-4 focus:ring-indigo-500/20 font-bold transition-all shadow-sm backdrop-blur-sm appearance-none;
    }
    .textarea-premium {
      @apply w-full pl-4 pr-4 py-3.5 rounded-2xl bg-white/60 dark:bg-black/30 text-zinc-900 dark:text-white border border-zinc-200/80 dark:border-zinc-700/50 focus:outline-none focus:border-indigo-500 focus:ring-4 focus:ring-indigo-500/20 font-mono text-sm transition-all shadow-sm backdrop-blur-sm resize-none;
    }
  `
})
export class BranchSettings implements OnInit {
  protected readonly Math = Math;
  branchService = inject(BranchService);

  activeTab = signal<SettingsTab>('general');
  saving = signal(false);
  saved = signal(false);

  form = signal<Partial<Branch>>({});

  readonly currencies = [
    { code: 'USD', symbol: '$',   label: 'USD – US Dollar' },
    { code: 'GHS', symbol: 'GH₵', label: 'GHS – Ghana Cedi' },
    { code: 'NGN', symbol: '₦',   label: 'NGN – Nigerian Naira' },
    { code: 'ZAR', symbol: 'R',   label: 'ZAR – South African Rand' },
    { code: 'EUR', symbol: '€',   label: 'EUR – Euro' },
    { code: 'GBP', symbol: '£',   label: 'GBP – British Pound' },
  ];

  ngOnInit() {
    const branch = this.branchService.activeBranch();
    if (branch) {
      this.form.set({ ...branch });
    }
  }

  onCurrencyChange(code: string) {
    const match = this.currencies.find(c => c.code === code);
    if (match) {
      this.form.update(f => ({ ...f, currency_code: match.code, currency_symbol: match.symbol }));
    }
  }

  save() {
    const branch = this.branchService.activeBranch();
    if (!branch?.id) return;

    this.saving.set(true);
    this.branchService.updateBranch(branch.id, this.form() as any).subscribe({
      next: (updated) => {
        // Keep localStorage in sync so branch header updates
        this.branchService.setActiveBranch(updated);
        this.saving.set(false);
        this.saved.set(true);
        setTimeout(() => this.saved.set(false), 3000);
      },
      error: () => this.saving.set(false)
    });
  }
}
