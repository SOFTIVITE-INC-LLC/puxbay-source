import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface PlanFeature {
  id?: string;
  plan_id?: string;
  text: string;
  is_available: boolean;
  order_index?: number;
}

export interface PricingPlan {
  id?: string;
  name: string;
  slug: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  currency: string;
  is_popular: boolean;
  button_text: string;
  order_index: number;
  max_branches?: number;
  max_staff?: number;
  features: PlanFeature[];
}

@Injectable({
  providedIn: 'root'
})
export class AdminPricingService {
  private http = inject(HttpClient);
  private apiUrl = '/api/v1/admin/pricing-plans';

  getPlans(): Observable<PricingPlan[]> {
    return this.http.get<PricingPlan[]>(this.apiUrl);
  }

  createPlan(plan: PricingPlan): Observable<PricingPlan> {
    return this.http.post<PricingPlan>(this.apiUrl, plan);
  }

  updatePlan(id: string, plan: PricingPlan): Observable<PricingPlan> {
    return this.http.put<PricingPlan>(`${this.apiUrl}/${id}`, plan);
  }

  deletePlan(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }
}
