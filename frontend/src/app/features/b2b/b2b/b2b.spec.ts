import { ComponentFixture, TestBed } from '@angular/core/testing';

import { B2b } from './b2b';

describe('B2b', () => {
  let component: B2b;
  let fixture: ComponentFixture<B2b>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [B2b],
    }).compileComponents();

    fixture = TestBed.createComponent(B2b);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
