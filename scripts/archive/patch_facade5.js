const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'utf8');

// Add state variable
code = code.replace("pointsRedeemed = signal(0);", "pointsRedeemed = signal(0);\n  isAutoPromoEnabled = signal(true);");

// Update computed property
code = code.replace(/readonly autoPromoDiscount = computed\(\(\) => \{[\s\S]*?\}\);/, `readonly autoPromoDiscount = computed(() => {
     if (!this.isAutoPromoEnabled()) return 0;
     const threshold = this.promoThreshold();
     const percent = this.promoPercent();
     return (this.cartSubtotal() > threshold && percent > 0) ? this.cartSubtotal() * (percent / 100) : 0;
  });`);

// Update resetCartState to re-enable it for the next sale
code = code.replace("this.pointsRedeemed.set(0);", "this.pointsRedeemed.set(0);\n    this.isAutoPromoEnabled.set(true);");

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', code);
