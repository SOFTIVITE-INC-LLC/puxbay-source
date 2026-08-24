import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Marketing } from './marketing';

describe('Marketing', () => {
  let component: Marketing;
  let fixture: ComponentFixture<Marketing>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Marketing],
    }).compileComponents();

    fixture = TestBed.createComponent(Marketing);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
