import { Component, Input, inject, computed, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ViewEncapsulation } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { QRCodeComponent } from 'angularx-qrcode';
import { StorefrontSettingsService } from '../../../core/store/services/storefront-settings.service';
import { SettingsService } from '../../../core/services/settings.service';
import { ImageUrlPipe } from '../../../core/pipes/image-url.pipe';

@Component({
  selector: 'app-receipt',
  standalone: true,
  imports: [CommonModule, AppCurrencyPipe, QRCodeComponent],
  // ViewEncapsulation.None so @media print styles apply globally (not shadow-scoped)
  encapsulation: ViewEncapsulation.None,
  template: `
  <div id="print-receipt" class="legacy-receipt">
   <div class="header">
     <div class="store-logo" *ngIf="logoUrl()">
       <img [src]="logoUrl()" alt="Store Logo" style="max-height: 48px; max-width: 140px; object-fit: contain; margin: 0 auto 8px auto; display: block;">
     </div>
     <div class="store-name">{{ storeName() }}</div>
     <div *ngIf="order?.branch?.address">{{ order?.branch?.address }}</div>
     <div *ngIf="order?.branch?.contact_email">{{ order?.branch?.contact_email }}</div>
     <div *ngIf="order?.branch?.phone">{{ order?.branch?.phone }}</div>
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
       <span>{{ order?.cashier || order?.user?.name || 'Staff' }}</span>
     </div>
     <div *ngIf="order?.customer_name || order?.customer?.name">
       <span>Customer:</span>
       <span>{{ order?.customer_name || order?.customer?.name }}</span>
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
     <div class="subtotal-row flex justify-between" *ngIf="order?.subtotal">
       <span>Subtotal:</span>
       <span>{{ (order?.subtotal || 0) | appCurrency }}</span>
     </div>
     <div class="subtotal-row flex justify-between" *ngIf="order?.tax">
       <span>Tax:</span>
       <span>{{ (order?.tax || 0) | appCurrency }}</span>
     </div>
     <div class="subtotal-row flex justify-between" *ngIf="order?.discount">
       <span>Discount:</span>
       <span>-{{ (order?.discount || 0) | appCurrency }}</span>
     </div>
     <div class="total-row flex justify-between">
       <span>TOTAL:</span>
       <span>{{ (order?.total || 0) | appCurrency }}</span>
     </div>
     <div class="subtotal-row flex justify-between" *ngIf="order?.amount_paid !== undefined">
       <span>Amount Paid ({{ (order?.payment_method || 'CASH') | uppercase }}):</span>
       <span>{{ (order?.amount_paid || 0) | appCurrency }}</span>
     </div>
   </div>

   <div class="footer">
     <p style="font-weight: bold; margin-bottom: 4px;">Thank you for your purchase!</p>
     <div *ngIf="order?.receipt_token" style="margin: 10px 0;">
       <qrcode [qrdata]="'https://puxbay.com/r/' + order?.receipt_token" [width]="80" [errorCorrectionLevel]="'M'"></qrcode>
       <span style="font-size: 9px; display: block; margin-top: 2px;">Scan to view digital receipt</span>
     </div>
     <div class="powered-by" style="font-size: 10px; color: #444; margin-top: 10px; font-weight: bold; letter-spacing: 0.5px; border-top: 1px dotted #ccc; padding-top: 6px;">
       Powered by PUXBAY
     </div>
   </div>
  </div>
  `,
  styles: [`
    @media screen {
      #print-receipt {
        display: none;
      }
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
      margin-top: 20px;
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
export class ReceiptComponent implements OnInit {
  @Input() order: any;

  storefrontSettings = inject(StorefrontSettingsService);
  settingsService = inject(SettingsService);
  private imageUrlPipe = inject(ImageUrlPipe);

  ngOnInit() {
    this.storefrontSettings.loadSettings().subscribe();
    this.settingsService.getSettings().subscribe();
  }

  logoUrl = computed(() => {
    const raw = this.order?.tenant?.logo_url
      || this.settingsService.settings()?.logo_url
      || this.storefrontSettings.settings()?.logo_image
      || null;
    return raw ? this.imageUrlPipe.transform(raw, false) : null;
  });

  storeName = computed(() => {
    return this.order?.tenant?.name || this.settingsService.settings()?.company_name || this.storefrontSettings.settings()?.store_name || this.order?.branch?.name || 'PUXBAY STORE';
  });
}
