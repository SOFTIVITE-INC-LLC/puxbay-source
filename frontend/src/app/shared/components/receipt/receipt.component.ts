import { Component, Input, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { QRCodeComponent } from 'angularx-qrcode';

@Component({
  selector: 'app-receipt',
  standalone: true,
  imports: [CommonModule, AppCurrencyPipe, QRCodeComponent],
  template: `
  <div id="print-receipt" class="legacy-receipt">
   <div class="header">
     <div class="store-logo" *ngIf="order?.tenant?.logo_url">
       <img [src]="order.tenant.logo_url" alt="Store Logo" style="max-height: 50px; margin-bottom: 10px;">
     </div>
     <div class="store-name">{{ order?.branch?.name || order?.tenant?.name || 'THINKCE' }}</div>
     <div *ngIf="order?.branch?.address">{{ order?.branch?.address }}</div>
     <div *ngIf="!order?.branch?.address">123 Commerce Avenue</div>
     <div *ngIf="order?.branch?.contact_email">{{ order?.branch?.contact_email }}</div>
     <div *ngIf="order?.branch?.phone">{{ order?.branch?.phone }}</div>
     <div *ngIf="!order?.branch?.phone">+233 55 123 4567</div>
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
     <div class="total-row">
       <span>TOTAL</span>
       <span>{{ (order?.total_amount || order?.total) | appCurrency }}</span>
     </div>
   </div>

   <div class="footer">
     <p>Thank you for shopping with us!</p>
     <p>Please keep this receipt for returns.</p>
     
     <div style="margin-top: 20px; padding: 15px; border: 1px solid #eee; border-radius: 8px;">
       <div style="font-weight: bold; margin-bottom: 5px;">Join MyWallet</div>
       <div style="font-size: 10px; color: #666; margin-bottom: 10px;">Track points & receipts on your phone</div>
       <qrcode [qrdata]="order?.order_number || '00000000'" [width]="100" [errorCorrectionLevel]="'M'"></qrcode>
     </div>
   </div>
  </div>
  `,
  styles: [`
    #print-receipt {
      display: none;
    }
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
    .legacy-receipt .footer {
      text-align: center;
      margin-top: 30px;
      font-size: 12px;
      border-top: 1px dashed #000;
      padding-top: 10px;
    }
    ::ng-deep .legacy-receipt qrcode {
      display: flex;
      justify-content: center;
    }

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
      .legacy-receipt {
        padding: 10px;
        margin: 0;
      }
      * {
        color: #000 !important;
        background: transparent !important;
      }
    }
  `]
})
export class ReceiptComponent {
  @Input() order: any;
}
