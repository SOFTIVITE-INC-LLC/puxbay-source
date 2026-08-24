const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.ts', 'utf8');
code = code.replace(`
    // Esc to close checkout
    if (event.key === 'Escape') {
      if (this.facade.isCheckoutModalOpen()) {
        this.facade.closeCheckout();
      }
    }
`, `
    if (event.key === 'Escape') {
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
`);
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.ts', code);
