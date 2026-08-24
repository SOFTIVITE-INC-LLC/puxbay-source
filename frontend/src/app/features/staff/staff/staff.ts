import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { StaffService, StaffCreateInput } from '../../../core/services/staff.service';
import { BranchService } from '../../../core/services/branch.service';
import { UserProfile } from '../../../core/models/user.models';
import { AlertService } from '../../../core/services/alert.service';

@Component({
  selector: 'app-staff',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './staff.html',
  styles: `
    @keyframes gradientShift {
      0% { background-position: 0% 50%; }
      50% { background-position: 100% 50%; }
      100% { background-position: 0% 50%; }
    }
    .animated-gradient-text {
      background-size: 300% 300%;
      animation: gradientShift 6s ease infinite;
    }
    .glass-panel {
      background: rgba(255, 255, 255, 0.7);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border: 1px solid rgba(255, 255, 255, 0.4);
    }
    :host-context(.dark) .glass-panel {
      background: rgba(24, 24, 27, 0.7);
      border: 1px solid rgba(255, 255, 255, 0.08);
    }
    .card-hover {
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }
    .card-hover:hover {
      transform: translateY(-4px) scale(1.01);
      box-shadow: 0 20px 40px -10px rgba(168, 85, 247, 0.15); /* Purple shadow for Command Center */
    }
    :host-context(.dark) .card-hover:hover {
      box-shadow: 0 20px 40px -10px rgba(0, 0, 0, 0.5);
    }
  `,
})
export class Staff implements OnInit {
  private staffService = inject(StaffService);
  private branchService = inject(BranchService);
  private alertService = inject(AlertService);

  // State
  staffList = signal<UserProfile[]>([]);
  branches = this.branchService.branches;
  searchQuery = signal('');
  isModalOpen = signal(false);
  isSaving = signal(false);
  editingStaffId = signal<string | null>(null);
  
  // Current Staff (for Create/Edit)
  currentStaff = signal<Partial<StaffCreateInput>>({
    username: '',
    email: '',
    phone: '',
    first_name: '',
    last_name: '',
    role: 'cashier',
    password: '',
    branch_id: ''
  });

  // Computed state for filtering
  filteredStaff = computed(() => {
    const query = this.searchQuery().toLowerCase();
    const list = this.staffList();
    if (!query) return list;
    
    return list.filter(staff => 
      (staff.user?.first_name || '').toLowerCase().includes(query) ||
      (staff.user?.last_name || '').toLowerCase().includes(query) ||
      (staff.user?.email || '').toLowerCase().includes(query) ||
      (staff.role || '').toLowerCase().includes(query)
    );
  });

  ngOnInit() {
    this.loadStaff();
    this.branchService.getBranches().subscribe();
  }

  loadStaff() {
    this.staffService.listStaff().subscribe({
      next: (data) => {
        this.staffList.set(data || []);
      },
      error: (err) => console.error('Failed to load staff', err)
    });
  }

  openStaffModal(staff?: UserProfile) {
    if (staff && staff.user) {
      this.editingStaffId.set(staff.id);
      this.currentStaff.set({
        username: staff.user.username || '',
        email: staff.user.email || '',
        phone: staff.user.phone || '',
        first_name: staff.user.first_name || '',
        last_name: staff.user.last_name || '',
        role: staff.role || 'cashier',
        branch_id: staff.branch_id || ''
      });
    } else {
      this.editingStaffId.set(null);
      this.currentStaff.set({
        username: '',
        email: '',
        phone: '',
        first_name: '',
        last_name: '',
        role: 'cashier',
        password: '',
        branch_id: ''
      });
    }
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
    this.currentStaff.set({});
    this.editingStaffId.set(null);
  }

  saveStaff() {
    this.isSaving.set(true);
    const payload = { ...this.currentStaff() } as StaffCreateInput;
    if (!payload.branch_id) {
      delete payload.branch_id;
    }
    
    const id = this.editingStaffId();
    if (id) {
      this.staffService.updateStaff(id, payload).subscribe({
        next: (updatedStaff) => {
          this.staffList.update(list => list.map(s => s.id === id ? updatedStaff : s));
          this.isSaving.set(false);
          this.closeModal();
        },
        error: (err) => {
          console.error('Failed to update staff', err);
          this.isSaving.set(false);
        }
      });
    } else {
      this.staffService.createStaff(payload).subscribe({
        next: (newStaff) => {
          this.staffList.update(list => [...list, newStaff]);
          this.isSaving.set(false);
          this.closeModal();
        },
        error: (err) => {
          console.error('Failed to save staff', err);
          this.isSaving.set(false);
        }
      });
    }
  }

  async deleteStaff(id: string) {
    if (await this.alertService.confirm('Are you sure you want to delete this staff member?', 'Delete Staff')) {
      this.staffService.deleteStaff(id).subscribe({
        next: () => {
          this.staffList.update(list => list.filter(s => s.id !== id));
        },
        error: (err) => console.error('Failed to delete staff', err)
      });
    }
  }

  getRoleColor(role: string): string {
    switch((role || '').toLowerCase()) {
      case 'admin':
      case 'superadmin':
        return 'bg-purple-500/10 text-purple-600 border-purple-500/20';
      case 'manager':
        return 'bg-blue-500/10 text-blue-600 border-blue-500/20';
      case 'cashier':
        return 'bg-emerald-500/10 text-emerald-600 border-emerald-500/20';
      default:
        return 'bg-zinc-500/10 text-zinc-600 border-zinc-500/20';
    }
  }

  getBranchName(branchId?: string): string {
    if (!branchId) return 'All Branches';
    const branch = this.branches().find(b => b.id === branchId);
    return branch ? branch.name : 'Unknown Branch';
  }
}
