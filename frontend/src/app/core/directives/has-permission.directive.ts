import { Directive, Input, TemplateRef, ViewContainerRef, inject, effect } from '@angular/core';
import { AuthService } from '../services/auth.service';

@Directive({
  selector: '[appHasPermission]',
  standalone: true
})
export class HasPermissionDirective {
  private templateRef = inject(TemplateRef<any>);
  private viewContainer = inject(ViewContainerRef);
  private authService = inject(AuthService);

  private hasView = false;
  private requiredPermissions: string[] = [];

  @Input() set appHasPermission(permission: string | string[]) {
    this.requiredPermissions = Array.isArray(permission) ? permission : [permission];
    this.updateView();
  }

  constructor() {
    effect(() => {
      // Re-evaluate when user state changes
      const user = this.authService.currentUser();
      this.updateView();
    });
  }

  private updateView() {
    const hasAccess = this.authService.hasPermission(this.requiredPermissions);
    
    if (hasAccess && !this.hasView) {
      this.viewContainer.createEmbeddedView(this.templateRef);
      this.hasView = true;
    } else if (!hasAccess && this.hasView) {
      this.viewContainer.clear();
      this.hasView = false;
    }
  }
}
