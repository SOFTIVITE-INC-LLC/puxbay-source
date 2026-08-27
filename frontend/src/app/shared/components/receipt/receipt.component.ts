import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ViewEncapsulation } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { QRCodeComponent } from 'angularx-qrcode';

@Component({
  selector: 'app-receipt',
  standalone: true,
  imports: [CommonModule, AppCurrencyPipe, QRCodeComponent],
  // ViewEncapsulation.None so @media print styles apply globally (not shadow-scoped)
  encapsulation: ViewEncapsulation.None,
  template: `
  <div id="print-receipt" class="legacy-receipt">
   <div class="header">
     <div class="store-logo" *ngIf="order?.tenant?.logo_url">
       <img [src]="order.tenant.logo_url" alt="Store Logo" style="max-height: 50px; margin-bottom: 10px;">
     </div>
     <div class="store-name">{{ order?.branch?.name || order?.tenant?.name || 'PUXBAY STORE' }}</div>
     <div *ngIf="order?.branch?.address">{{ order?.branch?.address }}</div>
     <div *ngIf="!order?.branch?.address">123 Commerce Avenue, Accra</div>
     <div *ngIf="order?.branch?.contact_email">{{ order?.branch?.contact_email }}</div>
     <div *ngIf="order?.branch?.phone">{{ order?.branch?.phone }}</div>
     <div *ngIf="!order?.branch?.phone">+233 24 613 6978</div>
   </div>

   <div class="meta">
     <div>
       <span>Order #:</span>
       <span>{{ order?.order_number || order?.id || 'N/A' }}</span>
     </div>
     <div>
       <span>Date:</span>
       <span>{{ (order?.created_at || order?.date) | date:'dd/MM/yyyy HH:mm' }}</span>
     </div>
     <div *ngIf="order?.cashier || order?.user">
       <span>Cashier:</span>
       <span>{{ order?.cashier || order?.user?.name || 'Admin' }}</span>
     </div>
   </div>

   <table>
     <thead>
       <tr>
         <th style="width: 45%;">Item</th>
         <th class="text-right">Qty</th>
         <th class="text-right">Price</th>
         <th class="text-right">Total</th>
       </tr>
     </thead>
     <tbody>
       <tr *ngFor="let item of order?.items">
         <td>{{ item.product?.name || item.product_name || item.name || 'Item' }}</td>
         <td class="text-right">{{ item.quantity }}</td>
         <td class="text-right">{{ (item.unit_price || item.price) | appCurrency }}</td>
         <td class="text-right">{{ ((item.unit_price || item.price) * item.quantity) | appCurrency }}</td>
       </tr>
     </tbody>
   </table>

   <div class="totals">
     <div *ngIf="order?.discount_amount > 0" class="total-row subtotal-row">
       <span>Subtotal</span>
       <span>{{ ((order?.total_amount || order?.total) + (order?.discount_amount || 0)) | appCurrency }}</span>
     </div>
     <div *ngIf="order?.discount_amount > 0" class="total-row subtotal-row" style="color:#666;">
       <span>Discount</span>
       <span>-{{ order?.discount_amount | appCurrency }}</span>
     </div>
     <div class="total-row">
       <span>TOTAL</span>
       <span>{{ (order?.total_amount || order?.total) | appCurrency }}</span>
     </div>
     <div *ngIf="order?.amount_paid > 0" class="total-row subtotal-row">
       <span>Cash Paid</span>
       <span>{{ order?.amount_paid | appCurrency }}</span>
     </div>
     <div *ngIf="order?.change_due > 0" class="total-row subtotal-row">
       <span>Change</span>
       <span>{{ order?.change_due | appCurrency }}</span>
     </div>
   </div>

   <div class="footer">
     <p>Thank you for shopping with us!</p>
     <p>Please keep this receipt for returns.</p>
     <div style="margin-top: 15px; padding: 12px; border: 1px solid #eee; border-radius: 8px;">
       <div style="font-weight: bold; margin-bottom: 5px;">Join MyWallet</div>
       <div style="font-size: 10px; color: #666; margin-bottom: 10px;">Track points &amp; receipts on your phone</div>
       <qrcode [qrdata]="order?.order_number || '00000000'" [width]="100" [errorCorrectionLevel]="'M'"></qrcode>
     </div>
   </div>
  </div>
  `,
  styles: [`
    /* ── Default: receipt is hidden in normal page view ── */
    #print-receipt {
      display: none;
    }

    /* ── Receipt Styles (global since encapsulation is None) ── */
    .legacy-receipt {
      font-family: 'Courier New', Courier, monospace;
      font-size: 14px;
      color: #000;
      background: #fff;
      padding: 20px;
      max-width: 350px;
      margin: 0 auto;
    }
    .legacy-receipt .header {
      text-align: center;
      margin-bottom: 20px;
      border-bottom: 1px dashed #000;
      padding-bottom: 10px;
    }
    .legacy-receipt .store-name {
      font-size: 18px;
      font-weight: bold;
      margin-bottom: 5px;
    }
    .legacy-receipt .meta {
      margin-bottom: 15px;
      font-size: 12px;
    }
    .legacy-receipt .meta div {
      display: flex;
      justify-content: space-between;
    }
    .legacy-receipt table {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 15px;
    }
    .legacy-receipt th {
      text-align: left;
      border-bottom: 1px solid #000;
      padding: 5px 0;
    }
    .legacy-receipt td {
      padding: 5px 0;
      vertical-align: top;
    }
    .legacy-receipt .text-right {
      text-align: right;
    }
    .legacy-receipt .totals {
      border-top: 1px dashed #000;
      padding-top: 10px;
      margin-top: 10px;
    }
    .legacy-receipt .total-row {
      display: flex;
      justify-content: space-between;
      font-weight: bold;
      font-size: 16px;
      margin-top: 5px;
    }
    .legacy-receipt .subtotal-row {
      font-size: 13px;
      font-weight: normal;
    }
    .legacy-receipt .footer {
      text-align: center;
      margin-top: 30px;
      font-size: 12px;
      border-top: 1px dashed #000;
      padding-top: 10px;
    }
    .legacy-receipt qrcode {
      display: flex;
      justify-content: center;
    }

    /* ── Print Media: show receipt, hide everything else ── */
    @media print {
      body * {
        visibility: hidden !important;
      }
      #print-receipt {
        display: block !important;
        visibility: visible !important;
        position: fixed !important;
        left: 0 !important;
        top: 0 !important;
        width: 100% !important;
        max-width: 80mm !important;
        margin: 0 auto !important;
        padding: 10px !important;
        background: #fff !important;
        color: #000 !important;
        z-index: 9999999 !important;
      }
      #print-receipt * {
        visibility: visible !important;
        color: #000 !important;
        background: transparent !important;
      }
    }
  `]
})
export class ReceiptComponent {
  @Input() order: any;
}
