const fs = require('fs');

let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/shared/components/receipt/receipt.component.ts', 'utf8');

// Replace width
code = code.replace('w-[80mm]', 'w-[58mm]');

// Import QRCode
if (!code.includes('QRCodeComponent')) {
    code = code.replace("import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';", "import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';\nimport { QRCodeComponent } from 'angularx-qrcode';");
    code = code.replace("imports: [CommonModule, AppCurrencyPipe],", "imports: [CommonModule, AppCurrencyPipe, QRCodeComponent],");
}

// Replace faux barcode with qr code
const barcodeRegex = /<!-- Faux Barcode using flex bars -->[\s\S]*?<\/div>\n      <p class="text-\[10px\] tracking-\[0.2em\] font-black mb-6">/g;

code = code.replace(barcodeRegex, `<qrcode [qrdata]="order?.order_number || '00000000'" [width]="100" [errorCorrectionLevel]="'M'" class="mb-1"></qrcode>
      <p class="text-[10px] tracking-[0.2em] font-black mb-6">`);

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/shared/components/receipt/receipt.component.ts', code);
