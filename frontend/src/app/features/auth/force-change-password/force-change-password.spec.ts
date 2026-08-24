import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ForceChangePassword } from './force-change-password';

describe('ForceChangePassword', () => {
  let component: ForceChangePassword;
  let fixture: ComponentFixture<ForceChangePassword>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ForceChangePassword],
    }).compileComponents();

    fixture = TestBed.createComponent(ForceChangePassword);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
