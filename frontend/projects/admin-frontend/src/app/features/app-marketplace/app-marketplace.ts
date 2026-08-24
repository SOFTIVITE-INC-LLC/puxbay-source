import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MarketplaceService, ExternalSystem } from '../../services/marketplace.service';

@Component({
  selector: 'app-app-marketplace',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './app-marketplace.html',
})
export class AppMarketplaceComponent implements OnInit {
  private service = inject(MarketplaceService);

  apps = signal<ExternalSystem[]>([]);
  isLoading = signal(true);

  ngOnInit() {
    this.loadApps();
  }

  loadApps() {
    this.isLoading.set(true);
    this.service.getApps().subscribe({
      next: (data) => {
        this.apps.set(data || []);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load apps', err);
        this.isLoading.set(false);
      }
    });
  }

  toggle(id: string, field: 'active' | 'public') {
    this.service.toggleApp(id, field).subscribe({
      next: () => this.loadApps(),
      error: (err) => console.error(`Failed to toggle ${field}`, err)
    });
  }
}
