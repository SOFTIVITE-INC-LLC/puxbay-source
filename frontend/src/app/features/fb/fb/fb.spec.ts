import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Fb } from './fb';

describe('Fb', () => {
  let component: Fb;
  let fixture: ComponentFixture<Fb>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Fb],
    }).compileComponents();

    fixture = TestBed.createComponent(Fb);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
