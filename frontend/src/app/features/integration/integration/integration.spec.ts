import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Integration } from './integration';

describe('Integration', () => {
  let component: Integration;
  let fixture: ComponentFixture<Integration>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Integration],
    }).compileComponents();

    fixture = TestBed.createComponent(Integration);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
