const fs = require('fs');

let tsCode = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/orders/orders/orders.ts', 'utf8');

// replace printReceipt
const printRegex = /printReceipt\(\) \{[\s\S]*?\}\);[\s\S]*?\}/;
tsCode = tsCode.replace(printRegex, `printReceipt() {
    window.print();
  }`);

if (!tsCode.includes('ReceiptComponent')) {
    tsCode = tsCode.replace("import { Order } from '../../../core/models/order.models';", "import { Order } from '../../../core/models/order.models';\nimport { ReceiptComponent } from '../../../shared/components/receipt/receipt.component';");
    tsCode = tsCode.replace("imports: [CommonModule, FormsModule, AppCurrencyPipe],", "imports: [CommonModule, FormsModule, AppCurrencyPipe, ReceiptComponent],");
}
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/orders/orders/orders.ts', tsCode);


let htmlCode = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/orders/orders/orders.html', 'utf8');
if (!htmlCode.includes('<app-receipt')) {
    htmlCode = htmlCode + `\n  <!-- PRINT RECEIPT BLOCK -->\n  <app-receipt [order]="selectedOrder()"></app-receipt>\n`;
    fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/orders/orders/orders.html', htmlCode);
}

