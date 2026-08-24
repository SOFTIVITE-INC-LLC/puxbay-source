import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PromoCodes } from './promo-codes';

describe('PromoCodes', () => {
  let component: PromoCodes;
  let fixture: ComponentFixture<PromoCodes>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PromoCodes],
    }).compileComponents();

    fixture = TestBed.createComponent(PromoCodes);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
