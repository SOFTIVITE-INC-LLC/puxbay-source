import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SupplierPortalService, SupplierASN, PurchaseOrder } from '../../services/supplier-portal.service';
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
  loading = signal<boolean>(false);

  showCreateModal = signal<boolean>(false);
  selectedPOId = '';
  carrier = 'DHL Express';
  trackingNumber = '';
  dispatchDate = new Date().toISOString().split('T')[0];
  expectedArrival = '';
  packageCount = 1;
  weightKg = 5;
  notes = '';

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
    if (!this.carrier) {
      this.toast.showError('Carrier is required');
      return;
    }

    const payload: Partial<SupplierASN> = {
      purchase_order_id: this.selectedPOId ? this.selectedPOId : undefined,
      carrier: this.carrier,
      tracking_number: this.trackingNumber,
      dispatch_date: new Date(this.dispatchDate).toISOString(),
      expected_arrival: this.expectedArrival ? new Date(this.expectedArrival).toISOString() : undefined,
      package_count: Number(this.packageCount) || 1,
      total_weight_kg: Number(this.weightKg) || 0,
      notes: this.notes,
      status: 'dispatched'
    };

    this.portalService.createShipment(payload).subscribe({
      next: () => {
        this.toast.showSuccess('Advanced Shipping Notice (ASN) created!');
        this.showCreateModal.set(false);
        this.loadData();
      },
      error: (err) => this.toast.showError(err.error?.error || 'Failed to create shipment')
    });
  }

  statusClass(status: string = ''): string {
    const s = status.toLowerCase();
    if (s === 'delivered') return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
    if (s === 'in_transit' || s === 'dispatched') return 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20';
    if (s === 'rejected') return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
    return 'bg-zinc-800 text-zinc-300 border-zinc-700';
  }
}
