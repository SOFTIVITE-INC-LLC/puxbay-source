const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'utf8');

code = code.replace(/openCheckout\(\) \{[\s\S]*?this\.isCheckoutModalOpen\.set\(true\);\n  \}/, `openCheckout() {
    if (this.cart().length === 0) return;
    this.isCheckoutModalOpen.set(true);
  }`);

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', code);
