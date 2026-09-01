import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators, FormArray } from '@angular/forms';
import { RolesService, Role, Permission } from '../../../core/services/roles.service';
import { ToastService } from '../../../core/services/toast';
import { AlertService } from '../../../core/services/alert.service';

@Component({
  selector: 'app-roles',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './roles.component.html',
  styleUrls: ['./roles.component.css']
})
export class RolesComponent implements OnInit {
  private rolesService = inject(RolesService);
  private toast = inject(ToastService);
  private alertService = inject(AlertService);
  private fb = inject(FormBuilder);

  roles = signal<Role[]>([]);
  permissions = signal<Permission[]>([]);
  modules = signal<{ name: string, permissions: Permission[] }[]>([]);

  selectedRole = signal<Role | null>(null);
  showModal = signal(false);

  roleForm: FormGroup;

  constructor() {
    this.roleForm = this.fb.group({
      name: ['', Validators.required],
      description: [''],
      permission_ids: [[]]
    });
  }

  ngOnInit(): void {
    this.loadRoles();
    this.loadPermissions();
  }

  loadRoles() {
    this.rolesService.getRoles().subscribe({
      next: (res) => this.roles.set(res),
      error: () => this.toast.showError('Failed to load roles')
    });
  }

  loadPermissions() {
    this.rolesService.getPermissions().subscribe({
      next: (res) => {
        this.permissions.set(res);
        this.groupPermissions(res);
      },
      error: () => this.toast.showError('Failed to load permissions')
    });
  }

  groupPermissions(perms: Permission[]) {
    const grouped = perms.reduce((acc, curr) => {
      if (!acc[curr.module]) acc[curr.module] = [];
      acc[curr.module].push(curr);
      return acc;
    }, {} as Record<string, Permission[]>);

    const arr = Object.keys(grouped).map(k => ({
      name: k,
      permissions: grouped[k]
    }));
    this.modules.set(arr);
  }

  openModal(role?: Role) {
    if (role) {
      if (role.is_system) {
        this.toast.showError('System roles cannot be edited');
        return;
      }
      this.selectedRole.set(role);
      this.roleForm.patchValue({
        name: role.name,
        description: role.description,
        permission_ids: role.permissions?.map(p => p.id) || []
      });
    } else {
      this.selectedRole.set(null);
      this.roleForm.reset({ permission_ids: [] });
    }
    this.showModal.set(true);
  }

  closeModal() {
    this.showModal.set(false);
  }

  togglePermission(permId: string) {
    const perms = this.roleForm.get('permission_ids')?.value || [];
    const idx = perms.indexOf(permId);
    if (idx === -1) {
      perms.push(permId);
    } else {
      perms.splice(idx, 1);
    }
    this.roleForm.patchValue({ permission_ids: perms });
  }

  hasPermissionSelected(permId: string): boolean {
    const perms = this.roleForm.get('permission_ids')?.value || [];
    return perms.includes(permId);
  }

  saveRole() {
    if (this.roleForm.invalid) return;

    const role = this.selectedRole();
    const obs$ = role
      ? this.rolesService.updateRole(role.id, this.roleForm.value)
      : this.rolesService.createRole(this.roleForm.value);

    obs$.subscribe({
      next: () => {
        this.toast.showSuccess(`Role ${role ? 'updated' : 'created'} successfully`);
        this.loadRoles();
        this.closeModal();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Operation failed')
    });
  }

  async deleteRole(role: Role) {
    if (role.is_system) {
      this.toast.showError('System roles cannot be deleted');
      return;
    }
    if (await this.alertService.confirm(`Are you sure you want to delete ${role.name}?`, 'Delete Role', 'Delete', 'Cancel', 'danger')) {
      this.rolesService.deleteRole(role.id).subscribe({
        next: () => {
          this.toast.showSuccess('Role deleted');
          this.loadRoles();
        },
        error: (err) => this.toast.showError(err.error?.error || 'Failed to delete role')
      });
    }
  }
}
