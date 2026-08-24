import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Storefront } from './storefront';

describe('Storefront', () => {
  let component: Storefront;
  let fixture: ComponentFixture<Storefront>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Storefront],
    }).compileComponents();

    fixture = TestBed.createComponent(Storefront);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
