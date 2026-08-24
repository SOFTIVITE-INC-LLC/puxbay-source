import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { AuthGuard } from './auth.guard';
import { AuthService } from '../services/auth.service';

describe('AuthGuard', () => {
  let authServiceSpy: jasmine.SpyObj<AuthService>;
  let routerSpy: jasmine.SpyObj<Router>;

  beforeEach(() => {
    authServiceSpy = jasmine.createSpyObj('AuthService', ['isAuthenticated']);
    routerSpy = jasmine.createSpyObj('Router', ['navigate']);

    TestBed.configureTestingModule({
      providers: [
        { provide: AuthService, useValue: authServiceSpy },
        { provide: Router, useValue: routerSpy }
      ]
    });
  });

  it('should be created', () => {
    TestBed.runInInjectionContext(() => {
      expect(AuthGuard).toBeTruthy();
    });
  });

  it('should return true if user is authenticated', () => {
    authServiceSpy.isAuthenticated.and.returnValue(true);
    
    let result: boolean = false;
    TestBed.runInInjectionContext(() => {
      // In Angular v16+, function guards are executed in an injection context
      result = AuthGuard({} as any, {} as any) as boolean;
    });

    expect(result).toBe(true);
    expect(routerSpy.navigate).not.toHaveBeenCalled();
  });

  it('should return false and navigate to login if user is not authenticated', () => {
    authServiceSpy.isAuthenticated.and.returnValue(false);
    
    let result: boolean = true;
    TestBed.runInInjectionContext(() => {
      result = AuthGuard({} as any, {} as any) as boolean;
    });

    expect(result).toBe(false);
    expect(routerSpy.navigate).toHaveBeenCalledWith(['/login']);
  });
});
