import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface PlanFeature {
  text: string;
  is_available: boolean;
}

export interface PricingPlan {
  id: string;
  name: string;
  slug: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  currency: string;
  is_popular: boolean;
  button_text: string;
  features: PlanFeature[];
}

@Injectable({
  providedIn: 'root'
})
export class PricingService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/marketing/pricing-plans`;

  getPublicPricingPlans(): Observable<PricingPlan[]> {
    return this.http.get<PricingPlan[]>(this.apiUrl);
  }

  // Admin endpoints can also go here later
}
