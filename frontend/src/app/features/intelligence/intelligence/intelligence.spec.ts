import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Intelligence } from './intelligence';

describe('Intelligence', () => {
  let component: Intelligence;
  let fixture: ComponentFixture<Intelligence>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Intelligence],
    }).compileComponents();

    fixture = TestBed.createComponent(Intelligence);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
