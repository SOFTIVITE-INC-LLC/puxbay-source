const fs = require('fs');
let htmlCode = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'utf8');

const regex = /<!-- PRINT RECEIPT BLOCK -->[\s\S]*?<\/div>\n  <\/div>\n/g;
htmlCode = htmlCode.replace(regex, `<!-- PRINT RECEIPT BLOCK -->
  <app-receipt [order]="facade.checkoutSuccessOrder()"></app-receipt>
`);
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', htmlCode);

let tsCode = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.ts', 'utf8');
if (!tsCode.includes('ReceiptComponent')) {
    tsCode = tsCode.replace("import { RouterModule } from '@angular/router';", "import { RouterModule } from '@angular/router';\nimport { ReceiptComponent } from '../../../shared/components/receipt/receipt.component';");
    tsCode = tsCode.replace("imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe],", "imports: [CommonModule, FormsModule, RouterModule, AppCurrencyPipe, ReceiptComponent],");
    fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.ts', tsCode);
}
