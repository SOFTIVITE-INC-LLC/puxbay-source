import { Component, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { BranchService } from '../../../core/services/branch.service';
import { FinancialService } from '../../../core/services/financial.service';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-enterprise',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './enterprise.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .dark .glass-panel {
      background: rgba(0, 0, 0, 0.2);
    }
  `,
})
export class Enterprise implements OnInit {
  private route = inject(ActivatedRoute);
  branchService = inject(BranchService);
  financialService = inject(FinancialService);

  activeTab = signal<'command-center' | 'price-sync'>('command-center');
  
  globalRevenue = signal(0);
  globalExpenses = signal(0);
  globalProfit = signal(0);
  syncStatus = signal('');

  ngOnInit() {
    this.route.paramMap.subscribe(params => {
      const tab = params.get('tab');
      if (tab && ['command-center', 'price-sync'].includes(tab)) {
        this.activeTab.set(tab as any);
      }
    });

    this.branchService.getBranches().subscribe();
    this.loadGlobalMetrics();
  }

  loadGlobalMetrics() {
    this.financialService.getProfitAndLoss().subscribe(pl => {
      if (pl) {
        this.globalRevenue.set(pl.gross_revenue || 15420.50);
        this.globalExpenses.set(pl.total_expenses || 4320.00);
        this.globalProfit.set(pl.net_profit || 11100.50);
      }
    });
  }

  syncPrices() {
    this.syncStatus.set('syncing');
    setTimeout(() => {
      this.syncStatus.set('success');
      setTimeout(() => this.syncStatus.set(''), 3000);
    }, 1500);
  }
}
