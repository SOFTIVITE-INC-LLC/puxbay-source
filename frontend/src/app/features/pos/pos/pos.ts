import { Component, inject, OnInit, HostListener } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Product } from '../../../core/models/models';
import { PosFacade } from '../pos.facade';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-pos',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe],
  templateUrl: './pos.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.95);
      border: 1px solid rgba(0, 0, 0, 0.06);
      box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
    }
    :root.dark .glass-panel {
      background: #1a1f2e;
      border: 1px solid rgba(255, 255, 255, 0.08);
      box-shadow: 0 4px 24px rgba(0, 0, 0, 0.4);
    }
    
    @keyframes pulse-gradient {
      0% { box-shadow: 0 0 0 0 rgba(99, 102, 241, 0.4); }
      70% { box-shadow: 0 0 0 15px rgba(99, 102, 241, 0); }
      100% { box-shadow: 0 0 0 0 rgba(99, 102, 241, 0); }
    }
    .btn-checkout-active {
      animation: pulse-gradient 2s infinite;
    }
    @media print {
      #pos-app {
        display: none !important;
      }
      #print-receipt {
        display: block !important;
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
      }
    }
  `,
})
export class Pos implements OnInit {
  facade = inject(PosFacade);
  Math = Math;
  current_date = new Date(); // expose for template use in shift variance calculation
  isCartOpen = false;

  toggleCart() {
    this.isCartOpen = !this.isCartOpen;
  }

  ngOnInit() {
    this.facade.init();
  }

  private barcodeBuffer = '';
  private barcodeTimeout: any = null;

  @HostListener('window:keydown', ['$event'])
  handleKeyboardEvent(event: KeyboardEvent) {
    // Ctrl + Enter to process payment
    if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
      event.preventDefault();
      if (this.facade.isCheckoutModalOpen()) {
        this.facade.processCheckout();
      } else {
        this.facade.openCheckout();
      }
    }
    
    if (event.key === 'Escape') {
      if (this.facade.isCheckoutModalOpen()) this.facade.closeCheckout();
      else if (this.facade.isParkedSalesModalOpen()) this.facade.isParkedSalesModalOpen.set(false);
      else {
         this.facade.searchQuery.set('');
         this.facade.selectedCategoryId.set('all');
      }
    }
    
    if (event.key === 'F9') {
      event.preventDefault();
      this.facade.openCheckout();
    }
    
    if (event.key === 'F4') {
      event.preventDefault();
      const searchInput = document.querySelector('input[placeholder*="Search"]') as HTMLInputElement;
      if (searchInput) searchInput.focus();
    }
  }

  @HostListener('window:keypress', ['$event'])
  handleBarcodeScan(event: KeyboardEvent) {
    if (this.facade.isCheckoutModalOpen() || event.target instanceof HTMLInputElement) return;

    if (event.key === 'Enter') {
      if (this.barcodeBuffer.length > 2) {
        this.facade.scanBarcode(this.barcodeBuffer);
      }
      this.barcodeBuffer = '';
      return;
    }
    
    if (event.key.length === 1) {
      this.barcodeBuffer += event.key;
      if (this.barcodeTimeout) clearTimeout(this.barcodeTimeout);
      this.barcodeTimeout = setTimeout(() => {
        this.barcodeBuffer = '';
      }, 50); // fast timeout specifically for barcode scanners
    }
  }
}
