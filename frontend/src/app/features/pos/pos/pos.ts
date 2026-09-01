import { Component, inject, OnInit, OnDestroy, HostListener, signal } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { ImageUrlPipe } from '../../../core/pipes/image-url.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Product } from '../../../core/models/models';
import { PosFacade } from '../pos.facade';
import { RouterModule } from '@angular/router';
import { Html5Qrcode, Html5QrcodeSupportedFormats } from 'html5-qrcode';
import { GiftCardService } from '../../../core/services/gift-card.service';
import { ToastrService } from 'ngx-toastr';
import { SettingsService } from '../../../core/services/settings.service';
import { QRCodeComponent } from 'angularx-qrcode';
import { ReceiptComponent } from '../../../shared/components/receipt/receipt.component';

@Component({
  selector: 'app-pos',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe, ImageUrlPipe, QRCodeComponent, ReceiptComponent],
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
    
    @keyframes pulse-gradient { 0%, 100% { opacity: 1; } 50% { opacity: 0.7; } }
    .btn-checkout-active {
      animation: pulse-gradient 2s infinite;
    }
    @media print {
      body * {
        visibility: hidden;
      }
      #pos-app, .no-print {
        display: none !important;
      }
      app-receipt, app-receipt *, #print-receipt, #print-receipt * {
        visibility: visible !important;
      }
      app-receipt {
        display: block !important;
        position: fixed !important;
        left: 0 !important;
        top: 0 !important;
        width: 100% !important;
        margin: 0 !important;
        padding: 0 !important;
        background: #fff !important;
        z-index: 9999999 !important;
      }
      #print-receipt {
        display: block !important;
        margin: 0 auto !important;
        width: 100% !important;
        max-width: 80mm !important;
      }
    }

    /* Scanner modal styles */
    #pos-camera-scanner-reader {
      width: 100%;
      border-radius: 12px;
      overflow: hidden;
    }
    #pos-camera-scanner-reader video {
      border-radius: 12px;
    }
    /* Override html5-qrcode default button styles */
    #pos-camera-scanner-reader button {
      background: #005b96 !important;
      border-radius: 8px !important;
      border: none !important;
      color: white !important;
      font-weight: bold !important;
      padding: 8px 16px !important;
    }
    #pos-camera-scanner-reader select {
      border-radius: 8px !important;
      padding: 6px 10px !important;
    }
    .scanner-aim-line {
      animation: scanner-scan 2s linear infinite;
    }
    @keyframes scanner-scan {
      0% { top: 10%; opacity: 1; }
      50% { top: 85%; opacity: 0.6; }
      100% { top: 10%; opacity: 1; }
    }
  `,
})
export class Pos implements OnInit, OnDestroy {
  facade = inject(PosFacade);
  giftCardService = inject(GiftCardService);
  toastr = inject(ToastrService);
  settingsService = inject(SettingsService);
  
  Math = Math;
  current_date = new Date();
  isCartOpen = false;

  // Gift Card Prompt State
  isGiftCardPromptOpen = signal(false);
  giftCardCodeInput = signal('');
  isCheckingGiftCard = signal(false);

  // Camera scanner state
  showCameraScanner = signal(false);
  scannerError = signal<string | null>(null);
  scannerSuccess = signal<string | null>(null);
  private html5QrCode: Html5Qrcode | null = null;

  toggleCart() {
    this.isCartOpen = !this.isCartOpen;
  }

  ngOnInit() {
    this.facade.init();
  }

  ngOnDestroy() {
    this.stopCameraScanner();
  }

  // ── Gift Card ───────────────────────────────────────────────
  promptGiftCard() {
    this.isGiftCardPromptOpen.set(true);
    this.giftCardCodeInput.set('');
    this.isCheckingGiftCard.set(false);
  }

  closeGiftCardPrompt() {
    this.isGiftCardPromptOpen.set(false);
    this.giftCardCodeInput.set('');
  }

  applyGiftCard() {
    const code = this.giftCardCodeInput().trim();
    if (!code) return;

    this.isCheckingGiftCard.set(true);
    this.giftCardService.checkBalance(code).subscribe({
      next: (res) => {
        this.isCheckingGiftCard.set(false);
        const balance = res.gift_card.current_balance;
        if (balance <= 0) {
          this.toastr.error('This gift card has no remaining balance.');
          return;
        }

        let amountToUse = this.facade.paymentAmountInput() || this.facade.remainingBalance();
        if (amountToUse > balance) {
          amountToUse = balance; // can only use up to the balance
        }

        this.facade.addPaymentMethod('gift_card', amountToUse, code);
        this.toastr.success(`Applied ${amountToUse} from Gift Card`);
        this.closeGiftCardPrompt();
      },
      error: () => {
        this.isCheckingGiftCard.set(false);
        this.toastr.error('Invalid or expired gift card code.');
      }
    });
  }

  // ── Camera Scanner ──────────────────────────────────────────
  openCameraScanner() {
    this.showCameraScanner.set(true);
    this.scannerError.set(null);
    this.scannerSuccess.set(null);
    // Give DOM time to render the container
    setTimeout(() => this.startScanner(), 150);
  }

  private startScanner() {
    try {
      this.html5QrCode = new Html5Qrcode('pos-camera-scanner-reader', {
        formatsToSupport: [
          Html5QrcodeSupportedFormats.QR_CODE,
          Html5QrcodeSupportedFormats.EAN_13,
          Html5QrcodeSupportedFormats.EAN_8,
          Html5QrcodeSupportedFormats.UPC_A,
          Html5QrcodeSupportedFormats.UPC_E,
          Html5QrcodeSupportedFormats.CODE_128,
          Html5QrcodeSupportedFormats.CODE_39,
          Html5QrcodeSupportedFormats.CODE_93,
          Html5QrcodeSupportedFormats.ITF,
          Html5QrcodeSupportedFormats.DATA_MATRIX,
        ],
        verbose: false,
      });

      this.html5QrCode
        .start(
          { facingMode: 'environment' }, // prefer back camera on mobile
          { fps: 10, qrbox: { width: 260, height: 180 } },
          (decodedText) => this.onScanSuccess(decodedText),
          () => {} // ignore per-frame errors
        )
        .catch((err) => {
          this.scannerError.set('Camera access denied or not available. Please check your browser permissions.');
          console.error('Camera scan error:', err);
        });
    } catch (e) {
      this.scannerError.set('Could not start the camera scanner.');
    }
  }

  private onScanSuccess(decodedText: string) {
    this.scannerSuccess.set(decodedText);
    // Pass to POS barcode handler
    this.facade.scanBarcode(decodedText);
    // Stop camera after successful scan, close modal after brief delay
    this.stopCameraScanner();
    setTimeout(() => {
      this.showCameraScanner.set(false);
      this.scannerSuccess.set(null);
    }, 1200);
  }

  closeCameraScanner() {
    this.stopCameraScanner();
    this.showCameraScanner.set(false);
    this.scannerError.set(null);
    this.scannerSuccess.set(null);
  }

  private stopCameraScanner() {
    if (this.html5QrCode) {
      this.html5QrCode
        .stop()
        .catch(() => {})
        .finally(() => {
          this.html5QrCode = null;
        });
    }
  }

  // ── Keyboard shortcuts ──────────────────────────────────────
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
      if (this.showCameraScanner()) { this.closeCameraScanner(); return; }
      if (this.isGiftCardPromptOpen()) { this.closeGiftCardPrompt(); return; }
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

    // F2 to open camera scanner
    if (event.key === 'F2') {
      event.preventDefault();
      if (!this.showCameraScanner()) this.openCameraScanner();
    }
  }

  @HostListener('window:keypress', ['$event'])
  handleBarcodeScan(event: KeyboardEvent) {
    if (this.facade.isCheckoutModalOpen() || this.showCameraScanner() || event.target instanceof HTMLInputElement) return;

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
