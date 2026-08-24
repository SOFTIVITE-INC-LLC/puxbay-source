const fs = require('fs');

// Update pos.ts styles
let tsCode = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.ts', 'utf8');
tsCode = tsCode.replace(/ \}\n  \`/, ` }\n    @media print {\n      #pos-app {\n        display: none !important;\n      }\n      #print-receipt {\n        display: block !important;\n        position: absolute;\n        top: 0;\n        left: 0;\n        width: 100%;\n      }\n    }\n  \``);
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.ts', tsCode);

// Update pos.html
let htmlCode = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'utf8');
htmlCode = htmlCode.replace('class="h-screen w-full flex bg-slate-50', 'id="pos-app" class="h-screen w-full flex bg-slate-50');

const receiptBlock = `
  <!-- PRINT RECEIPT BLOCK -->
  <div id="print-receipt" class="hidden text-black bg-white w-[80mm] p-4 text-sm font-mono mx-auto">
    <div class="text-center font-bold text-2xl mb-2">ThinkCE POS</div>
    <div class="text-center text-xs mb-4">
      Order: #{{ facade.checkoutSuccessOrder()?.order_number }}<br>
      Date: {{ (facade.checkoutSuccessOrder()?.created_at || current_date) | date:'short' }}
    </div>
    
    <div class="border-b-2 border-black mb-2 pb-2 border-dashed">
      <div *ngFor="let item of facade.checkoutSuccessOrder()?.items" class="flex justify-between mb-1">
        <span>{{ item.quantity }}x {{ item.product_name }}</span>
        <span>{{ (item.unit_price * item.quantity) | appCurrency }}</span>
      </div>
    </div>
    
    <div class="flex justify-between font-bold mb-1">
      <span>Subtotal</span>
      <span>{{ facade.checkoutSuccessOrder()?.subtotal | appCurrency }}</span>
    </div>
    <div *ngIf="facade.checkoutSuccessOrder()?.discount > 0" class="flex justify-between mb-1">
      <span>Discount</span>
      <span>-{{ facade.checkoutSuccessOrder()?.discount | appCurrency }}</span>
    </div>
    <div class="flex justify-between font-black text-xl mt-2 pt-2 border-t-2 border-black border-dashed">
      <span>Total</span>
      <span>{{ facade.checkoutSuccessOrder()?.total_amount | appCurrency }}</span>
    </div>
    
    <div class="text-center text-xs mt-8">
      Thank you for your business!
    </div>
  </div>
`;
htmlCode += receiptBlock;
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', htmlCode);
