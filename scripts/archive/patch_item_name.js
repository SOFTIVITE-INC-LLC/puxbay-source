const fs = require('fs');

let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/shared/components/receipt/receipt.component.ts', 'utf8');

code = code.replace(/item\.product_name \|\| item\.name \|\| 'Item'/g, "item.product?.name || item.product_name || item.name || 'Item'");

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/shared/components/receipt/receipt.component.ts', code);
