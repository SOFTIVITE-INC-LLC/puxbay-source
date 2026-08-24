import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AppMarketplace } from './app-marketplace';

describe('AppMarketplace', () => {
  let component: AppMarketplace;
  let fixture: ComponentFixture<AppMarketplace>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppMarketplace],
    }).compileComponents();

    fixture = TestBed.createComponent(AppMarketplace);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
