import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PricingPlans } from './pricing-plans';

describe('PricingPlans', () => {
  let component: PricingPlans;
  let fixture: ComponentFixture<PricingPlans>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PricingPlans],
    }).compileComponents();

    fixture = TestBed.createComponent(PricingPlans);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
