import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { GrowthService, ReferralReward } from '../../services/growth.service';

@Component({
  selector: 'app-referrals',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './referrals.html',
})
export class ReferralsComponent implements OnInit {
  private service = inject(GrowthService);
  referrals = signal<ReferralReward[]>([]);
  isLoading = signal(true);

  get totalRewards() {
    return this.referrals().reduce((s, r) => s + r.reward_amount, 0);
  }
  get applied() { return this.referrals().filter(r => r.is_applied).length; }
  get pending() { return this.referrals().filter(r => !r.is_applied).length; }

  ngOnInit() {
    this.service.getReferrals().subscribe({
      next: (res) => { this.referrals.set(res.data || []); this.isLoading.set(false); },
      error: () => this.isLoading.set(false)
    });
  }
}
