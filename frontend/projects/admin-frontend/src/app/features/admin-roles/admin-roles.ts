import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SecurityService, AdminRole, AVAILABLE_PERMISSIONS } from '../../services/security.service';

@Component({
  selector: 'app-admin-roles',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './admin-roles.html',
})
export class AdminRolesComponent implements OnInit {
  private service = inject(SecurityService);
  roles = signal<AdminRole[]>([]);
  isLoading = signal(true);
  isSaving = signal(false);

  availablePermissions = AVAILABLE_PERMISSIONS;

  selectedRole = signal<AdminRole | null>(null);
  selectedPermissions = signal<string[]>([]);
  isModalOpen = signal(false);
  newRoleName = signal('');

  ngOnInit() {
    this.loadRoles();
  }

  loadRoles() {
    this.isLoading.set(true);
    this.service.getRoles().subscribe({
      next: (res) => { this.roles.set(res.data || []); this.isLoading.set(false); },
      error: () => this.isLoading.set(false)
    });
  }

  openRole(role: AdminRole) {
    this.selectedRole.set(role);
    let perms = [];
    try {
      perms = typeof role.permissions === 'string' ? JSON.parse(role.permissions || '[]') : (role.permissions || []);
    } catch { perms = []; }
    this.selectedPermissions.set(perms);
  }

  togglePermission(permId: string) {
    const perms = this.selectedPermissions();
    if (perms.includes(permId)) {
      this.selectedPermissions.set(perms.filter(p => p !== permId));
    } else {
      this.selectedPermissions.set([...perms, permId]);
    }
  }

  savePermissions() {
    const role = this.selectedRole();
    if (!role) return;
    this.isSaving.set(true);
    const permsStr = JSON.stringify(this.selectedPermissions());
    this.service.updateRolePermissions(role.id, permsStr).subscribe({
      next: () => {
        this.isSaving.set(false);
        this.loadRoles();
      },
      error: () => this.isSaving.set(false)
    });
  }

  openCreateModal() {
    this.newRoleName.set('');
    this.isModalOpen.set(true);
  }

  closeCreateModal() {
    this.isModalOpen.set(false);
  }

  createRole() {
    if (!this.newRoleName()) return;
    this.isSaving.set(true);
    this.service.createRole({ name: this.newRoleName(), permissions: '[]' }).subscribe({
      next: () => {
        this.isSaving.set(false);
        this.closeCreateModal();
        this.loadRoles();
      },
      error: () => this.isSaving.set(false)
    });
  }

  getPermCount(role: AdminRole): number {
    try {
      const perms = typeof role.permissions === 'string' ? JSON.parse(role.permissions || '[]') : (role.permissions || []);
      return perms.length;
    } catch { return 0; }
  }

  selectAll() {
    this.selectedPermissions.set(this.availablePermissions.map(p => p.id));
  }

  clearAll() {
    this.selectedPermissions.set([]);
  }
}
