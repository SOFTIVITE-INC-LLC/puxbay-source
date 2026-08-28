import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SupplierPortalService, DashboardStats } from '../../services/supplier-portal.service';

@Component({
  selector: 'app-supplier-portal-scorecard',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './scorecard.component.html'
})
export class SupplierPortalScorecardComponent implements OnInit {
  portalService = inject(SupplierPortalService);

  stats = signal<DashboardStats>({
    total_pos: 0,
    pending_deliveries: 0,
    total_invoiced: 0,
    open_quotes: 0,
    otd_score: 98.5
  });

  ngOnInit() {
    this.portalService.getDashboard().subscribe({
      next: (res) => {
        if (res) this.stats.set(res);
      }
    });
  }
}
