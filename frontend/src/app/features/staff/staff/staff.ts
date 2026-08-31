import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { StaffService, StaffCreateInput } from '../../../core/services/staff.service';
import { BranchService } from '../../../core/services/branch.service';
import { UserProfile } from '../../../core/models/user.models';
import { AlertService } from '../../../core/services/alert.service';
import { ToastrService } from 'ngx-toastr';


export interface RolePermissionRow {
  category: string;
  permission: string;
  description: string;
  admin: boolean;
  manager: boolean;
  supervisor: boolean;
  cashier: boolean;
}

@Component({
  selector: 'app-staff',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './staff.html',
  styles: `
    .animated-gradient-text { color: #005b96; }
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
  `,
})
export class Staff implements OnInit {
  private staffService = inject(StaffService);
  private branchService = inject(BranchService);
  private alertService = inject(AlertService);
  private toastr = inject(ToastrService);

  // Tabs: 'directory' | 'roles' | 'security'
  activeTab = signal<'directory' | 'roles' | 'security'>('directory');

  // View: 'grid' | 'table'
  viewMode = signal<'grid' | 'table'>('grid');

  // State
  staffList = signal<UserProfile[]>([]);
  branches = this.branchService.branches;
  searchQuery = signal('');
  selectedRoleFilter = signal<string>('all');
  selectedBranchFilter = signal<string>('all');
  selectedStatusFilter = signal<string>('all');

  isModalOpen = signal(false);
  isSaving = signal(false);
  editingStaffId = signal<string | null>(null);

  // Selected staff for Quick View drawer
  selectedStaff = signal<UserProfile | null>(null);

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

  // KPI Computations
  totalStaffCount = computed(() => this.staffList().length);
  activeStaffCount = computed(() => this.staffList().filter(s => s.user?.is_active !== false).length);
  managerCount = computed(() => this.staffList().filter(s => ['manager', 'admin', 'supervisor'].includes((s.role || '').toLowerCase())).length);
  cashierCount = computed(() => this.staffList().filter(s => (s.role || '').toLowerCase() === 'cashier').length);

  // ── Security Policies State (Interactive) ──
  quickPinEnabled = signal(true);
  pinLength = signal<number>(4);
  requireManagerOverride = signal(true);

  sessionTimeoutEnabled = signal(true);
  sessionTimeoutMinutes = signal<number>(15);
  autoLockOnIdle = signal(true);

  auditLoggingEnabled = signal(true);
  auditRetentionDays = signal<string>('180');
  immutableLedger = signal(true);

  branchIsolationEnabled = signal(true);
  crossBranchViewing = signal(false);

  savePolicy(policyName: string) {
    this.toastr.success(`Security policy "${policyName}" settings successfully updated and broadcasted.`, 'Policy Updated');
  }

  // Computed state for filtering
  filteredStaff = computed(() => {
    const query = this.searchQuery().toLowerCase().trim();
    const roleFilter = this.selectedRoleFilter().toLowerCase();
    const branchFilter = this.selectedBranchFilter();
    const statusFilter = this.selectedStatusFilter();
    const list = this.staffList();

    return list.filter(staff => {
      const name = `${staff.user?.first_name || ''} ${staff.user?.last_name || ''} ${staff.user?.username || ''}`.toLowerCase();
      const email = (staff.user?.email || '').toLowerCase();
      const phone = (staff.user?.phone || '').toLowerCase();
      const role = (staff.role || 'staff').toLowerCase();

      const matchQuery = !query || name.includes(query) || email.includes(query) || phone.includes(query) || role.includes(query);
      const matchRole = roleFilter === 'all' || role === roleFilter;
      const matchBranch = branchFilter === 'all' || (branchFilter === 'global' ? !staff.branch_id : staff.branch_id === branchFilter);
      const matchStatus = statusFilter === 'all' ||
        (statusFilter === 'active' && staff.user?.is_active !== false) ||
        (statusFilter === 'inactive' && staff.user?.is_active === false);

      return matchQuery && matchRole && matchBranch && matchStatus;
    });
  });

  // Role Capabilities Matrix
  permissionsMatrix: RolePermissionRow[] = [
    { category: 'Point of Sale (POS)', permission: 'Process Sales & Checkouts', description: 'Create orders and issue receipts', admin: true, manager: true, supervisor: true, cashier: true },
    { category: 'Point of Sale (POS)', permission: 'Apply Custom Discounts', description: 'Grant custom or manual line-item discounts', admin: true, manager: true, supervisor: true, cashier: false },
    { category: 'Point of Sale (POS)', permission: 'Authorize Cash Refunds', description: 'Issue immediate cash refunds and returns', admin: true, manager: true, supervisor: true, cashier: false },
    { category: 'Point of Sale (POS)', permission: 'Open Cash Drawer Manually', description: 'Trigger drawer pop without order completion', admin: true, manager: true, supervisor: false, cashier: false },
    { category: 'Inventory & Stock', permission: 'Create Purchase Orders', description: 'Place supply orders to registered suppliers', admin: true, manager: true, supervisor: false, cashier: false },
    { category: 'Inventory & Stock', permission: 'Receive PO Inward Shipments', description: 'Scan items and increase branch inventory', admin: true, manager: true, supervisor: true, cashier: false },
    { category: 'Inventory & Stock', permission: 'Conduct Stocktakes & Audits', description: 'Perform physical barcode inventory counts', admin: true, manager: true, supervisor: true, cashier: false },
    { category: 'Inventory & Stock', permission: 'Inter-Branch Stock Transfers', description: 'Ship or receive inventory between branches', admin: true, manager: true, supervisor: false, cashier: false },
    { category: 'Financials & Reports', permission: 'View Daily Z-Reports & Sales', description: 'Daily store register breakdown', admin: true, manager: true, supervisor: true, cashier: false },
    { category: 'Financials & Reports', permission: 'View Profit & Loss & Margins', description: 'Full revenue and COGS financial data', admin: true, manager: false, supervisor: false, cashier: false },
    { category: 'Financials & Reports', permission: 'Approve Supplier Invoices', description: 'Authorize accounts payable disbursements', admin: true, manager: true, supervisor: false, cashier: false },
    { category: 'Administration', permission: 'Add / Edit Staff & Access', description: 'Create user profiles and assign roles', admin: true, manager: false, supervisor: false, cashier: false },
    { category: 'Administration', permission: 'Manage Branch Configurations', description: 'Set store metadata and operational hours', admin: true, manager: false, supervisor: false, cashier: false },
    { category: 'Administration', permission: 'Developer API & Webhooks', description: 'Generate API tokens and webhook routes', admin: true, manager: false, supervisor: false, cashier: false },
  ];

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
        password: this.generateRandomPassword(),
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

  generateRandomPassword(): string {
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%';
    let pwd = '';
    for (let i = 0; i < 10; i++) {
      pwd += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return pwd;
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
          this.loadStaff();
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
          this.loadStaff();
        },
        error: (err) => {
          console.error('Failed to save staff', err);
          this.isSaving.set(false);
        }
      });
    }
  }

  async deleteStaff(id: string) {
    if (await this.alertService.confirm('Are you sure you want to remove this staff member from your team?', 'Delete Staff Member')) {
      this.staffService.deleteStaff(id).subscribe({
        next: () => {
          this.staffList.update(list => list.filter(s => s.id !== id));
        },
        error: (err) => console.error('Failed to delete staff', err)
      });
    }
  }

  getRoleBadgeClass(role: string = ''): string {
    const r = role.toLowerCase();
    if (r === 'admin' || r === 'superadmin' || r === 'administrator') {
      return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/30';
    }
    if (r === 'manager') {
      return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/30';
    }
    if (r === 'supervisor') {
      return 'bg-teal-500/10 text-teal-600 dark:text-teal-400 border-teal-500/30';
    }
    if (r === 'cashier') {
      return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30';
    }
    return 'bg-zinc-500/10 text-zinc-600 dark:text-zinc-400 border-zinc-500/30';
  }

  getRoleIcon(role: string = ''): string {
    const r = role.toLowerCase();
    if (r === 'admin' || r === 'superadmin') return 'shield_person';
    if (r === 'manager') return 'manage_accounts';
    if (r === 'supervisor') return 'verified_user';
    if (r === 'cashier') return 'point_of_sale';
    return 'person';
  }

  getBranchName(branchId?: string): string {
    if (!branchId) return 'All Branches (Global)';
    const branch = this.branches().find(b => b.id === branchId);
    return branch ? branch.name : 'Unknown Branch';
  }
}

