import { Injectable, signal } from '@angular/core';

/**
 * ESC/POS command constants for thermal printers.
 */
const ESC = 0x1b;
const GS = 0x1d;

const CMD = {
  INIT:             [ESC, 0x40],
  ALIGN_LEFT:       [ESC, 0x61, 0x00],
  ALIGN_CENTER:     [ESC, 0x61, 0x01],
  ALIGN_RIGHT:      [ESC, 0x61, 0x02],
  BOLD_ON:          [ESC, 0x45, 0x01],
  BOLD_OFF:         [ESC, 0x45, 0x00],
  DOUBLE_HEIGHT:    [ESC, 0x21, 0x10],
  NORMAL_SIZE:      [ESC, 0x21, 0x00],
  FEED_LINE:        [0x0a],
  FEED_LINES:       (n: number) => [ESC, 0x64, n],
  CUT:              [GS, 0x56, 0x01],
};

export interface PrintableReceipt {
  storeName: string;
  branchName?: string;
  address?: string;
  phone?: string;
  orderNumber: string;
  cashier?: string;
  date: Date;
  items: { name: string; qty: number; price: number; total: number }[];
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  amountPaid: number;
  paymentMethod: string;
  change?: number;
  footer?: string;
}

@Injectable({ providedIn: 'root' })
export class PrinterService {
  readonly isSupported = signal('serial' in navigator);
  private port: any = null;
  private writer: WritableStreamDefaultWriter<Uint8Array> | null = null;
  readonly isConnected = signal(false);
  readonly isConnecting = signal(false);
  readonly lastError = signal<string | null>(null);

  async connect(): Promise<boolean> {
    if (!this.isSupported()) {
      this.lastError.set('Web Serial API not supported. Use Chrome or Edge.');
      return false;
    }
    try {
      this.isConnecting.set(true);
      this.lastError.set(null);
      this.port = await (navigator as any).serial.requestPort({
        filters: [
          { usbVendorId: 0x04b8 }, // Epson
          { usbVendorId: 0x0483 }, // Generic
          { usbVendorId: 0x067b }, // Prolific USB-serial
          { usbVendorId: 0x0519 }, // Star Micronics
          { usbVendorId: 0x1fc9 }, // Bixolon
        ],
      });
      await this.port.open({ baudRate: 9600 });
      this.writer = this.port.writable?.getWriter() ?? null;
      this.isConnected.set(true);
      return true;
    } catch (err: any) {
      if (err?.name !== 'NotFoundError') {
        this.lastError.set(err?.message ?? 'Failed to connect to printer.');
      }
      return false;
    } finally {
      this.isConnecting.set(false);
    }
  }

  async disconnect(): Promise<void> {
    try { this.writer?.releaseLock(); await this.port?.close(); } catch {}
    this.port = null;
    this.writer = null;
    this.isConnected.set(false);
  }

  async printReceipt(receipt: PrintableReceipt, fallbackUrl?: string): Promise<void> {
    if (!this.isConnected() || !this.writer) {
      if (fallbackUrl) this._printViaIframe(fallbackUrl);
      return;
    }
    const lines: number[] = [];
    const write = (bytes: number[]) => lines.push(...bytes);
    const text = (str: string) => lines.push(...str.split('').map(c => c.charCodeAt(0)));
    const newline = () => write(CMD.FEED_LINE);
    const divider = (c = '-', n = 42) => { text(c.repeat(n)); newline(); };
    const leftRight = (l: string, r: string, w = 42) => {
      text(l + ' '.repeat(Math.max(1, w - l.length - r.length)) + r); newline();
    };

    write(CMD.INIT);
    write(CMD.ALIGN_CENTER);
    write(CMD.BOLD_ON);
    write(CMD.DOUBLE_HEIGHT);
    text(receipt.storeName.toUpperCase()); newline();
    write(CMD.NORMAL_SIZE);
    write(CMD.BOLD_OFF);
    if (receipt.branchName) { text(receipt.branchName); newline(); }
    if (receipt.address)    { text(receipt.address); newline(); }
    if (receipt.phone)      { text('Tel: ' + receipt.phone); newline(); }
    newline();
    write(CMD.ALIGN_LEFT);
    divider('=');
    write(CMD.BOLD_ON); text('SALES RECEIPT'); newline(); write(CMD.BOLD_OFF);
    divider('-');
    text('Order: #' + receipt.orderNumber); newline();
    text('Date:  ' + receipt.date.toLocaleDateString()); newline();
    text('Time:  ' + receipt.date.toLocaleTimeString()); newline();
    if (receipt.cashier) { text('Cashier: ' + receipt.cashier); newline(); }
    divider('-');

    for (const item of receipt.items) {
      text(item.name.substring(0, 28)); newline();
      leftRight(`  ${item.qty} x ${this._fmt(item.price)}`, this._fmt(item.total));
    }

    divider('-');
    leftRight('Subtotal', this._fmt(receipt.subtotal));
    if (receipt.tax > 0)      leftRight('Tax', this._fmt(receipt.tax));
    if (receipt.discount > 0) leftRight('Discount', '-' + this._fmt(receipt.discount));
    divider('=');
    write(CMD.BOLD_ON); leftRight('TOTAL', this._fmt(receipt.total)); write(CMD.BOLD_OFF);
    divider('-');
    leftRight('Paid (' + receipt.paymentMethod.toUpperCase() + ')', this._fmt(receipt.amountPaid));
    if ((receipt.change ?? 0) > 0) leftRight('Change', this._fmt(receipt.change!));
    divider('=');
    newline();
    write(CMD.ALIGN_CENTER);
    text(receipt.footer ?? 'Thank you for your purchase!'); newline();
    text('Powered by Puxbay'); newline();
    write(CMD.FEED_LINES(4));
    write(CMD.CUT);

    await this.writer.write(new Uint8Array(lines));
  }

  private _fmt(val: number): string {
    return val.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  private _printViaIframe(url: string): void {
    // Opening in a new window ensures the browser doesn't block the print dialog
    // and also ensures we send proper Accept headers for HTML rendering.
    const win = window.open(url, '_blank', 'width=400,height=600');
    if (win) {
      win.onload = () => {
        setTimeout(() => win.print(), 500);
      };
    }
  }
}
