import { Component, OnInit, inject, signal } from '@angular/core';
import { ToastrService } from 'ngx-toastr';
import { DeviceService, Device, DeviceCreateInput } from '../../../core/services/device.service';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { BranchService } from '../../../core/services/branch.service';
import { firstValueFrom } from 'rxjs';

@Component({
  selector: 'app-terminal',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './terminal.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .dark .glass-panel {
      background: rgba(0, 0, 0, 0.2);
    }
  `,
})
export class Terminal implements OnInit {
  private toastr = inject(ToastrService);
  private deviceService = inject(DeviceService);
  private branchService = inject(BranchService);

  devices = signal<Device[]>([]);
  loading = signal<boolean>(false);
  
  showAddModal = signal<boolean>(false);
  saving = signal<boolean>(false);
  
  newDevice = signal<DeviceCreateInput>({
    name: '',
    device_type: 'printer',
    ip_address: '',
    config: ''
  });

  ngOnInit() {
    this.loadDevices();
  }

  async loadDevices() {
    try {
      this.loading.set(true);
      const branchId = this.branchService.activeBranch()?.id;
      const res = await firstValueFrom(this.deviceService.listDevices(branchId));
      this.devices.set(res || []);
    } catch (e: any) {
      this.toastr.error('Failed to load devices');
    } finally {
      this.loading.set(false);
    }
  }

  openAddModal() {
    this.newDevice.set({
      name: '',
      device_type: 'printer',
      ip_address: '',
      config: ''
    });
    this.showAddModal.set(true);
  }

  closeAddModal() {
    this.showAddModal.set(false);
  }

  async saveDevice() {
    try {
      this.saving.set(true);
      const input = this.newDevice();
      const branchId = this.branchService.activeBranch()?.id;
      if (branchId) {
        input.branch_id = branchId;
      }
      await firstValueFrom(this.deviceService.createDevice(input));
      this.toastr.success('Device added successfully');
      this.closeAddModal();
      this.loadDevices();
    } catch (e: any) {
      this.toastr.error(e.error?.error || 'Failed to add device');
    } finally {
      this.saving.set(false);
    }
  }

  async deleteDevice(id: string) {
    if (confirm('Are you sure you want to delete this device?')) {
      try {
        await firstValueFrom(this.deviceService.deleteDevice(id));
        this.toastr.success('Device deleted');
        this.loadDevices();
      } catch (e: any) {
        this.toastr.error('Failed to delete device');
      }
    }
  }

  getIcon(type: string): string {
    switch (type) {
      case 'printer': return 'print';
      case 'payment_terminal': return 'point_of_sale';
      case 'cash_drawer': return 'payments';
      default: return 'devices';
    }
  }
}
