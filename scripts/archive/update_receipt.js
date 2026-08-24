const fs = require('fs');

const code = `import { Component, Input, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-receipt',
  standalone: true,
  imports: [CommonModule, AppCurrencyPipe],
  template: \`
  <div id="print-receipt" class="hidden text-black bg-white w-[80mm] p-6 text-sm font-mono mx-auto shadow-none print:shadow-none">
    
    <!-- Logo & Header -->
    <div class="flex flex-col items-center mb-4">
      <div class="w-12 h-12 bg-black text-white rounded-xl flex items-center justify-center mb-2 print:border print:border-black print:bg-white print:text-black">
        <span class="material-symbols-outlined text-2xl font-black">point_of_sale</span>
      </div>
      <h1 class="text-center font-black text-2xl uppercase tracking-widest leading-none">THINKCE</h1>
      <p class="text-center text-[10px] mt-1 font-bold">123 Commerce Avenue</p>
      <p class="text-center text-[10px] font-bold">Accra, Ghana</p>
      <p class="text-center text-[10px] font-bold">+233 55 123 4567</p>
    </div>
    
    <!-- Order Metadata -->
    <div class="border-t-2 border-b-2 border-black border-dotted py-2 mb-4 text-xs font-bold space-y-1">
      <div class="flex justify-between">
        <span>Order #</span>
        <span>{{ order?.order_number || 'N/A' }}</span>
      </div>
      <div class="flex justify-between">
        <span>Date</span>
        <span>{{ (order?.created_at || order?.date) | date:'medium' }}</span>
      </div>
      <div class="flex justify-between" *ngIf="order?.cashier || order?.user">
        <span>Cashier</span>
        <span>{{ order?.cashier || order?.user?.name || 'Admin' }}</span>
      </div>
      <div class="flex justify-between" *ngIf="order?.customer">
        <span>Customer</span>
        <span>{{ order?.customer?.name || 'Walk-in' }}</span>
      </div>
    </div>
    
    <!-- Items Table -->
    <div class="mb-4">
      <div class="flex justify-between font-black border-b border-black pb-1 mb-2 text-[11px] uppercase tracking-wider">
        <span>Item</span>
        <span>Amount</span>
      </div>
      <div *ngFor="let item of order?.items" class="mb-2">
        <div class="flex justify-between text-xs font-bold">
          <span class="truncate pr-2">{{ item.product_name || item.name || 'Item' }}</span>
          <span>{{ ((item.unit_price || item.price) * item.quantity) | appCurrency }}</span>
        </div>
        <div class="text-[10px] text-gray-600 print:text-black font-semibold">
          {{ item.quantity }} x {{ (item.unit_price || item.price) | appCurrency }}
        </div>
      </div>
    </div>
    
    <!-- Totals -->
    <div class="border-t-2 border-black border-dotted pt-2 mb-6 text-xs font-bold space-y-1">
      <div class="flex justify-between">
        <span>Subtotal</span>
        <span>{{ (order?.subtotal || order?.total) | appCurrency }}</span>
      </div>
      <div *ngIf="order?.discount > 0" class="flex justify-between">
        <span>Discount</span>
        <span>-{{ order?.discount | appCurrency }}</span>
      </div>
      <div *ngIf="order?.tax > 0" class="flex justify-between">
        <span>Tax</span>
        <span>{{ order?.tax | appCurrency }}</span>
      </div>
      
      <div class="flex justify-between font-black text-xl py-2 mt-2 border-y-2 border-black">
        <span>TOTAL</span>
        <span>{{ (order?.total_amount || order?.total) | appCurrency }}</span>
      </div>
      
      <div class="flex justify-between pt-2">
        <span>Payment Method</span>
        <span class="uppercase">{{ order?.payment_method || 'CASH' }}</span>
      </div>
      <div class="flex justify-between" *ngIf="order?.amount_paid">
        <span>Amount Paid</span>
        <span>{{ order?.amount_paid | appCurrency }}</span>
      </div>
      <div class="flex justify-between" *ngIf="order?.change > 0">
        <span>Change</span>
        <span>{{ order?.change | appCurrency }}</span>
      </div>
    </div>
    
    <!-- Barcode & Footer -->
    <div class="flex flex-col items-center mt-8">
      <!-- Faux Barcode using flex bars -->
      <div class="flex h-12 w-full justify-center items-end gap-[2px] mb-1">
        <div class="w-1 bg-black h-full"></div>
        <div class="w-2 bg-black h-[90%]"></div>
        <div class="w-[2px] bg-black h-full"></div>
        <div class="w-1 bg-black h-[80%]"></div>
        <div class="w-[3px] bg-black h-[95%]"></div>
        <div class="w-2 bg-black h-[100%]"></div>
        <div class="w-1 bg-black h-full"></div>
        <div class="w-[2px] bg-black h-[85%]"></div>
        <div class="w-1 bg-black h-full"></div>
        <div class="w-2 bg-black h-[90%]"></div>
        <div class="w-[2px] bg-black h-full"></div>
        <div class="w-[3px] bg-black h-full"></div>
        <div class="w-1 bg-black h-[80%]"></div>
        <div class="w-2 bg-black h-full"></div>
        <div class="w-1 bg-black h-[95%]"></div>
      </div>
      <p class="text-[10px] tracking-[0.2em] font-black mb-6">{{ order?.order_number || '00000000' }}</p>
      
      <h2 class="text-center font-black text-lg">Thank You!</h2>
      <p class="text-center text-xs font-bold mt-1">Please come again</p>
    </div>
    
  </div>
  \`,
  styles: [\`
    @media print {
      :host {
        display: block !important;
      }
      ::ng-deep body * {
        visibility: hidden;
      }
      ::ng-deep app-receipt, ::ng-deep app-receipt * {
        visibility: visible;
      }
      ::ng-deep app-receipt {
        position: absolute;
        left: 0;
        top: 0;
        width: 100%;
        margin: 0;
        padding: 0;
      }
      #print-receipt {
        display: block !important;
        margin: 0 !important;
        width: 100% !important;
        padding: 0 !important;
      }
      
      /* Black and white optimizations for thermal printers */
      * {
        color: #000 !important;
        background: transparent !important;
      }
      
      .bg-black {
        background-color: #000 !important;
        -webkit-print-color-adjust: exact;
        print-color-adjust: exact;
      }
      
      .text-white {
        color: #fff !important;
        -webkit-print-color-adjust: exact;
        print-color-adjust: exact;
      }
    }
  \`]
})
export class ReceiptComponent {
  @Input() order: any;
}
`;

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/shared/components/receipt/receipt.component.ts', code);
