import { ComponentFixture, TestBed } from '@angular/core/testing';

import { StocktakeDetail } from './stocktake-detail';

describe('StocktakeDetail', () => {
  let component: StocktakeDetail;
  let fixture: ComponentFixture<StocktakeDetail>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [StocktakeDetail],
    }).compileComponents();

    fixture = TestBed.createComponent(StocktakeDetail);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
