import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SupplierPortalService, SupplierASN, PurchaseOrder, SupplierDeliverySlot } from '../../services/supplier-portal.service';
import { ToastService } from '../../../../core/services/toast';

@Component({
  selector: 'app-supplier-portal-shipments',
  standalone: true,
  imports: [CommonModule, DatePipe, FormsModule],
  templateUrl: './shipments.component.html'
})
export class SupplierPortalShipmentsComponent implements OnInit {
  portalService = inject(SupplierPortalService);
  private toast = inject(ToastService);

  shipments = signal<SupplierASN[]>([]);
  purchaseOrders = signal<PurchaseOrder[]>([]);
  dockSlots = signal<SupplierDeliverySlot[]>([]);
  activeTab = signal<'shipments' | 'dock_slots'>('shipments');
  loading = signal<boolean>(false);

  // ASN creation state
  showCreateModal = signal<boolean>(false);
  selectedPOId = '';
  carrier = 'DHL Express';
  trackingNumber = '';
  dispatchDate = new Date().toISOString().split('T')[0];
  expectedArrival = '';
  packageCount = 1;
  weightKg = 5;
  notes = '';

  // Dock slot booking state
  showDockModal = signal<boolean>(false);
  dockDate = new Date().toISOString().split('T')[0];
  dockWindow = '08:00 - 10:00';
  dockNumber = 'Dock 1';
  vehiclePlate = '';
  driverPhone = '';

  ngOnInit() {
    this.loadData();
  }

  loadData() {
    this.loading.set(true);
    this.portalService.getShipments().subscribe({
      next: (res) => {
        this.shipments.set(res || []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });

    this.portalService.getPurchaseOrders().subscribe({
      next: (orders) => this.purchaseOrders.set(orders || [])
    });

    this.portalService.getDockSlots().subscribe({
      next: (slots) => this.dockSlots.set(slots || [])
    });
  }

  openCreateModal() {
    this.carrier = 'DHL Express';
    this.trackingNumber = '';
    this.dispatchDate = new Date().toISOString().split('T')[0];
    this.expectedArrival = '';
    this.packageCount = 1;
    this.weightKg = 5;
    this.notes = '';
    this.showCreateModal.set(true);
  }

  createASN() {
    const payload: Partial<SupplierASN> = {
      purchase_order_id: this.selectedPOId || undefined,
      carrier: this.carrier,
      tracking_number: this.trackingNumber,
      dispatch_date: new Date(this.dispatchDate).toISOString(),
      expected_arrival: this.expectedArrival ? new Date(this.expectedArrival).toISOString() : undefined,
      package_count: Number(this.packageCount),
      total_weight_kg: Number(this.weightKg),
      status: 'dispatched',
      notes: this.notes
    };

    this.portalService.createShipment(payload).subscribe({
      next: () => {
        this.toast.showSuccess('Advanced Shipping Notice (ASN) created successfully!');
        this.showCreateModal.set(false);
        this.loadData();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to create shipment notice')
    });
  }

  openDockModal() {
    this.dockDate = new Date().toISOString().split('T')[0];
    this.dockWindow = '08:00 - 10:00';
    this.dockNumber = 'Dock 1';
    this.vehiclePlate = '';
    this.driverPhone = '';
    this.showDockModal.set(true);
  }

  bookDockSlot() {
    this.portalService.bookDockSlot({
      slot_date: new Date(this.dockDate).toISOString(),
      time_window: this.dockWindow,
      dock_number: this.dockNumber,
      vehicle_plate: this.vehiclePlate || undefined,
      driver_phone: this.driverPhone || undefined
    }).subscribe({
      next: () => {
        this.toast.showSuccess('Dock delivery window booked successfully!');
        this.showDockModal.set(false);
        this.activeTab.set('dock_slots');
        this.loadData();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to book dock slot')
    });
  }

  statusClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s === 'delivered' || s === 'completed') return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30';
    if (s === 'in_transit' || s === 'dispatched' || s === 'scheduled') return 'bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 border-indigo-500/30';
    if (s === 'rejected' || s === 'cancelled') return 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30';
    return 'bg-slate-100 dark:bg-zinc-800 text-slate-700 dark:text-zinc-300 border-slate-200 dark:border-zinc-700';
  }
}
