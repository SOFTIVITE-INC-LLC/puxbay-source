import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DeliveryService } from '../../../core/services/delivery.service';
import { ToastService } from '../../../core/services/toast';
import { OrderService } from '../../../core/services/order.service';

@Component({
  selector: 'app-delivery-dashboard',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './delivery-dashboard.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.6);
      backdrop-filter: blur(16px);
      border: 1px solid rgba(255, 255, 255, 0.3);
    }
    .dark .glass-panel {
      background: rgba(24, 24, 27, 0.6);
      border: 1px solid rgba(255, 255, 255, 0.05);
    }
  `
})
export class DeliveryDashboard implements OnInit {
  deliveryService = inject(DeliveryService);
  toast = inject(ToastService);

  isDriverModalOpen = signal(false);
  
  newDriver = signal({
    name: '',
    phone: '',
    vehicle_info: ''
  });

  // Fetch real orders from OrderService
  pendingOrders = signal<any[]>([]);

  // Active deliveries
  activeDeliveries = signal<any[]>([]);

  // Derived Stats
  totalDrivers = computed(() => this.deliveryService.drivers().length);
  availableDrivers = computed(() => this.deliveryService.drivers().filter(d => d.current_status === 'available' || !d.current_status).length);
  activeDispatchCount = computed(() => this.activeDeliveries().length);
  
  dispatchModalOpen = signal(false);
  activeOrder = signal<any>(null);
  dispatchData = signal({
    driver_id: '',
    delivery_fee: 5.00,
    delivery_notes: ''
  });

  orderService = inject(OrderService);

  ngOnInit() {
    this.deliveryService.getDrivers().subscribe();
    this.loadPendingOrders();
    this.loadActiveDeliveries();
  }

  loadActiveDeliveries() {
    this.deliveryService.getActiveDeliveries().subscribe(res => {
      this.activeDeliveries.set(res || []);
    });
  }

  loadPendingOrders() {
    // Fetch orders that are not yet dispatched (e.g. status 'completed' from POS)
    this.orderService.getOrders({ limit: 50 }).subscribe(res => {
      // Map to the format expected by the dashboard
      const mappedOrders = (res.data || []).map(o => ({
        id: o.id,
        order_number: o.order_number,
        customer_name: o.customer?.name || 'Walk-in Customer',
        address: o.customer?.address || 'No Address Provided',
        total: o.total
      }));
      this.pendingOrders.set(mappedOrders);
    });
  }

  openDriverModal() {
    this.newDriver.set({ name: '', phone: '', vehicle_info: '' });
    this.isDriverModalOpen.set(true);
  }

  saveDriver() {
    if (!this.newDriver().name || !this.newDriver().phone) return;
    this.deliveryService.addDriver(this.newDriver()).subscribe({
      next: () => {
        this.toast.showSuccess('Driver added successfully');
        this.isDriverModalOpen.set(false);
      },
      error: () => this.toast.showError('Failed to add driver')
    });
  }

  openDispatchModal(order: any) {
    this.activeOrder.set(order);
    this.dispatchData.set({
      driver_id: '',
      delivery_fee: 5.00,
      delivery_notes: ''
    });
    this.dispatchModalOpen.set(true);
  }

  dispatchOrder() {
    if (!this.dispatchData().driver_id) {
      this.toast.showError('Please select a driver');
      return;
    }
    
    this.deliveryService.dispatchOrder({
      order_id: this.activeOrder().id,
      driver_id: this.dispatchData().driver_id,
      delivery_fee: this.dispatchData().delivery_fee,
      delivery_notes: this.dispatchData().delivery_notes
    }).subscribe({
      next: () => {
        this.toast.showSuccess('Order dispatched successfully');
        this.pendingOrders.update(orders => orders.filter(o => o.id !== this.activeOrder().id));
        this.dispatchModalOpen.set(false);
        this.loadActiveDeliveries(); // Refresh active deliveries
        this.deliveryService.getDrivers().subscribe(); // Refresh driver status
      },
      error: () => this.toast.showError('Failed to dispatch order')
    });
  }

  completeDelivery(id: string) {
    this.deliveryService.completeDelivery(id).subscribe({
      next: () => {
        this.toast.showSuccess('Delivery marked as completed');
        this.loadActiveDeliveries();
        this.deliveryService.getDrivers().subscribe(); // Refresh driver status
      },
      error: () => this.toast.showError('Failed to complete delivery')
    });
  }
}
