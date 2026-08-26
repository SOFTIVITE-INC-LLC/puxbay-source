import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SecurityService, AdminRole, AVAILABLE_PERMISSIONS } from '../../services/security.service';

@Component({
  selector: 'app-admin-users',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './admin-users.html',
})
export class AdminUsersComponent implements OnInit {
  private securityService = inject(SecurityService);

  users = signal<any[]>([]);
  roles = signal<AdminRole[]>([]);
  
  availablePermissions = AVAILABLE_PERMISSIONS;

  selectedUser = signal<any | null>(null);
  selectedPermissions = signal<string[]>([]);

  isLoading = signal(true);
  isSaving = signal(false);
  isModalOpen = signal(false);
  
  newUser = signal({
    first_name: '',
    last_name: '',
    email: '',
    username: '',
    password: '',
    admin_role_id: null as string | null,
    is_superuser: false,
    permissions: '[]'
  });

  ngOnInit() {
    this.loadData();
  }

  loadData() {
    this.isLoading.set(true);
    
    // Load users
    this.securityService.getAdminUsers().subscribe({
      next: (res) => {
        this.users.set(res.data || []);
        
        // Also load roles for the dropdown
        this.securityService.getRoles().subscribe({
          next: (roleRes) => {
            this.roles.set(roleRes.data || []);
            
            // Re-select user if one was selected
            const selUser = this.selectedUser();
            if (selUser) {
              const updatedUser = (res.data || []).find((u: any) => u.id === selUser.id);
              if (updatedUser) {
                this.selectUser(updatedUser);
              } else {
                this.selectedUser.set(null);
                this.selectedPermissions.set([]);
              }
            }

            this.isLoading.set(false);
          },
          error: () => this.isLoading.set(false)
        });
      },
      error: () => this.isLoading.set(false)
    });
  }

  selectUser(user: any) {
    this.selectedUser.set(user);
    let perms = [];
    try {
      perms = typeof user.permissions === 'string' ? JSON.parse(user.permissions || '[]') : (user.permissions || []);
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

  selectAll() {
    this.selectedPermissions.set(this.availablePermissions.map(p => p.id));
  }

  clearAll() {
    this.selectedPermissions.set([]);
  }

  savePermissions() {
    const user = this.selectedUser();
    if (!user) return;
    this.isSaving.set(true);
    const permsStr = JSON.stringify(this.selectedPermissions());
    this.securityService.updateAdminUserRole(user.id, {
      admin_role_id: user.admin_role_id,
      is_superuser: user.is_superuser,
      permissions: permsStr
    }).subscribe({
      next: () => {
        this.isSaving.set(false);
        user.permissions = permsStr;
      },
      error: () => this.isSaving.set(false)
    });
  }

  openCreateModal() {
    this.newUser.set({
      first_name: '',
      last_name: '',
      email: '',
      username: '',
      password: '',
      admin_role_id: null,
      is_superuser: false,
      permissions: '[]'
    });
    this.isModalOpen.set(true);
  }

  closeCreateModal() {
    this.isModalOpen.set(false);
  }

  createUser() {
    const user = this.newUser();
    if (!user.first_name || !user.last_name || !user.email || !user.username || !user.password) return;
    
    this.isSaving.set(true);
    this.securityService.createAdminUser(user).subscribe({
      next: () => {
        this.isSaving.set(false);
        this.closeCreateModal();
        this.loadData();
      },
      error: (err) => {
        console.error('Failed to create user', err);
        this.isSaving.set(false);
      }
    });
  }

  toggleSuperuser(user: any, event?: any) {
    if (event) event.stopPropagation();
    const newStatus = !user.is_superuser;
    this.securityService.updateAdminUserRole(user.id, {
      admin_role_id: user.admin_role_id,
      is_superuser: newStatus,
      permissions: user.permissions
    }).subscribe({
      next: () => {
        user.is_superuser = newStatus;
      },
      error: (err) => console.error('Failed to update superuser status', err)
    });
  }

  changeRole(user: any, event: any) {
    event.stopPropagation();
    const roleId = event.target.value === 'null' ? null : event.target.value;
    this.securityService.updateAdminUserRole(user.id, {
      admin_role_id: roleId,
      is_superuser: user.is_superuser,
      permissions: user.permissions
    }).subscribe({
      next: () => {
        user.admin_role_id = roleId;
        const role = this.roles().find(r => r.id === roleId);
        user.admin_role_name = role ? role.name : null;
      },
      error: (err) => {
        console.error('Failed to update role', err);
        event.target.value = user.admin_role_id || 'null';
      }
    });
  }

  getPermCount(user: any): number {
    try {
      const perms = typeof user.permissions === 'string' ? JSON.parse(user.permissions || '[]') : (user.permissions || []);
      return perms.length;
    } catch { return 0; }
  }

  deleteUser(user: any, event: any) {
    event.stopPropagation();
    if (confirm(`Are you sure you want to revoke admin access for ${user.user?.first_name} ${user.user?.last_name}?`)) {
      this.isSaving.set(true);
      this.securityService.deleteAdminUser(user.user_id).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.loadData(); // reload list
        },
        error: (err) => {
          console.error('Failed to delete admin user', err);
          this.isSaving.set(false);
          alert('Failed to revoke admin access.');
        }
      });
    }
  }
}
