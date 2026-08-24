import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Broadcasts } from './broadcasts';

describe('Broadcasts', () => {
  let component: Broadcasts;
  let fixture: ComponentFixture<Broadcasts>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Broadcasts],
    }).compileComponents();

    fixture = TestBed.createComponent(Broadcasts);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
