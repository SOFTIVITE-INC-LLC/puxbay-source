import { ComponentFixture, TestBed } from '@angular/core/testing';

import { StocktakePortal } from './stocktake-portal';

describe('StocktakePortal', () => {
  let component: StocktakePortal;
  let fixture: ComponentFixture<StocktakePortal>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [StocktakePortal],
    }).compileComponents();

    fixture = TestBed.createComponent(StocktakePortal);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
