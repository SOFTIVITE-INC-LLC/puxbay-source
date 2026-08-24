import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Branches } from './branches';

describe('Branches', () => {
  let component: Branches;
  let fixture: ComponentFixture<Branches>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Branches],
    }).compileComponents();

    fixture = TestBed.createComponent(Branches);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
