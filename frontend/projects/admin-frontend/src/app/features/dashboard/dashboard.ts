import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DashboardService } from '../../services/dashboard.service';
import { BaseChartDirective } from 'ng2-charts';
import { ChartConfiguration, ChartOptions, ChartType } from 'chart.js';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, BaseChartDirective],
  templateUrl: './dashboard.html',
})
export class DashboardComponent implements OnInit {
  private service = inject(DashboardService);

  healthStatus = signal<{status: string, version: string, latency_ms: number} | null>(null);
  isLoading = signal(true);
  
  // Stats for the UI
  stats = signal<any>({
    mrr: 0,
    total_tenants: 0,
    active_users: 0,
    total_orders: 0,
    active_trials: 0,
    trial_conversion_rate: 0,
    churn_rate: 0,
  });

  // Chart JS Config
  public lineChartData: ChartConfiguration<'line'>['data'] = {
    labels: [],
    datasets: [
      {
        data: [],
        label: 'Active Tenants',
        fill: true,
        tension: 0.5,
        borderColor: '#005b96',
        backgroundColor: 'rgba(0, 91, 150, 0.12)',
        pointBackgroundColor: '#005b96',
        pointBorderColor: '#fff',
      }
    ]
  };
  public lineChartOptions: ChartOptions<'line'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false }
    },
    scales: {
      y: { beginAtZero: true, grid: { display: true, color: '#f1f5f9' }, border: { dash: [4, 4] } },
      x: { grid: { display: false } }
    }
  };

  recentActivities: any[] = [];

  ngOnInit() {
    this.checkHealth();
    this.loadStats();
  }

  checkHealth() {
    this.service.getSystemHealth().subscribe({
      next: (data) => {
        this.healthStatus.set(data);
      },
      error: (err) => {
        console.error('Failed to load system health', err);
      }
    });
  }

  loadStats() {
    this.isLoading.set(true);
    this.service.getDashboardStats().subscribe({
      next: (data) => {
        try {
          if (data) {
            this.stats.set(data);
            
            if (data.recent_activities) {
              this.recentActivities = data.recent_activities;
            }

            if (Array.isArray(data.platform_growth) && data.platform_growth.length > 0) {
              const growth: any[] = data.platform_growth;
              this.lineChartData = {
                labels: growth.map(g => g.label || ''),
                datasets: [
                  {
                    data: growth.map(g => g.value || 0),
                    label: 'Active Tenants',
                    fill: true,
                    tension: 0.5,
                    borderColor: '#005b96',
                    backgroundColor: 'rgba(0, 91, 150, 0.12)',
                    pointBackgroundColor: '#005b96',
                    pointBorderColor: '#fff',
                  }
                ]
              };
            }
          }
        } catch (e) {
          console.error('Error processing dashboard data', e);
        } finally {
          this.isLoading.set(false);
        }
      },
      error: (err) => {
        console.error('Failed to load dashboard stats', err);
        this.isLoading.set(false);
      }
    });
  }
}
